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

package cm

import (
	"fmt"
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"strings"
)

const (
	kind      = "ConfigMap"
	groups    = ""
	resources = "configmaps"
	versions  = "v1"

	mwPath = commons.ConfigMapMutatingWebhookPath
	mwName = "mconfigmap.kb.flomesh.io"
	vwPath = commons.ConfigMapValidatingWebhookPath
	vwName = "vconfigmap.kb.flomesh.io"
)

func RegisterWebhooks(webhookSvcNs, webhookSvcName string, caBundle []byte) {
	rule := flomeshadmission.NewRule(
		[]admissionregv1.OperationType{admissionregv1.Create, admissionregv1.Update},
		[]string{groups},
		[]string{versions},
		[]string{resources},
	)

	nsSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			commons.FlomeshControlPlaneLabel: "true",
		},
	}

	mutatingWebhook := flomeshadmission.NewMutatingWebhook(
		mwName,
		webhookSvcNs,
		webhookSvcName,
		mwPath,
		caBundle,
		nsSelector,
		[]admissionregv1.RuleWithOperations{rule},
	)

	validatingWebhook := flomeshadmission.NewValidatingWebhook(
		vwName,
		webhookSvcNs,
		webhookSvcName,
		vwPath,
		caBundle,
		nsSelector,
		[]admissionregv1.RuleWithOperations{rule},
	)

	flomeshadmission.RegisterMutatingWebhook(mwName, mutatingWebhook)
	flomeshadmission.RegisterValidatingWebhook(vwName, validatingWebhook)
}

type ConfigMapDefaulter struct {
	k8sAPI *kube.K8sAPI
}

func isNotWatchedConfigmap(cm *corev1.ConfigMap, erieCanalNamespace string) bool {
	klog.V(5).Infof("Configmap namespace = %q, name = %q.", cm.Namespace, cm.Name)
	return cm.Namespace != erieCanalNamespace || !config.DefaultWatchedConfigMaps.Has(cm.Name)
}

func NewDefaulter(k8sAPI *kube.K8sAPI) *ConfigMapDefaulter {
	return &ConfigMapDefaulter{
		k8sAPI: k8sAPI,
	}
}

func (w *ConfigMapDefaulter) RuntimeObject() runtime.Object {
	return &corev1.ConfigMap{}
}

func (w *ConfigMapDefaulter) SetDefaults(obj interface{}) {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return
	}

	if isNotWatchedConfigmap(cm, config.GetErieCanalNamespace()) {
		return
	}

	switch cm.Name {
	case commons.MeshConfigName:
		cfg, err := config.ParseMeshConfig(cm)
		if err != nil {
			return
		}

		if cfg.Images.Repository == "" {
			cfg.Images.Repository = "flomesh"
		}

		if cfg.Images.PipyImage == "" {
			cfg.Images.PipyImage = "pipy:latest"
		}

		if cfg.Images.ProxyInitImage == "" {
			cfg.Images.ProxyInitImage = "erie-canal-proxy-init:latest"
		}

		if cfg.Images.KlipperLbImage == "" {
			cfg.Images.KlipperLbImage = "mirrored-klipper-lb:v0.3.5"
		}

		if strings.HasSuffix(cfg.Repo.RootURL, "/") {
			cfg.Repo.RootURL = strings.TrimSuffix(cfg.Repo.RootURL, "/")
		}

		if cfg.Certificate.Manager == "" {
			cfg.Certificate.Manager = string(certificate.Archon)
		}

		if cfg.Webhook.ServiceName == "" {
			cfg.Webhook.ServiceName = commons.DefaultWebhookServiceName
		}

		cm.Data[commons.MeshConfigJsonName] = cfg.ToJson()
	default:
		// ignore
	}
}

type ConfigMapValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *ConfigMapValidator) RuntimeObject() runtime.Object {
	return &corev1.ConfigMap{}
}

func (w *ConfigMapValidator) ValidateCreate(obj interface{}) error {
	return w.doValidation(obj)
}

func (w *ConfigMapValidator) ValidateUpdate(oldObj, obj interface{}) error {
	return w.doValidation(obj)
}

func (w *ConfigMapValidator) ValidateDelete(obj interface{}) error {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil
	}

	if isNotWatchedConfigmap(cm, config.GetErieCanalNamespace()) {
		return nil
	}

	switch cm.Name {
	case commons.MeshConfigName:
		// protect the MeshConfig from deletion
		return fmt.Errorf("ConfigMap %s/%s cannot be deleted", cm.Namespace, cm.Name)
	default:
		// ignore
	}

	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *ConfigMapValidator {
	return &ConfigMapValidator{
		k8sAPI: k8sAPI,
	}
}

func (w *ConfigMapValidator) doValidation(obj interface{}) error {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil
	}

	if isNotWatchedConfigmap(cm, config.GetErieCanalNamespace()) {
		return nil
	}

	switch cm.Name {
	case commons.MeshConfigName:
		// validate the config
	default:
		// ignore
	}

	return nil
}
