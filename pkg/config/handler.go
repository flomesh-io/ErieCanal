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

package config

import (
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type configChangeListener struct {
	meshConfig []MeshConfigChangeListener
}

type FlomeshConfigurationHandler struct {
	configStore *Store
	listeners   *configChangeListener
}

var _ ConfigEventHandler = &FlomeshConfigurationHandler{}

func NewFlomeshConfigurationHandler(client client.Client, k8sApi *kube.K8sAPI, store *Store) *FlomeshConfigurationHandler {
	return &FlomeshConfigurationHandler{
		configStore: store,
		listeners: &configChangeListener{
			meshConfig: []MeshConfigChangeListener{
				&meshCfgChangeListenerForBasicConfig{client: client, k8sApi: k8sApi, configStore: store},
				&meshCfgChangeListenerForIngress{k8sApi: k8sApi, configStore: store},
			},
			//clusterConfig: []ClusterConfigChangeListener{},
		},
	}
}

func (f FlomeshConfigurationHandler) OnConfigMapAdd(cm *corev1.ConfigMap) {
	klog.V(5).Infof("OnConfigMapAdd(), ConfigMap namespace = %q, name = %q", cm.Namespace, cm.Name)

	switch cm.Name {
	case commons.MeshConfigName:
		// create the config, and set default values according to the cm
		cfg, err := ParseMeshConfig(cm)
		if err != nil {
			return
		}

		for _, listener := range f.listeners.meshConfig {
			go listener.OnConfigCreate(cfg)
		}
	default:
		//ignore
	}
}

func (f FlomeshConfigurationHandler) OnConfigMapUpdate(oldCm, cm *corev1.ConfigMap) {
	klog.V(5).Infof("OnConfigMapUpdate(), ConfigMap namespace = %q, name = %q", cm.Namespace, cm.Name)

	switch cm.Name {
	case commons.MeshConfigName:
		// update the config
		oldCfg, err := ParseMeshConfig(oldCm)
		if err != nil {
			return
		}

		cfg, err := ParseMeshConfig(cm)
		if err != nil {
			return
		}

		for _, listener := range f.listeners.meshConfig {
			go listener.OnConfigUpdate(oldCfg, cfg)
		}
	default:
		//ignore
	}
}

func (f FlomeshConfigurationHandler) OnConfigMapDelete(cm *corev1.ConfigMap) {
	klog.V(5).Infof("OnConfigMapDelete(), ConfigMap namespace = %q, name = %q", cm.Namespace, cm.Name)

	switch cm.Name {
	case commons.MeshConfigName:
		// Reset the config to default values
		// Actually for now, as ingress-controller mounts the erie-canal-mesh-config, if it's deleted will cause an error
		//f.updateMeshConfig(nil)
		cfg, err := ParseMeshConfig(cm)
		if err != nil {
			return
		}

		for _, listener := range f.listeners.meshConfig {
			go listener.OnConfigDelete(cfg)
		}

		klog.V(5).Infof("Operator Config is reverted to default, new values: %#v", f.configStore.MeshConfig)
	default:
		//ignore
	}
}
