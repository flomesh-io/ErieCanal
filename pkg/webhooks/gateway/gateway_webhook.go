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

package gateway

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
	kind      = "Gateway"
	groups    = "gateway.networking.k8s.io"
	resources = "gateways"
	versions  = "v1beta1"

	mwPath = commons.GatewayMutatingWebhookPath
	mwName = "mgateway.kb.flomesh.io"
	vwPath = commons.GatewayValidatingWebhookPath
	vwName = "vgateway.kb.flomesh.io"
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

type GatewayDefaulter struct {
	k8sAPI      *kube.K8sAPI
	configStore *config.Store
}

func NewDefaulter(k8sAPI *kube.K8sAPI, configStore *config.Store) *GatewayDefaulter {
	return &GatewayDefaulter{
		k8sAPI:      k8sAPI,
		configStore: configStore,
	}
}

func (w *GatewayDefaulter) RuntimeObject() runtime.Object {
	return &gwv1beta1.Gateway{}
}

func (w *GatewayDefaulter) SetDefaults(obj interface{}) {
	gateway, ok := obj.(*gwv1beta1.Gateway)
	if !ok {
		return
	}

	klog.V(5).Infof("Default Webhook, name=%s", gateway.Name)
	klog.V(4).Infof("Before setting default values, spec=%#v", gateway.Spec)

	meshConfig := w.configStore.MeshConfig.GetConfig()

	if meshConfig == nil {
		return
	}

	klog.V(4).Infof("After setting default values, spec=%#v", gateway.Spec)
}

type GatewayValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *GatewayValidator) RuntimeObject() runtime.Object {
	return &gwv1beta1.Gateway{}
}

func (w *GatewayValidator) ValidateCreate(obj interface{}) error {
	return doValidation(obj)
}

func (w *GatewayValidator) ValidateUpdate(oldObj, obj interface{}) error {
	return doValidation(obj)
}

func (w *GatewayValidator) ValidateDelete(obj interface{}) error {
	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *GatewayValidator {
	return &GatewayValidator{
		k8sAPI: k8sAPI,
	}
}

func doValidation(obj interface{}) error {
	gateway, ok := obj.(*gwv1beta1.Gateway)
	if !ok {
		return nil
	}

	errorList := gwv1beta1validation.ValidateGateway(gateway)
	if len(errorList) > 0 {
		return util.ErrorListToError(errorList)
	}

	return nil
}
