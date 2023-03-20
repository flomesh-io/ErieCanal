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
	"github.com/flomesh-io/ErieCanal/controllers"
	clusterv1alpha1 "github.com/flomesh-io/ErieCanal/controllers/cluster/v1alpha1"
	gatewayv1beta1 "github.com/flomesh-io/ErieCanal/controllers/gateway/v1beta1"
	mcsv1alpha1 "github.com/flomesh-io/ErieCanal/controllers/mcs/v1alpha1"
	nsigv1alpha1 "github.com/flomesh-io/ErieCanal/controllers/namespacedingress/v1alpha1"
	svclb "github.com/flomesh-io/ErieCanal/controllers/servicelb"
	"github.com/flomesh-io/ErieCanal/pkg/version"
	"k8s.io/klog/v2"
)

func (c *ManagerConfig) RegisterReconcilers() error {
	mc := c.configStore.MeshConfig.GetConfig()
	rc := &controllers.ReconcilerConfig{
		Manager:            c.manager,
		ConfigStore:        c.configStore,
		K8sAPI:             c.k8sAPI,
		CertificateManager: c.certificateManager,
		RepoClient:         c.repoClient,
		Broker:             c.broker,
		Scheme:             c.manager.GetScheme(),
		Client:             c.manager.GetClient(),
	}
	reconcilers := make(map[string]controllers.Reconciler)

	reconcilers["MCS(Cluster)"] = clusterv1alpha1.NewReconciler(rc)
	reconcilers["MCS(ServiceExport)"] = mcsv1alpha1.NewServiceExportReconciler(rc)

	if mc.ShouldCreateServiceAndEndpointSlicesForMCS() && version.IsEndpointSliceEnabled(c.k8sAPI) {
		reconcilers["MCS(ServiceImport)"] = mcsv1alpha1.NewServiceImportReconciler(rc)
		reconcilers["MCS(Service)"] = mcsv1alpha1.NewServiceReconciler(rc)
		reconcilers["MCS(EndpointSlice)"] = mcsv1alpha1.NewEndpointSliceReconciler(rc)
	}

	if mc.IsGatewayApiEnabled() {
		reconcilers["GatewayAPI(GatewayClass)"] = gatewayv1beta1.NewGatewayClassReconciler(rc)
		reconcilers["GatewayAPI(Gateway)"] = gatewayv1beta1.NewGatewayReconciler(rc)
		reconcilers["GatewayAPI(HTTPRoute)"] = gatewayv1beta1.NewHTTPRouteReconciler(rc)
	}

	if mc.IsNamespacedIngressEnabled() {
		reconcilers["NamespacedIngress"] = nsigv1alpha1.NewReconciler(rc)
	}

	if mc.IsServiceLBEnabled() {
		reconcilers["ServiceLB(Service)"] = svclb.NewServiceReconciler(rc)
		reconcilers["ServiceLB(Node)"] = svclb.NewNodeReconciler(rc)
	}

	for name, r := range reconcilers {
		if err := r.SetupWithManager(c.manager); err != nil {
			klog.Errorf("Failed to setup reconciler %s: %s", name, err)
			return err
		}
	}

	return nil
}
