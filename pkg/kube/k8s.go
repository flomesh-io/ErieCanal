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

package kube

import (
	"fmt"
	flomesh "github.com/flomesh-io/ErieCanal/pkg/generated/clientset/versioned"
	extensionsClientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	cfg "sigs.k8s.io/controller-runtime/pkg/client/config"
	"time"
)

type K8sAPI struct {
	*rest.Config
	Client           kubernetes.Interface
	EventClient      v1core.EventsGetter
	DynamicClient    dynamic.Interface
	DiscoveryClient  discovery.DiscoveryInterface
	FlomeshClient    flomesh.Interface
	ExtensionsClient extensionsClientset.Interface
}

/**
 * Config precedence
 *      --kubeconfig flag pointing at a file
 *      KUBECONFIG environment variable pointing at a file
 *      In-cluster config if running in cluster
 *      $HOME/.kube/config if exists.
 */

func NewAPI(timeout time.Duration) (*K8sAPI, error) {
	config, err := cfg.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("error get config for K8s API client: %v", err)
	}
	return NewAPIForConfig(config, timeout)
}

func NewAPIForContext(kubeContext string, timeout time.Duration) (*K8sAPI, error) {
	config, err := cfg.GetConfigWithContext(kubeContext)
	if err != nil {
		return nil, fmt.Errorf("error get config for K8s API client: %v", err)
	}
	return NewAPIForConfig(config, timeout)
}

func NewAPIForConfig(config *rest.Config, timeout time.Duration) (*K8sAPI, error) {
	return NewAPIForConfigOrDie(config, timeout)
}

func NewAPIForConfigOrDie(config *rest.Config, timeout time.Duration) (*K8sAPI, error) {
	config.Timeout = timeout

	clientset := kubernetes.NewForConfigOrDie(config)
	eventClient := kubernetes.NewForConfigOrDie(config)
	dynamicClient := dynamic.NewForConfigOrDie(config)
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(config)
	flomeshClient := flomesh.NewForConfigOrDie(config)
	extensionsClient := extensionsClientset.NewForConfigOrDie(config)

	return &K8sAPI{
		Config:           config,
		Client:           clientset,
		EventClient:      eventClient.CoreV1(),
		DynamicClient:    dynamicClient,
		DiscoveryClient:  discoveryClient,
		FlomeshClient:    flomeshClient,
		ExtensionsClient: extensionsClient,
	}, nil
}

func MetaNamespaceKey(obj interface{}) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Warning(err)
	}

	return key
}
