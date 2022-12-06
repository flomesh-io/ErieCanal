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

package serviceimport

import (
	svcimpv1alpha1 "github.com/flomesh-io/ErieCanal/apis/serviceimport/v1alpha1"
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

const (
	kind      = "ServiceImport"
	groups    = "flomesh.io"
	resources = "serviceimports"
	versions  = "v1alpha1"

	mwPath = commons.ServiceImportMutatingWebhookPath
	mwName = "mserviceimport.kb.flomesh.io"
	vwPath = commons.ServiceImportValidatingWebhookPath
	vwName = "vserviceimport.kb.flomesh.io"
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

type ServiceImportDefaulter struct {
	k8sAPI      *kube.K8sAPI
	configStore *config.Store
}

func NewDefaulter(k8sAPI *kube.K8sAPI, configStore *config.Store) *ServiceImportDefaulter {
	return &ServiceImportDefaulter{
		k8sAPI:      k8sAPI,
		configStore: configStore,
	}
}

func (w *ServiceImportDefaulter) RuntimeObject() runtime.Object {
	return &svcimpv1alpha1.ServiceImport{}
}

func (w *ServiceImportDefaulter) SetDefaults(obj interface{}) {
	serviceImport, ok := obj.(*svcimpv1alpha1.ServiceImport)
	if !ok {
		return
	}

	klog.V(5).Infof("Default Webhook, name=%s", serviceImport.Name)
	klog.V(4).Infof("Before setting default values, spec=%#v", serviceImport.Spec)

	meshConfig := w.configStore.MeshConfig.GetConfig()

	if meshConfig == nil {
		return
	}

	if serviceImport.Spec.Type == "" {
		// ONLY set the value, there's no any logic to handle the type yet
		serviceImport.Spec.Type = svcimpv1alpha1.ClusterSetIP
	}

	klog.V(4).Infof("After setting default values, spec=%#v", serviceImport.Spec)
}

type ServiceImportValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *ServiceImportValidator) RuntimeObject() runtime.Object {
	return &svcimpv1alpha1.ServiceImport{}
}

func (w *ServiceImportValidator) ValidateCreate(obj interface{}) error {
	return doValidation(obj)
}

func (w *ServiceImportValidator) ValidateUpdate(oldObj, obj interface{}) error {
	return doValidation(obj)
}

func (w *ServiceImportValidator) ValidateDelete(obj interface{}) error {
	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *ServiceImportValidator {
	return &ServiceImportValidator{
		k8sAPI: k8sAPI,
	}
}

func doValidation(obj interface{}) error {
	//serviceImport, ok := obj.(*svcimpv1alpha1.ServiceImport)
	//if !ok {
	//    return nil
	//}

	return nil
}
