/*
 * MIT License
 *
 * Copyright (c) since 2021,  flomesh.io Authors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/event"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"github.com/flomesh-io/ErieCanal/pkg/util"
	"github.com/flomesh-io/ErieCanal/pkg/util/tls"
	"github.com/flomesh-io/ErieCanal/pkg/version"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"math/rand"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	flomeshscheme "github.com/flomesh-io/ErieCanal/pkg/generated/clientset/versioned/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	gwschema "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/scheme"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
	//setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(flomeshscheme.AddToScheme(scheme))
	utilruntime.Must(gwschema.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

type startArgs struct {
	managerConfigFile string
	namespace         string
}

func main() {
	// process CLI arguments and parse them to flags
	args := processFlags()
	options := loadManagerOptions(args.managerConfigFile)

	klog.Infof(commons.AppVersionTemplate, version.Version, version.ImageVersion, version.GitVersion, version.GitCommit, version.BuildDate)

	kubeconfig := ctrl.GetConfigOrDie()
	k8sApi := newK8sAPI(kubeconfig, args)
	if !version.IsSupportedK8sVersion(k8sApi) {
		klog.Error(fmt.Errorf("kubernetes server version %s is not supported, requires at least %s",
			version.ServerVersion.String(), version.MinK8sVersion.String()))
		os.Exit(1)
	}

	// ErieCanal configurations
	controlPlaneConfigStore := config.NewStore(k8sApi)
	mcClient := controlPlaneConfigStore.MeshConfig
	mc := mcClient.GetConfig()
	mc.Cluster.UID = getClusterUID(k8sApi)
	mc, err := mcClient.UpdateConfig(mc)
	if err != nil {
		os.Exit(1)
	}

	// generate certificate and store it in k8s secret flomesh-ca-bundle
	certMgr, err := tls.GetCertificateManager(k8sApi, mc)
	if err != nil {
		os.Exit(1)
	}

	// upload init scripts to pipy repo
	repoClient := repo.NewRepoClient(mc.RepoRootURL())
	initRepo(repoClient)

	// setup TLS config
	setupTLS(certMgr, repoClient, mc)

	// create a new manager for controllers
	mgr := newManager(kubeconfig, options)

	stopCh := util.RegisterOSExitHandlers()
	broker := event.NewBroker(stopCh)

	// create mutating and validating webhook configurations
	createWebhookConfigurations(k8sApi, controlPlaneConfigStore, certMgr)

	// register Reconcilers
	registerReconcilers(mgr, k8sApi, controlPlaneConfigStore, certMgr, broker)

	// register webhooks
	registerToWebhookServer(mgr, k8sApi, controlPlaneConfigStore)

	registerEventHandler(mgr, k8sApi, controlPlaneConfigStore)

	// add endpoints for Liveness and Readiness check
	addLivenessAndReadinessCheck(mgr)
	//+kubebuilder:scaffold:builder

	// start the controller manager
	startManager(mgr, mc)
}

func processFlags() *startArgs {
	var configFile string
	flag.StringVar(&configFile, "config", "manager_config.yaml",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")

	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	rand.Seed(time.Now().UnixNano())
	ctrl.SetLogger(klogr.New())

	return &startArgs{
		managerConfigFile: configFile,
		namespace:         config.GetErieCanalNamespace(),
	}
}

func loadManagerOptions(configFile string) ctrl.Options {
	var err error
	options := ctrl.Options{Scheme: scheme}
	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile))
		if err != nil {
			klog.Error(err, "unable to load the config file")
			os.Exit(1)
		}
	}

	return options
}

func newManager(kubeconfig *rest.Config, options ctrl.Options) manager.Manager {
	mgr, err := ctrl.NewManager(kubeconfig, options)
	if err != nil {
		klog.Error(err, "unable to create manager")
		os.Exit(1)
	}

	return mgr
}

func newK8sAPI(kubeconfig *rest.Config, args *startArgs) *kube.K8sAPI {
	api, err := kube.NewAPIForConfig(kubeconfig, 30*time.Second)
	if err != nil {
		klog.Error(err, "unable to create k8s client")
		os.Exit(1)
	}

	return api
}

func getClusterUID(api *kube.K8sAPI) string {
	ns, err := api.Client.CoreV1().Namespaces().Get(context.TODO(), config.GetErieCanalNamespace(), metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get ErieCanal namespace: %s", err)
		os.Exit(1)
	}

	return string(ns.UID)
}

func addLivenessAndReadinessCheck(mgr manager.Manager) {
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
}

func startManager(mgr manager.Manager, mc *config.MeshConfig) {
	//err := mgr.Add(manager.RunnableFunc(func(context.Context) error {
	//	aggregatorAddr := fmt.Sprintf(":%s", mc.AggregatorPort())
	//	return aggregator.NewAggregator(aggregatorAddr, mc.RepoAddr()).Run()
	//}))
	//if err != nil {
	//	klog.Error(err, "unable add aggregator server to the manager")
	//	os.Exit(1)
	//}

	klog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Fatalf("problem running manager, %s", err.Error())
		os.Exit(1)
	}
}
