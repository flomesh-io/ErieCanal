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

package admission

import (
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewMutatingWebhookConfiguration(webhooks []admissionregv1.MutatingWebhook) *admissionregv1.MutatingWebhookConfiguration {
	if len(webhooks) == 0 {
		return nil
	}

	return &admissionregv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: commons.DefaultMutatingWebhookConfigurationName,
		},
		Webhooks: webhooks,
	}
}

func NewValidatingWebhookConfiguration(webhooks []admissionregv1.ValidatingWebhook) *admissionregv1.ValidatingWebhookConfiguration {
	if len(webhooks) == 0 {
		return nil
	}

	return &admissionregv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: commons.DefaultValidatingWebhookConfigurationName,
		},
		Webhooks: webhooks,
	}
}

func NewMutatingWebhook(
	mutatingWebhookName,
	webhookServiceNamespace,
	webhookServiceName,
	webhookPath string,
	caBundle []byte,
	namespaceSelector *metav1.LabelSelector,
	rules []admissionregv1.RuleWithOperations,
) admissionregv1.MutatingWebhook {
	failurePolicy := admissionregv1.Fail
	matchPolicy := admissionregv1.Exact
	sideEffect := admissionregv1.SideEffectClassNone

	result := admissionregv1.MutatingWebhook{
		Name: mutatingWebhookName,
		ClientConfig: admissionregv1.WebhookClientConfig{
			Service: &admissionregv1.ServiceReference{
				Namespace: webhookServiceNamespace,
				Name:      webhookServiceName,
				Path:      &webhookPath,
			},
			CABundle: caBundle,
		},
		FailurePolicy:           &failurePolicy,
		MatchPolicy:             &matchPolicy,
		Rules:                   rules,
		SideEffects:             &sideEffect,
		AdmissionReviewVersions: []string{"v1"},
	}

	if namespaceSelector != nil {
		result.NamespaceSelector = namespaceSelector
	}

	return result
}

func NewValidatingWebhook(
	validatingWebhookName,
	webhookServiceNamespace,
	webhookServiceName,
	webhookPath string,
	caBundle []byte,
	namespaceSelector *metav1.LabelSelector,
	rules []admissionregv1.RuleWithOperations,
) admissionregv1.ValidatingWebhook {
	failurePolicy := admissionregv1.Fail
	matchPolicy := admissionregv1.Exact
	sideEffect := admissionregv1.SideEffectClassNone

	result := admissionregv1.ValidatingWebhook{
		Name: validatingWebhookName,
		ClientConfig: admissionregv1.WebhookClientConfig{
			Service: &admissionregv1.ServiceReference{
				Namespace: webhookServiceNamespace,
				Name:      webhookServiceName,
				Path:      &webhookPath,
			},
			CABundle: caBundle,
		},
		FailurePolicy:           &failurePolicy,
		MatchPolicy:             &matchPolicy,
		Rules:                   rules,
		SideEffects:             &sideEffect,
		AdmissionReviewVersions: []string{"v1"},
	}

	if namespaceSelector != nil {
		result.NamespaceSelector = namespaceSelector
	}

	return result
}

func NewRule(
	operations []admissionregv1.OperationType,
	apiGroups, apiVersions, resources []string,
) admissionregv1.RuleWithOperations {
	return admissionregv1.RuleWithOperations{
		Operations: operations,
		Rule: admissionregv1.Rule{
			APIGroups:   apiGroups,
			APIVersions: apiVersions,
			Resources:   resources,
		},
	}
}
