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

package tlsroute

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
	kind      = "TLSRoute"
	groups    = "gateway.networking.k8s.io"
	resources = "tlsroutes"
	versions  = "v1alpha2"

	mwPath = commons.TLSRouteMutatingWebhookPath
	mwName = "mtlsroute.kb.flomesh.io"
	vwPath = commons.TLSRouteValidatingWebhookPath
	vwName = "vtlsroute.kb.flomesh.io"
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

type TLSRouteDefaulter struct {
	k8sAPI      *kube.K8sAPI
	configStore *config.Store
}

func NewDefaulter(k8sAPI *kube.K8sAPI, configStore *config.Store) *TLSRouteDefaulter {
	return &TLSRouteDefaulter{
		k8sAPI:      k8sAPI,
		configStore: configStore,
	}
}

func (w *TLSRouteDefaulter) RuntimeObject() runtime.Object {
	return &gwv1alpha2.TLSRoute{}
}

func (w *TLSRouteDefaulter) SetDefaults(obj interface{}) {
	route, ok := obj.(*gwv1alpha2.TLSRoute)
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

type TLSRouteValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *TLSRouteValidator) RuntimeObject() runtime.Object {
	return &gwv1alpha2.TLSRoute{}
}

func (w *TLSRouteValidator) ValidateCreate(obj interface{}) error {
	return doValidation(obj)
}

func (w *TLSRouteValidator) ValidateUpdate(oldObj, obj interface{}) error {
	return doValidation(obj)
}

func (w *TLSRouteValidator) ValidateDelete(obj interface{}) error {
	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *TLSRouteValidator {
	return &TLSRouteValidator{
		k8sAPI: k8sAPI,
	}
}

func doValidation(obj interface{}) error {
	//route, ok := obj.(*gwv1alpha2.TLSRoute)
	//if !ok {
	//    return nil
	//}

	return nil
}
