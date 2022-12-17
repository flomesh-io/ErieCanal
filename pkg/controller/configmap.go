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

package controller

import (
	"fmt"
	cfghandler "github.com/flomesh-io/ErieCanal/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"time"
)

type ConfigMapHandler interface {
	OnConfigMapAdd(cm *corev1.ConfigMap)
	OnConfigMapUpdate(oldCm, cm *corev1.ConfigMap)
	OnConfigMapDelete(cm *corev1.ConfigMap)
	OnConfigMapSynced()
}

type ConfigMapController struct {
	Informer     cache.SharedIndexInformer
	Store        ConfigMapStore
	HasSynced    cache.InformerSynced
	Lister       v1.ConfigMapLister
	eventHandler ConfigMapHandler
}

type ConfigMapFilterFunc func(obj interface{}) bool

type ConfigMapStore struct {
	cache.Store
}

func (l *ConfigMapStore) ByKey(key string) (*corev1.ConfigMap, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*corev1.ConfigMap), nil
}

func NewConfigMapControllerWithEventHandler(configmapInformer coreinformers.ConfigMapInformer, resyncPeriod time.Duration, handler ConfigMapHandler, filter ConfigMapFilterFunc) *ConfigMapController {
	informer := configmapInformer.Informer()

	result := &ConfigMapController{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    configmapInformer.Lister(),
		Store: ConfigMapStore{
			Store: informer.GetStore(),
		},
	}

	if filter == nil {
		informer.AddEventHandlerWithResyncPeriod(
			cache.FilteringResourceEventHandler{
				FilterFunc: cfghandler.DefaultConfigurationFilter,
				Handler: cache.ResourceEventHandlerFuncs{
					AddFunc:    result.handleAddConfigMap,
					UpdateFunc: result.handleUpdateConfigMap,
					DeleteFunc: result.handleDeleteConfigMap,
				},
			},
			resyncPeriod,
		)
	} else {
		informer.AddEventHandlerWithResyncPeriod(
			cache.FilteringResourceEventHandler{
				FilterFunc: filter,
				Handler: cache.ResourceEventHandlerFuncs{
					AddFunc:    result.handleAddConfigMap,
					UpdateFunc: result.handleUpdateConfigMap,
					DeleteFunc: result.handleDeleteConfigMap,
				},
			},
			resyncPeriod,
		)
	}

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *ConfigMapController) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting configmap config controller")

	if !cache.WaitForNamedCacheSync("configmap config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnConfigMapSynced()")
		c.eventHandler.OnConfigMapSynced()
	}
}

func (c *ConfigMapController) handleAddConfigMap(obj interface{}) {
	configmap, ok := obj.(*corev1.ConfigMap)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnConfigMapAdd")
		c.eventHandler.OnConfigMapAdd(configmap)
	}
}

func (c *ConfigMapController) handleUpdateConfigMap(oldObj, newObj interface{}) {
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

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnConfigMapUpdate")
		c.eventHandler.OnConfigMapUpdate(oldConfigMap, configmap)
	}
}

func (c *ConfigMapController) handleDeleteConfigMap(obj interface{}) {
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

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnConfigMapDelete")
		c.eventHandler.OnConfigMapDelete(configmap)
	}
}
