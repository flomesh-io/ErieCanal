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

package serviceexport

import (
	svcexpv1alpha1 "github.com/flomesh-io/ErieCanal/apis/serviceexport/v1alpha1"
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	kind      = "ServiceExport"
	groups    = "flomesh.io"
	resources = "serviceexports"
	versions  = "v1alpha1"

	mwPath = commons.ServiceExportMutatingWebhookPath
	mwName = "mserviceexport.kb.flomesh.io"
	vwPath = commons.ServiceExportValidatingWebhookPath
	vwName = "vserviceexport.kb.flomesh.io"
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

type ServiceExportDefaulter struct {
	k8sAPI      *kube.K8sAPI
	configStore *config.Store
}

func NewDefaulter(k8sAPI *kube.K8sAPI, configStore *config.Store) *ServiceExportDefaulter {
	return &ServiceExportDefaulter{
		k8sAPI:      k8sAPI,
		configStore: configStore,
	}
}

func (w *ServiceExportDefaulter) RuntimeObject() runtime.Object {
	return &svcexpv1alpha1.ServiceExport{}
}

func (w *ServiceExportDefaulter) SetDefaults(obj interface{}) {
	//serviceExport, ok := obj.(*svcexpv1alpha1.ServiceExport)
	//if !ok {
	//	return
	//}
	//
	//klog.V(5).Infof("Default Webhook, name=%s", serviceExport.Name)
	//klog.V(4).Infof("Before setting default values, spec=%#v", serviceExport.Spec)
	//
	//meshConfig := w.configStore.MeshConfig.GetConfig()
	//
	//if meshConfig == nil {
	//	return
	//}
	//
	//klog.V(4).Infof("After setting default values, spec=%#v", serviceExport.Spec)
}

type ServiceExportValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *ServiceExportValidator) RuntimeObject() runtime.Object {
	return &svcexpv1alpha1.ServiceExport{}
}

func (w *ServiceExportValidator) ValidateCreate(obj interface{}) error {
	return doValidation(obj)
}

func (w *ServiceExportValidator) ValidateUpdate(oldObj, obj interface{}) error {
	return doValidation(obj)
}

func (w *ServiceExportValidator) ValidateDelete(obj interface{}) error {
	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *ServiceExportValidator {
	return &ServiceExportValidator{
		k8sAPI: k8sAPI,
	}
}

func doValidation(obj interface{}) error {
	//serviceExport, ok := obj.(*svcexpv1alpha1.ServiceExport)
	//if !ok {
	//    return nil
	//}

	return nil
}
