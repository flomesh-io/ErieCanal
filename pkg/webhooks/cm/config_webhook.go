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
