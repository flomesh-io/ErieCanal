/*
 * Copyright 2022 The flomesh.io Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	mcsevent "github.com/flomesh-io/ErieCanal/pkg/mcs/event"
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
	// create a new manager for controllers
	mgr := newManager(kubeconfig, options)
	stopCh := util.RegisterOSExitHandlers()
	broker := mcsevent.NewBroker(stopCh)

	managerCfg := &ManagerConfig{
		manager:            mgr,
		k8sAPI:             k8sApi,
		configStore:        controlPlaneConfigStore,
		certificateManager: certMgr,
		repoClient:         repoClient,
		broker:             broker,
		stopCh:             stopCh,
	}

	if mc.IsIngressEnabled() {
		managerCfg.connector = managerCfg.GetLocalConnector()
	}

	if mc.IsGatewayApiEnabled() {
		if !version.IsSupportedK8sVersionForGatewayAPI(k8sApi) {
			klog.Errorf("kubernetes server version %s is not supported, requires at least %s",
				version.ServerVersion.String(), version.MinK8sVersionForGatewayAPI.String())
			os.Exit(1)
		}

		managerCfg.eventHandler = managerCfg.GetResourceEventHandler()
	}

	for _, f := range []func() error{
		managerCfg.InitRepo,
		managerCfg.SetupHTTP,
		managerCfg.SetupTLS,
		managerCfg.RegisterWebHooks,
		managerCfg.RegisterEventHandlers,
		managerCfg.RegisterReconcilers,
		managerCfg.AddLivenessAndReadinessCheck,
		managerCfg.StartManager,
	} {
		if err := f(); err != nil {
			klog.Errorf("Failed to startup: %s", err)
			os.Exit(1)
		}
	}
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

func (c *ManagerConfig) StartManager() error {
	if c.connector != nil {
		if err := c.manager.Add(manager.RunnableFunc(func(ctx context.Context) error {
			return c.connector.Run(c.stopCh)
		})); err != nil {
			return err
		}
	}

	if c.eventHandler != nil {
		if err := c.manager.Add(c.eventHandler); err != nil {
			return err
		}
	}

	klog.Info("starting manager")
	if err := c.manager.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Fatalf("problem running manager, %s", err)
		return err
	}

	return nil
}
