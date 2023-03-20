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

package globaltrafficpolicy

import (
	"fmt"
	gtpv1alpha1 "github.com/flomesh-io/ErieCanal/apis/globaltrafficpolicy/v1alpha1"
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
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
		[]string{"flomesh.io"},
		[]string{"v1alpha1"},
		[]string{"globaltrafficpolicies"},
	)

	return []admissionregv1.MutatingWebhook{flomeshadmission.NewMutatingWebhook(
			"mglobaltrafficpolicy.kb.flomesh.io",
			r.WebhookSvcNs,
			r.WebhookSvcName,
			commons.GlobalTrafficPolicyMutatingWebhookPath,
			r.CaBundle,
			nil,
			[]admissionregv1.RuleWithOperations{rule},
		)}, []admissionregv1.ValidatingWebhook{flomeshadmission.NewValidatingWebhook(
			"vglobaltrafficpolicy.kb.flomesh.io",
			r.WebhookSvcNs,
			r.WebhookSvcName,
			commons.GlobalTrafficPolicyValidatingWebhookPath,
			r.CaBundle,
			nil,
			[]admissionregv1.RuleWithOperations{rule},
		)}
}

func (r *register) GetHandlers() map[string]http.Handler {
	return map[string]http.Handler{
		commons.GlobalTrafficPolicyMutatingWebhookPath:   webhooks.DefaultingWebhookFor(newDefaulter(r.K8sAPI, r.ConfigStore)),
		commons.GlobalTrafficPolicyValidatingWebhookPath: webhooks.ValidatingWebhookFor(newValidator(r.K8sAPI)),
	}
}

type defaulter struct {
	k8sAPI      *kube.K8sAPI
	configStore *config.Store
}

func newDefaulter(k8sAPI *kube.K8sAPI, configStore *config.Store) *defaulter {
	return &defaulter{
		k8sAPI:      k8sAPI,
		configStore: configStore,
	}
}

func (w *defaulter) RuntimeObject() runtime.Object {
	return &gtpv1alpha1.GlobalTrafficPolicy{}
}

func (w *defaulter) SetDefaults(obj interface{}) {
	policy, ok := obj.(*gtpv1alpha1.GlobalTrafficPolicy)
	if !ok {
		return
	}

	klog.V(5).Infof("Default Webhook, name=%s", policy.Name)
	klog.V(4).Infof("Before setting default values, spec=%#v", policy.Spec)

	meshConfig := w.configStore.MeshConfig.GetConfig()

	if meshConfig == nil {
		return
	}

	if policy.Spec.LbType == "" {
		policy.Spec.LbType = gtpv1alpha1.LocalityLbType
	}

	klog.V(4).Infof("After setting default values, spec=%#v", policy.Spec)
}

type validator struct {
	k8sAPI *kube.K8sAPI
}

func (w *validator) RuntimeObject() runtime.Object {
	return &gtpv1alpha1.GlobalTrafficPolicy{}
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
	policy, ok := obj.(*gtpv1alpha1.GlobalTrafficPolicy)
	if !ok {
		return nil
	}

	switch policy.Spec.LbType {
	case gtpv1alpha1.LocalityLbType:
		if len(policy.Spec.Targets) > 1 {
			return fmt.Errorf("in case of Locality load balancer, the traffic can only be sticky to exact one cluster, either in cluster or a specific remote cluster")
		}
	case gtpv1alpha1.FailOverLbType:
		if len(policy.Spec.Targets) == 0 {
			return fmt.Errorf("requires at least one cluster for failover")
		}
	case gtpv1alpha1.ActiveActiveLbType:
		//if len(policy.Spec.Targets) == 0 {
		//	return fmt.Errorf("requires at least another one cluster for active-active load balancing")
		//}

		for _, t := range policy.Spec.Targets {
			if *t.Weight < 0 {
				return fmt.Errorf("weight %d of %s is invalid for active-active load balancing, it must be >= 0", t.Weight, t.ClusterKey)
			}
		}
	default:

	}

	return nil
}
