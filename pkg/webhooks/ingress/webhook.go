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

package ingress

import (
	"context"
	"fmt"
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	ingresspipy "github.com/flomesh-io/ErieCanal/pkg/ingress"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/util"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	kind      = "Ingress"
	groups    = "networking.k8s.io"
	resources = "ingresses"
	versions  = "v1"

	mwPath = commons.IngressMutatingWebhookPath
	mwName = "mingress.kb.flomesh.io"
	vwPath = commons.IngressValidatingWebhookPath
	vwName = "vingress.kb.flomesh.io"
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

type IngressDefaulter struct {
	k8sAPI *kube.K8sAPI
}

func NewDefaulter(k8sAPI *kube.K8sAPI) *IngressDefaulter {
	return &IngressDefaulter{
		k8sAPI: k8sAPI,
	}
}

func (w *IngressDefaulter) RuntimeObject() runtime.Object {
	return &networkingv1.Ingress{}
}

func (w *IngressDefaulter) SetDefaults(obj interface{}) {
	ing, ok := obj.(*networkingv1.Ingress)
	if !ok {
		return
	}

	if !ingresspipy.IsValidPipyIngress(ing) {
		return
	}

}

type IngressValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *IngressValidator) RuntimeObject() runtime.Object {
	return &networkingv1.Ingress{}
}

func (w *IngressValidator) ValidateCreate(obj interface{}) error {
	return w.doValidation(obj)
}

func (w *IngressValidator) ValidateUpdate(oldObj, obj interface{}) error {
	return w.doValidation(obj)
}

func (w *IngressValidator) ValidateDelete(obj interface{}) error {
	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *IngressValidator {
	return &IngressValidator{
		k8sAPI: k8sAPI,
	}
}

func (w *IngressValidator) doValidation(obj interface{}) error {
	ing, ok := obj.(*networkingv1.Ingress)
	if !ok {
		return nil
	}

	if !ingresspipy.IsValidPipyIngress(ing) {
		return nil
	}

	upstreamSSLSecret := ing.Annotations[ingresspipy.PipyIngressAnnotationUpstreamSSLSecret]
	if upstreamSSLSecret != "" {
		if err := w.secretExists(upstreamSSLSecret, ing); err != nil {
			return fmt.Errorf("secert %q doesn't exist: %s, please check annotation 'pipy.ingress.kubernetes.io/upstream-ssl-secret' of Ingress %s/%s", upstreamSSLSecret, err, ing.Namespace, ing.Name)
		}
	}

	trustedCASecret := ing.Annotations[ingresspipy.PipyIngressAnnotationTLSTrustedCASecret]
	if trustedCASecret != "" {
		if err := w.secretExists(trustedCASecret, ing); err != nil {
			return fmt.Errorf("secert %q doesn't exist: %s, please check annotation 'pipy.ingress.kubernetes.io/tls-trusted-ca-secret' of Ingress %s/%s", trustedCASecret, err, ing.Namespace, ing.Name)
		}
	}

	for _, tls := range ing.Spec.TLS {
		if tls.SecretName == "" {
			continue
		}

		if err := w.secretExists(tls.SecretName, ing); err != nil {
			return fmt.Errorf("TLS secret %q of Ingress %s/%s doesn't exist, please check spec.tls section of Ingress", tls.SecretName, ing.Namespace, ing.Name)
		}
	}

	return nil
}

func (w *IngressValidator) secretExists(secretName string, ing *networkingv1.Ingress) error {
	ns, name, err := util.SecretNamespaceAndName(secretName, ing)
	if err != nil {
		return err
	}

	if name == "" {
		return fmt.Errorf("secret name of Ingress %s/%s is empty or invalid", ing.Namespace, ing.Name)
	}

	if _, err := w.k8sAPI.Client.CoreV1().
		Secrets(ns).
		Get(context.TODO(), name, metav1.GetOptions{}); err != nil {
		return err
	}

	return nil
}
