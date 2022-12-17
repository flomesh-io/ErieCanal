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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	ctrlruntimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"time"
)

type ConfigEventHandler interface {
	OnConfigMapAdd(cm *corev1.ConfigMap)
	OnConfigMapUpdate(oldCm, cm *corev1.ConfigMap)
	OnConfigMapDelete(cm *corev1.ConfigMap)
}

type MeshConfigChangeListener interface {
	OnConfigCreate(cfg *MeshConfig)
	OnConfigUpdate(oldCfg, cfg *MeshConfig)
	OnConfigDelete(cfg *MeshConfig)
}

//type ClusterConfigChangeListener interface {
//	OnConfigCreate(cfg *ClusterConfig)
//	OnConfigUpdate(oldCfg, cfg *ClusterConfig)
//	OnConfigDelete(cfg *ClusterConfig)
//}

type flomeshConfigurationHanlder struct {
	handler ConfigEventHandler
}

func (c *flomeshConfigurationHanlder) handleAddConfigMap(obj interface{}) {
	configmap, ok := obj.(*corev1.ConfigMap)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	klog.V(4).InfoS("Calling handler.OnConfigMapAdd")
	c.handler.OnConfigMapAdd(configmap)
}

func (c *flomeshConfigurationHanlder) handleUpdateConfigMap(oldObj, newObj interface{}) {
	oldConfigMap, ok := oldObj.(*corev1.ConfigMap)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	configmap, ok := newObj.(*corev1.ConfigMap)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	klog.V(4).InfoS("Calling handler.OnConfigMapUpdate")
	c.handler.OnConfigMapUpdate(oldConfigMap, configmap)
}

func (c *flomeshConfigurationHanlder) handleDeleteConfigMap(obj interface{}) {
	configmap, ok := obj.(*corev1.ConfigMap)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if configmap, ok = tombstone.Obj.(*corev1.ConfigMap); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}
	klog.V(4).InfoS("Calling handler.OnConfigMapDelete")
	c.handler.OnConfigMapDelete(configmap)
}

func RegisterConfigurationHanlder(handler ConfigEventHandler, configmapInformer ctrlruntimecache.Informer, resyncPeriod time.Duration) {
	internalHandler := &flomeshConfigurationHanlder{
		handler: handler,
	}

	configMapEventHandler := &cache.ResourceEventHandlerFuncs{
		AddFunc:    internalHandler.handleAddConfigMap,
		UpdateFunc: internalHandler.handleUpdateConfigMap,
		DeleteFunc: internalHandler.handleDeleteConfigMap,
	}

	filteringResourceEventHandler := cache.FilteringResourceEventHandler{
		FilterFunc: DefaultConfigurationFilter,
		Handler:    configMapEventHandler,
	}

	configmapInformer.AddEventHandlerWithResyncPeriod(filteringResourceEventHandler, resyncPeriod)
}

func DefaultConfigurationFilter(obj interface{}) bool {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	if cm.Namespace == GetErieCanalNamespace() && DefaultWatchedConfigMaps.Has(cm.Name) {
		return true
	}

	return false
}
