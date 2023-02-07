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
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

func registerEventHandler(mgr manager.Manager, api *kube.K8sAPI, controlPlaneConfigStore *config.Store, certMgr certificate.Manager) {

	// FIXME: make it configurable
	resyncPeriod := 15 * time.Minute

	configmapInformer, err := mgr.GetCache().GetInformer(context.TODO(), &corev1.ConfigMap{})

	if err != nil {
		klog.Error(err, "unable to get informer for ConfigMap")
		os.Exit(1)
	}

	config.RegisterConfigurationHanlder(
		config.NewFlomeshConfigurationHandler(
			mgr.GetClient(),
			api,
			controlPlaneConfigStore,
			certMgr,
		),
		configmapInformer,
		resyncPeriod,
	)
}
