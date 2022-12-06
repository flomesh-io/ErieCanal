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
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	flomeshscheme "github.com/flomesh-io/ErieCanal/pkg/generated/clientset/versioned/scheme"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"github.com/flomesh-io/ErieCanal/pkg/version"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"math/rand"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(flomeshscheme.AddToScheme(scheme))
}

func main() {
	processFlags()
	//proxyInitConfig := getProxyInitConfig()

	//lockId := uuid.New().String()
	//lockName := fmt.Sprintf("%s.pf.flomesh.io", proxyInitConfig.MatchedProxyProfile)
	//lockNamespace := "kube-system"

	klog.Infof(commons.AppVersionTemplate, version.Version, version.ImageVersion, version.GitVersion, version.GitCommit, version.BuildDate)

	//kubeconfig := ctrl.GetConfigOrDie()
	//k8sApi := newK8sAPI(kubeconfig)
	//if !version.IsSupportedK8sVersion(k8sApi) {
	//	klog.Error(fmt.Errorf("kubernetes server version %s is not supported, requires at least %s",
	//		version.ServerVersion.String(), version.MinK8sVersion.String()))
	//	os.Exit(1)
	//}

	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//
	//lock := getNewLock(k8sApi, lockId, lockName, lockNamespace)
	//runLeaderElection(lock, ctx, lockId, proxyInitConfig)
}

func processFlags() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	rand.Seed(time.Now().UnixNano())
	ctrl.SetLogger(klogr.New())
}

func getProxyInitConfig() config.ProxyInitEnvironmentConfiguration {
	var cfg config.ProxyInitEnvironmentConfiguration

	err := envconfig.Process("FLOMESH", &cfg)
	if err != nil {
		klog.Error(err, "unable to load the configuration from environment")
		os.Exit(1)
	}

	return cfg
}

func newK8sAPI(kubeconfig *rest.Config) *kube.K8sAPI {
	api, err := kube.NewAPIForConfig(kubeconfig, 30*time.Second)
	if err != nil {
		klog.Error(err, "unable to create k8s client")
		os.Exit(1)
	}

	return api
}

func getNewLock(api *kube.K8sAPI, lockId, lockName, lockNamespace string) *resourcelock.LeaseLock {
	return &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      lockName,
			Namespace: lockNamespace,
		},
		Client: api.Client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: lockId,
		},
	}
}

func runLeaderElection(lock *resourcelock.LeaseLock, ctx context.Context, lockId string, cfg config.ProxyInitEnvironmentConfiguration) {
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {
				deriveCodebases(cfg)
			},
			OnStoppedLeading: func() {
				klog.Info("no longer the leader, staying inactive.")
			},
			OnNewLeader: func(currentId string) {
				if currentId == lockId {
					klog.Info("still the leader!")
					return
				}
				klog.Info("new leader is %s", currentId)
			},
		},
	})
}

func deriveCodebases(cfg config.ProxyInitEnvironmentConfiguration) {
	repoClient := repo.NewRepoClient(cfg.ProxyRepoRootUrl)
	parentPath := cfg.ProxyParentPath

	for _, sidecarPath := range cfg.ProxyPaths {
		if err := repoClient.DeriveCodebase(sidecarPath, parentPath); err != nil {
			os.Exit(1)
		}
	}
}
