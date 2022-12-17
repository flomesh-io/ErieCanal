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
