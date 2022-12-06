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

package config

import (
	"context"
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"time"
)

type meshCfgChangeListenerForIngress struct {
	k8sApi      *kube.K8sAPI
	configStore *Store
}

var _ MeshConfigChangeListener = &meshCfgChangeListenerForIngress{}

func (l meshCfgChangeListenerForIngress) OnConfigCreate(cfg *MeshConfig) {
	l.onUpdate(nil, cfg)
}

func (l meshCfgChangeListenerForIngress) OnConfigUpdate(oldCfg, cfg *MeshConfig) {
	l.onUpdate(oldCfg, cfg)
}

func (l meshCfgChangeListenerForIngress) OnConfigDelete(cfg *MeshConfig) {
	l.onUpdate(cfg, nil)
}

func (l meshCfgChangeListenerForIngress) onUpdate(oldCfg, cfg *MeshConfig) {
	if oldCfg == nil {
		oldCfg = l.configStore.MeshConfig.GetConfig()
	}

	if cfg == nil { // cfg is deleted
		cfg = &MeshConfig{}
	}

	klog.V(5).Infof("Operator Config is updated, new values: %#v", l.configStore.MeshConfig)
	//klog.V(5).Infof("Old RepoBaseURL = %q", oldCfg.RepoBaseURL())
	//klog.V(5).Infof("New RepoBaseURL = %q", cfg.RepoBaseURL())
	klog.V(5).Infof("Old IngressCodebasePath = %q", oldCfg.IngressCodebasePath())
	klog.V(5).Infof("New IngressCodebasePath = %q", cfg.IngressCodebasePath())

	// if ingress codebase path is changed, we need to edit ingress-controller deployment
	if oldCfg.IngressCodebasePath() != cfg.IngressCodebasePath() {
		l.updateIngressController(cfg)
	}
}

func (l meshCfgChangeListenerForIngress) updateIngressController(mc *MeshConfig) {
	// patch the deployment spec template triggers the action of rollout restart like with kubectl
	patch := fmt.Sprintf(
		`{"spec": {"template":{"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`,
		time.Now().Format(commons.ProxyProfileLastUpdatedTimeFormat),
	)
	klog.V(5).Infof("patch = %s", patch)

	selector := labels.SelectorFromSet(
		map[string]string{
			"app.kubernetes.io/component": "controller",
			"app.kubernetes.io/instance":  "erie-canal-ingress-pipy",
		},
	)
	ingressList, err := l.k8sApi.Client.AppsV1().
		Deployments(v1.NamespaceAll).
		List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		klog.Errorf("Error listing all ingress-pipy instances: %s", err)
		return
	}

	for _, ing := range ingressList.Items {
		_, err := l.k8sApi.Client.AppsV1().
			Deployments(ing.Namespace).
			Patch(context.TODO(), ing.Name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
		if err != nil {
			klog.Errorf("Patch deployment %s/%s error, %s", ing.Namespace, ing.Name, err)
		}
	}
}
