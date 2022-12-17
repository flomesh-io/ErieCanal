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

package gatewayclass

import (
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/util"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	gwv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	gwv1beta1validation "sigs.k8s.io/gateway-api/apis/v1beta1/validation"
)

const (
	kind      = "GatewayClass"
	groups    = "gateway.networking.k8s.io"
	resources = "gatewayclasses"
	versions  = "v1beta1"

	mwPath = commons.GatewayClassMutatingWebhookPath
	mwName = "mgatewayclass.kb.flomesh.io"
	vwPath = commons.GatewayClassValidatingWebhookPath
	vwName = "vgatewayclass.kb.flomesh.io"
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

type GatewayClassDefaulter struct {
	k8sAPI      *kube.K8sAPI
	configStore *config.Store
}

func NewDefaulter(k8sAPI *kube.K8sAPI, configStore *config.Store) *GatewayClassDefaulter {
	return &GatewayClassDefaulter{
		k8sAPI:      k8sAPI,
		configStore: configStore,
	}
}

func (w *GatewayClassDefaulter) RuntimeObject() runtime.Object {
	return &gwv1beta1.GatewayClass{}
}

func (w *GatewayClassDefaulter) SetDefaults(obj interface{}) {
	gatewayClass, ok := obj.(*gwv1beta1.GatewayClass)
	if !ok {
		return
	}

	klog.V(5).Infof("Default Webhook, name=%s", gatewayClass.Name)
	klog.V(4).Infof("Before setting default values, spec=%#v", gatewayClass.Spec)

	meshConfig := w.configStore.MeshConfig.GetConfig()

	if meshConfig == nil {
		return
	}

	klog.V(4).Infof("After setting default values, spec=%#v", gatewayClass.Spec)
}

type GatewayClassValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *GatewayClassValidator) RuntimeObject() runtime.Object {
	return &gwv1beta1.GatewayClass{}
}

func (w *GatewayClassValidator) ValidateCreate(obj interface{}) error {
	return doValidation(obj)
}

func (w *GatewayClassValidator) ValidateUpdate(oldObj, obj interface{}) error {
	oldGatewayClass, ok := oldObj.(*gwv1beta1.GatewayClass)
	if !ok {
		return nil
	}

	gatewayClass, ok := obj.(*gwv1beta1.GatewayClass)
	if !ok {
		return nil
	}

	errorList := gwv1beta1validation.ValidateGatewayClassUpdate(oldGatewayClass, gatewayClass)
	if len(errorList) > 0 {
		return util.ErrorListToError(errorList)
	}

	return nil
}

func (w *GatewayClassValidator) ValidateDelete(obj interface{}) error {
	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *GatewayClassValidator {
	return &GatewayClassValidator{
		k8sAPI: k8sAPI,
	}
}

func doValidation(obj interface{}) error {
	//gatewayClass, ok := obj.(*gwv1beta1.GatewayClass)
	//if !ok {
	//    return nil
	//}

	return nil
}
