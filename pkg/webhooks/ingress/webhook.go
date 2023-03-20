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
	"github.com/flomesh-io/ErieCanal/pkg/webhooks"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
)

type register struct {
	*webhooks.RegisterConfig
}

func NewRegister(cfg *webhooks.RegisterConfig) webhooks.Register {
	return &register{
		RegisterConfig: cfg,
	}
}

func (r *register) GetWebhooks() ([]admissionregv1.MutatingWebhook, []admissionregv1.ValidatingWebhook) {
	rule := flomeshadmission.NewRule(
		[]admissionregv1.OperationType{admissionregv1.Create, admissionregv1.Update},
		[]string{"networking.k8s.io"},
		[]string{"v1"},
		[]string{"ingresses"},
	)

	return []admissionregv1.MutatingWebhook{flomeshadmission.NewMutatingWebhook(
			"mingress.kb.flomesh.io",
			r.WebhookSvcNs,
			r.WebhookSvcName,
			commons.IngressMutatingWebhookPath,
			r.CaBundle,
			nil,
			[]admissionregv1.RuleWithOperations{rule},
		)}, []admissionregv1.ValidatingWebhook{flomeshadmission.NewValidatingWebhook(
			"vingress.kb.flomesh.io",
			r.WebhookSvcNs,
			r.WebhookSvcName,
			commons.IngressValidatingWebhookPath,
			r.CaBundle,
			nil,
			[]admissionregv1.RuleWithOperations{rule},
		)}
}

func (r *register) GetHandlers() map[string]http.Handler {
	return map[string]http.Handler{
		commons.IngressMutatingWebhookPath:   webhooks.DefaultingWebhookFor(newDefaulter(r.K8sAPI)),
		commons.IngressValidatingWebhookPath: webhooks.ValidatingWebhookFor(newValidator(r.K8sAPI)),
	}
}

type defaulter struct {
	k8sAPI *kube.K8sAPI
}

func newDefaulter(k8sAPI *kube.K8sAPI) *defaulter {
	return &defaulter{
		k8sAPI: k8sAPI,
	}
}

func (w *defaulter) RuntimeObject() runtime.Object {
	return &networkingv1.Ingress{}
}

func (w *defaulter) SetDefaults(obj interface{}) {
	ing, ok := obj.(*networkingv1.Ingress)
	if !ok {
		return
	}

	if !ingresspipy.IsValidPipyIngress(ing) {
		return
	}

}

type validator struct {
	k8sAPI *kube.K8sAPI
}

func (w *validator) RuntimeObject() runtime.Object {
	return &networkingv1.Ingress{}
}

func (w *validator) ValidateCreate(obj interface{}) error {
	return w.doValidation(obj)
}

func (w *validator) ValidateUpdate(oldObj, obj interface{}) error {
	return w.doValidation(obj)
}

func (w *validator) ValidateDelete(obj interface{}) error {
	return nil
}

func newValidator(k8sAPI *kube.K8sAPI) *validator {
	return &validator{
		k8sAPI: k8sAPI,
	}
}

func (w *validator) doValidation(obj interface{}) error {
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

func (w *validator) secretExists(secretName string, ing *networkingv1.Ingress) error {
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
