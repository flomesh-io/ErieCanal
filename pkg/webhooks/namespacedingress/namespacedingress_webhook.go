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

package namespacedingress

import (
	"context"
	nsigv1alpha1 "github.com/flomesh-io/ErieCanal/apis/namespacedingress/v1alpha1"
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/pkg/errors"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

const (
	kind      = "NamespacedIngress"
	groups    = "flomesh.io"
	resources = "namespacedingresses"
	versions  = "v1alpha1"

	mwPath = commons.NamespacedIngressMutatingWebhookPath
	mwName = "mnamespacedingress.kb.flomesh.io"
	vwPath = commons.NamespacedIngressValidatingWebhookPath
	vwName = "vnamespacedingress.kb.flomesh.io"
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

type NamespacedIngressDefaulter struct {
	k8sAPI      *kube.K8sAPI
	configStore *config.Store
}

func NewDefaulter(k8sAPI *kube.K8sAPI, configStore *config.Store) *NamespacedIngressDefaulter {
	return &NamespacedIngressDefaulter{
		k8sAPI:      k8sAPI,
		configStore: configStore,
	}
}

func (w *NamespacedIngressDefaulter) RuntimeObject() runtime.Object {
	return &nsigv1alpha1.NamespacedIngress{}
}

func (w *NamespacedIngressDefaulter) SetDefaults(obj interface{}) {
	c, ok := obj.(*nsigv1alpha1.NamespacedIngress)
	if !ok {
		return
	}

	klog.V(5).Infof("Default Webhook, name=%s", c.Name)
	klog.V(4).Infof("Before setting default values, spec=%#v", c.Spec)

	meshConfig := w.configStore.MeshConfig.GetConfig()

	if meshConfig == nil {
		return
	}

	if c.Spec.ServiceAccountName == "" {
		c.Spec.ServiceAccountName = "erie-canal-namespaced-ingress"
	}

	if c.Spec.LogLevel == nil {
		c.Spec.LogLevel = pointer.Int(2)
	}

	if c.Spec.Replicas == nil {
		c.Spec.Replicas = pointer.Int32(1)
	}

	if c.Spec.TLS.SSLPassthrough.UpstreamPort == nil {
		c.Spec.TLS.SSLPassthrough.UpstreamPort = pointer.Int32(443)
	}

	if c.Spec.PodSecurityContext == nil {
		c.Spec.PodSecurityContext = &corev1.PodSecurityContext{
			RunAsNonRoot: pointer.Bool(true),
			RunAsUser:    pointer.Int64(65532),
			RunAsGroup:   pointer.Int64(65532),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		}
	}

	if c.Spec.SecurityContext == nil {
		c.Spec.SecurityContext = &corev1.SecurityContext{
			AllowPrivilegeEscalation: pointer.Bool(false),
		}
	}

	klog.V(4).Infof("After setting default values, spec=%#v", c.Spec)
}

type NamespacedIngressValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *NamespacedIngressValidator) RuntimeObject() runtime.Object {
	return &nsigv1alpha1.NamespacedIngress{}
}

func (w *NamespacedIngressValidator) ValidateCreate(obj interface{}) error {
	namespacedingress, ok := obj.(*nsigv1alpha1.NamespacedIngress)
	if !ok {
		return nil
	}

	list, err := w.k8sAPI.FlomeshClient.
		NamespacedingressV1alpha1().
		NamespacedIngresses(namespacedingress.Namespace).
		List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return err
	}

	// There's already an NamespacedIngress in this namespace, return error
	if len(list.Items) > 0 {
		return errors.Errorf(
			"There's already %d IngressDeploymnent(s) in namespace %q. Each namespace can have ONLY ONE NamespacedIngress.",
			len(list.Items),
			namespacedingress.Namespace,
		)
	}

	return doValidation(namespacedingress)
}

func (w *NamespacedIngressValidator) ValidateUpdate(oldObj, obj interface{}) error {
	//oldNamespacedIngress, ok := oldObj.(*nsigv1alpha1.NamespacedIngress)
	//if !ok {
	//	return nil
	//}
	//
	//namespacedingress, ok := obj.(*nsigv1alpha1.NamespacedIngress)
	//if !ok {
	//	return nil
	//}
	//
	//if oldNamespacedIngress.Namespace != namespacedingress.Namespace {
	//    return errors.Errorf("")
	//}

	return doValidation(obj)
}

func (w *NamespacedIngressValidator) ValidateDelete(obj interface{}) error {
	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *NamespacedIngressValidator {
	return &NamespacedIngressValidator{
		k8sAPI: k8sAPI,
	}
}

func doValidation(obj interface{}) error {
	return nil
}
