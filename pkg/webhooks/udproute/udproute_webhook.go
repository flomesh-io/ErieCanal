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

package udproute

import (
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	kind      = "UDPRoute"
	groups    = "gateway.networking.k8s.io"
	resources = "udproutes"
	versions  = "v1alpha2"

	mwPath = commons.UDPRouteMutatingWebhookPath
	mwName = "mudproute.kb.flomesh.io"
	vwPath = commons.UDPRouteValidatingWebhookPath
	vwName = "vudproute.kb.flomesh.io"
)

func RegisterWebhooks(webhookSvcNs, webhookSvcName string, caBundle []byte) {
	rule := flomeshadmission.NewRule(
		[]admissionregv1.OperationType{admissionregv1.Create, admissionregv1.Update},
		[]string{groups},
		[]string{versions},
		[]string{resources},
	)

	mutatingWebhook := flomeshadmission.NewMutatingWebhook(
		mwName,
		webhookSvcNs,
		webhookSvcName,
		mwPath,
		caBundle,
		nil,
		[]admissionregv1.RuleWithOperations{rule},
	)

	validatingWebhook := flomeshadmission.NewValidatingWebhook(
		vwName,
		webhookSvcNs,
		webhookSvcName,
		vwPath,
		caBundle,
		nil,
		[]admissionregv1.RuleWithOperations{rule},
	)

	flomeshadmission.RegisterMutatingWebhook(mwName, mutatingWebhook)
	flomeshadmission.RegisterValidatingWebhook(vwName, validatingWebhook)
}

type UDPRouteDefaulter struct {
	k8sAPI      *kube.K8sAPI
	configStore *config.Store
}

func NewDefaulter(k8sAPI *kube.K8sAPI, configStore *config.Store) *UDPRouteDefaulter {
	return &UDPRouteDefaulter{
		k8sAPI:      k8sAPI,
		configStore: configStore,
	}
}

func (w *UDPRouteDefaulter) RuntimeObject() runtime.Object {
	return &gwv1alpha2.UDPRoute{}
}

func (w *UDPRouteDefaulter) SetDefaults(obj interface{}) {
	route, ok := obj.(*gwv1alpha2.UDPRoute)
	if !ok {
		return
	}

	klog.V(5).Infof("Default Webhook, name=%s", route.Name)
	klog.V(4).Infof("Before setting default values, spec=%#v", route.Spec)

	meshConfig := w.configStore.MeshConfig.GetConfig()

	if meshConfig == nil {
		return
	}

	klog.V(4).Infof("After setting default values, spec=%#v", route.Spec)
}

type UDPRouteValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *UDPRouteValidator) RuntimeObject() runtime.Object {
	return &gwv1alpha2.UDPRoute{}
}

func (w *UDPRouteValidator) ValidateCreate(obj interface{}) error {
	return doValidation(obj)
}

func (w *UDPRouteValidator) ValidateUpdate(oldObj, obj interface{}) error {
	return doValidation(obj)
}

func (w *UDPRouteValidator) ValidateDelete(obj interface{}) error {
	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *UDPRouteValidator {
	return &UDPRouteValidator{
		k8sAPI: k8sAPI,
	}
}

func doValidation(obj interface{}) error {
	//route, ok := obj.(*gwv1alpha2.UDPRoute)
	//if !ok {
	//    return nil
	//}

	return nil
}
