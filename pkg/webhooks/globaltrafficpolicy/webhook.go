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

package globaltrafficpolicy

import (
	"fmt"
	gtpv1alpha1 "github.com/flomesh-io/ErieCanal/apis/globaltrafficpolicy/v1alpha1"
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

const (
	kind      = "GlobalTrafficPolicy"
	groups    = "flomesh.io"
	resources = "globaltrafficpolicies"
	versions  = "v1alpha1"

	mwPath = commons.GlobalTrafficPolicyMutatingWebhookPath
	mwName = "mglobaltrafficpolicy.kb.flomesh.io"
	vwPath = commons.GlobalTrafficPolicyValidatingWebhookPath
	vwName = "vglobaltrafficpolicy.kb.flomesh.io"
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

type GlobalTrafficPolicyDefaulter struct {
	k8sAPI      *kube.K8sAPI
	configStore *config.Store
}

func NewDefaulter(k8sAPI *kube.K8sAPI, configStore *config.Store) *GlobalTrafficPolicyDefaulter {
	return &GlobalTrafficPolicyDefaulter{
		k8sAPI:      k8sAPI,
		configStore: configStore,
	}
}

func (w *GlobalTrafficPolicyDefaulter) RuntimeObject() runtime.Object {
	return &gtpv1alpha1.GlobalTrafficPolicy{}
}

func (w *GlobalTrafficPolicyDefaulter) SetDefaults(obj interface{}) {
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

type GlobalTrafficPolicyValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *GlobalTrafficPolicyValidator) RuntimeObject() runtime.Object {
	return &gtpv1alpha1.GlobalTrafficPolicy{}
}

func (w *GlobalTrafficPolicyValidator) ValidateCreate(obj interface{}) error {
	return w.doValidation(obj)
}

func (w *GlobalTrafficPolicyValidator) ValidateUpdate(oldObj, obj interface{}) error {
	return w.doValidation(obj)
}

func (w *GlobalTrafficPolicyValidator) ValidateDelete(obj interface{}) error {
	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *GlobalTrafficPolicyValidator {
	return &GlobalTrafficPolicyValidator{
		k8sAPI: k8sAPI,
	}
}

func (w *GlobalTrafficPolicyValidator) doValidation(obj interface{}) error {
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
			if t.Weight < 0 {
				return fmt.Errorf("weight %d of %s is invalid for active-active load balancing, it must be >= 0", t.Weight, t.ClusterKey)
			}
		}
	default:

	}

	return nil
}
