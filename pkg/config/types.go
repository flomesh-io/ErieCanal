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
