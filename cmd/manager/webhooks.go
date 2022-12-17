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
	"fmt"
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks"
	clusterwh "github.com/flomesh-io/ErieCanal/pkg/webhooks/cluster"
	cmwh "github.com/flomesh-io/ErieCanal/pkg/webhooks/cm"
	gatewaywh "github.com/flomesh-io/ErieCanal/pkg/webhooks/gateway"
	gatewayclasswh "github.com/flomesh-io/ErieCanal/pkg/webhooks/gatewayclass"
	gtpwh "github.com/flomesh-io/ErieCanal/pkg/webhooks/globaltrafficpolicy"
	httproutewh "github.com/flomesh-io/ErieCanal/pkg/webhooks/httproute"
	idwh "github.com/flomesh-io/ErieCanal/pkg/webhooks/namespacedingress"
	referencepolicywh "github.com/flomesh-io/ErieCanal/pkg/webhooks/referencepolicy"
	svcexpwh "github.com/flomesh-io/ErieCanal/pkg/webhooks/serviceexport"
	svcimpwh "github.com/flomesh-io/ErieCanal/pkg/webhooks/serviceimport"
	tcproutewh "github.com/flomesh-io/ErieCanal/pkg/webhooks/tcproute"
	tlsroutewh "github.com/flomesh-io/ErieCanal/pkg/webhooks/tlsroute"
	udproutewh "github.com/flomesh-io/ErieCanal/pkg/webhooks/udproute"
	"io/ioutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func createWebhookConfigurations(k8sApi *kube.K8sAPI, configStore *config.Store, certMgr certificate.Manager) {
	mc := configStore.MeshConfig.GetConfig()
	cert, err := issueCertForWebhook(certMgr, mc)
	if err != nil {
		os.Exit(1)
	}

	ns := config.GetErieCanalNamespace()
	svcName := mc.Webhook.ServiceName
	caBundle := cert.CA
	webhooks.RegisterWebhooks(ns, svcName, caBundle)
	if mc.GatewayApi.Enabled {
		webhooks.RegisterGatewayApiWebhooks(ns, svcName, caBundle)
	}

	// Mutating
	mwc := flomeshadmission.NewMutatingWebhookConfiguration()
	mutating := k8sApi.Client.
		AdmissionregistrationV1().
		MutatingWebhookConfigurations()
	if _, err = mutating.Create(context.Background(), mwc, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			existingMwc, err := mutating.Get(context.Background(), mwc.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Unable to get MutatingWebhookConfigurations %q, %s", mwc.Name, err.Error())
				os.Exit(1)
			}

			existingMwc.Webhooks = mwc.Webhooks
			_, err = mutating.Update(context.Background(), existingMwc, metav1.UpdateOptions{})
			if err != nil {
				// Should be not conflict for a leader-election manager, error is error
				klog.Errorf("Unable to update MutatingWebhookConfigurations %q, %s", mwc.Name, err.Error())
				os.Exit(1)
			}
		} else {
			klog.Errorf("Unable to create MutatingWebhookConfigurations %q, %s", mwc.Name, err.Error())
			os.Exit(1)
		}
	}

	// Validating
	vmc := flomeshadmission.NewValidatingWebhookConfiguration()
	validating := k8sApi.Client.
		AdmissionregistrationV1().
		ValidatingWebhookConfigurations()
	if _, err = validating.Create(context.Background(), vmc, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			existingVmc, err := validating.Get(context.Background(), vmc.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("Unable to get ValidatingWebhookConfigurations %q, %s", mwc.Name, err.Error())
				os.Exit(1)
			}

			existingVmc.Webhooks = vmc.Webhooks
			_, err = validating.Update(context.Background(), existingVmc, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("Unable to update ValidatingWebhookConfigurations %q, %s", vmc.Name, err.Error())
				os.Exit(1)
			}
		} else {
			klog.Errorf("Unable to create ValidatingWebhookConfigurations %q, %s", vmc.Name, err.Error())
			os.Exit(1)
		}
	}
}

func issueCertForWebhook(certMgr certificate.Manager, mc *config.MeshConfig) (*certificate.Certificate, error) {
	// TODO: refactoring it later, configurable CN and dns names
	cert, err := certMgr.IssueCertificate(
		mc.Webhook.ServiceName,
		commons.DefaultCAValidityPeriod,
		[]string{
			mc.Webhook.ServiceName,
			fmt.Sprintf("%s.%s.svc", mc.Webhook.ServiceName, config.GetErieCanalNamespace()),
			fmt.Sprintf("%s.%s.svc.cluster.local", mc.Webhook.ServiceName, config.GetErieCanalNamespace()),
		},
	)
	if err != nil {
		klog.Error("Error issuing certificate, ", err)
		return nil, err
	}

	// write ca.crt, tls.crt & tls.key to file
	if err := os.MkdirAll(commons.WebhookServerServingCertsPath, 755); err != nil {
		klog.Errorf("error creating dir %q, %s", commons.WebhookServerServingCertsPath, err.Error())
		return nil, err
	}

	certFiles := map[string][]byte{
		commons.RootCACertName:    cert.CA,
		commons.TLSCertName:       cert.CrtPEM,
		commons.TLSPrivateKeyName: cert.KeyPEM,
	}

	for file, data := range certFiles {
		fileName := fmt.Sprintf("%s/%s", commons.WebhookServerServingCertsPath, file)
		if err := ioutil.WriteFile(
			fileName,
			data,
			420); err != nil {
			klog.Errorf("error writing file %q, %s", fileName, err.Error())
			return nil, err
		}
	}

	return cert, nil
}

func registerToWebhookServer(mgr manager.Manager, api *kube.K8sAPI, controlPlaneConfigStore *config.Store) {
	hookServer := mgr.GetWebhookServer()
	mc := controlPlaneConfigStore.MeshConfig.GetConfig()

	// Cluster
	hookServer.Register(commons.ClusterMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(clusterwh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.ClusterValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(clusterwh.NewValidator(api)),
	)

	// NamespacedIngress
	hookServer.Register(commons.NamespacedIngressMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(idwh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.NamespacedIngressValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(idwh.NewValidator(api)),
	)

	// ServiceExport
	hookServer.Register(commons.ServiceExportMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(svcexpwh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.ServiceExportValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(svcexpwh.NewValidator(api)),
	)

	// ServiceImport
	hookServer.Register(commons.ServiceImportMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(svcimpwh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.ServiceImportValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(svcimpwh.NewValidator(api)),
	)

	// GlobalTrafficPolicy
	hookServer.Register(commons.GlobalTrafficPolicyMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(gtpwh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.GlobalTrafficPolicyValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(gtpwh.NewValidator(api)),
	)

	// core ConfigMap
	hookServer.Register(commons.ConfigMapMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(cmwh.NewDefaulter(api)),
	)
	hookServer.Register(commons.ConfigMapValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(cmwh.NewValidator(api)),
	)

	// Gateway API
	if mc.GatewayApi.Enabled {
		registerGatewayApiToWebhookServer(mgr, api, controlPlaneConfigStore)
	}
}

func registerGatewayApiToWebhookServer(mgr manager.Manager, api *kube.K8sAPI, controlPlaneConfigStore *config.Store) {
	hookServer := mgr.GetWebhookServer()

	// Gateway
	hookServer.Register(commons.GatewayMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(gatewaywh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.GatewayValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(gatewaywh.NewValidator(api)),
	)

	// GatewayClass
	hookServer.Register(commons.GatewayClassMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(gatewayclasswh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.GatewayClassValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(gatewayclasswh.NewValidator(api)),
	)

	// ReferencePolicy
	hookServer.Register(commons.ReferencePolicyMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(referencepolicywh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.ReferencePolicyValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(referencepolicywh.NewValidator(api)),
	)

	// HTTPRoute
	hookServer.Register(commons.HTTPRouteMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(httproutewh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.HTTPRouteValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(httproutewh.NewValidator(api)),
	)

	// TCPRoute
	hookServer.Register(commons.TCPRouteMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(tcproutewh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.TCPRouteValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(tcproutewh.NewValidator(api)),
	)

	// TLSRoute
	hookServer.Register(commons.TLSRouteMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(tlsroutewh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.TLSRouteValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(tlsroutewh.NewValidator(api)),
	)

	// UDPRoute
	hookServer.Register(commons.UDPRouteMutatingWebhookPath,
		webhooks.DefaultingWebhookFor(udproutewh.NewDefaulter(api, controlPlaneConfigStore)),
	)
	hookServer.Register(commons.UDPRouteValidatingWebhookPath,
		webhooks.ValidatingWebhookFor(udproutewh.NewValidator(api)),
	)
}
