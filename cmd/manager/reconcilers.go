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
	clusterv1alpha1 "github.com/flomesh-io/ErieCanal/controllers/cluster/v1alpha1"
	gatewayv1beta1 "github.com/flomesh-io/ErieCanal/controllers/gateway/v1beta1"
	nsigv1alpha1 "github.com/flomesh-io/ErieCanal/controllers/namespacedingress/v1alpha1"
	svcexpv1alpha1 "github.com/flomesh-io/ErieCanal/controllers/serviceexport/v1alpha1"
	svcimpv1alpha1 "github.com/flomesh-io/ErieCanal/controllers/serviceimport/v1alpha1"
	svclb "github.com/flomesh-io/ErieCanal/controllers/servicelb"
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/event"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/util"
	"k8s.io/klog/v2"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func registerReconcilers(mgr manager.Manager, api *kube.K8sAPI, controlPlaneConfigStore *config.Store, certMgr certificate.Manager, broker *event.Broker) {
	registerCluster(mgr, api, controlPlaneConfigStore, broker, certMgr)
	registerServiceExport(mgr, api, controlPlaneConfigStore, broker)
	registerServiceImport(mgr, api, controlPlaneConfigStore)

	mc := controlPlaneConfigStore.MeshConfig.GetConfig()
	if mc.GatewayApi.Enabled {
		registerGatewayAPIs(mgr, api, controlPlaneConfigStore)
	}

	if mc.Ingress.Namespaced {
		registerNamespacedIngress(mgr, api, controlPlaneConfigStore, certMgr)
	}

	if mc.ServiceLB.Enabled {
		registerServiceLB(mgr, api, controlPlaneConfigStore)
	}
}

func registerCluster(mgr manager.Manager, api *kube.K8sAPI, controlPlaneConfigStore *config.Store, broker *event.Broker, certMgr certificate.Manager) {
	if err := (clusterv1alpha1.New(
		mgr.GetClient(),
		api,
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("Cluster"),
		controlPlaneConfigStore,
		broker,
		certMgr,
		util.RegisterOSExitHandlers(),
	)).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}
}

func registerServiceExport(mgr manager.Manager, api *kube.K8sAPI, controlPlaneConfigStore *config.Store, broker *event.Broker) {
	if err := (&svcexpv1alpha1.ServiceExportReconciler{
		Client:                  mgr.GetClient(),
		K8sAPI:                  api,
		Scheme:                  mgr.GetScheme(),
		Recorder:                mgr.GetEventRecorderFor("ServiceExport"),
		ControlPlaneConfigStore: controlPlaneConfigStore,
		Broker:                  broker,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, "unable to create controller", "controller", "ServiceExport")
		os.Exit(1)
	}
}

func registerServiceImport(mgr manager.Manager, api *kube.K8sAPI, controlPlaneConfigStore *config.Store) {
	if err := (&svcimpv1alpha1.ServiceImportReconciler{
		Client:                  mgr.GetClient(),
		K8sAPI:                  api,
		Scheme:                  mgr.GetScheme(),
		Recorder:                mgr.GetEventRecorderFor("ServiceImport"),
		ControlPlaneConfigStore: controlPlaneConfigStore,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, "unable to create controller", "controller", "ServiceImport")
		os.Exit(1)
	}
}

func registerNamespacedIngress(mgr manager.Manager, api *kube.K8sAPI, controlPlaneConfigStore *config.Store, certMgr certificate.Manager) {
	if err := (&nsigv1alpha1.NamespacedIngressReconciler{
		Client:                  mgr.GetClient(),
		K8sAPI:                  api,
		Scheme:                  mgr.GetScheme(),
		Recorder:                mgr.GetEventRecorderFor("NamespacedIngress"),
		ControlPlaneConfigStore: controlPlaneConfigStore,
		CertMgr:                 certMgr,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, "unable to create controller", "controller", "NamespacedIngress")
		os.Exit(1)
	}
}

func registerGatewayAPIs(mgr manager.Manager, api *kube.K8sAPI, controlPlaneConfigStore *config.Store) {
	if err := (&gatewayv1beta1.GatewayReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("Gateway"),
		K8sAPI:   api,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, "unable to create controller", "controller", "Gateway")
		os.Exit(1)
	}

	if err := (&gatewayv1beta1.GatewayClassReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("GatewayClass"),
		K8sAPI:   api,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, "unable to create controller", "controller", "GatewayClass")
		os.Exit(1)
	}

	if err := (&gatewayv1beta1.HTTPRouteReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("HTTPRoute"),
		K8sAPI:   api,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, "unable to create controller", "controller", "HTTPRoute")
		os.Exit(1)
	}
}

func registerServiceLB(mgr manager.Manager, api *kube.K8sAPI, store *config.Store) {
	if err := (&svclb.ServiceReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		Recorder:                mgr.GetEventRecorderFor("ServiceLB"),
		K8sAPI:                  api,
		ControlPlaneConfigStore: store,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, "unable to create controller", "controller", "ServiceLB(Service)")
		os.Exit(1)
	}
	if err := (&svclb.NodeReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		Recorder:                mgr.GetEventRecorderFor("ServiceLB"),
		K8sAPI:                  api,
		ControlPlaneConfigStore: store,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, "unable to create controller", "controller", "ServiceLB(Node)")
		os.Exit(1)
	}
}
