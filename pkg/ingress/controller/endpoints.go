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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"time"
)

type EndpointsHandler interface {
	OnEndpointsAdd(endpoints *corev1.Endpoints)
	OnEndpointsUpdate(oldEndpoints, endpoints *corev1.Endpoints)
	OnEndpointsDelete(endpoints *corev1.Endpoints)
	OnEndpointsSynced()
}

type EndpointsController struct {
	Informer     cache.SharedIndexInformer
	Store        EndpointStore
	HasSynced    cache.InformerSynced
	Lister       v1.EndpointsLister
	eventHandler EndpointsHandler
}

type EndpointStore struct {
	cache.Store
}

func (l *EndpointStore) ByKey(key string) (*corev1.Endpoints, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*corev1.Endpoints), nil
}

func NewEndpointsControllerWithEventHandler(endpointsInformer coreinformers.EndpointsInformer, resyncPeriod time.Duration, handler EndpointsHandler) *EndpointsController {
	informer := endpointsInformer.Informer()

	result := &EndpointsController{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    endpointsInformer.Lister(),
		Store: EndpointStore{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddEndpoints,
			UpdateFunc: result.handleUpdateEndpoints,
			DeleteFunc: result.handleDeleteEndpoints,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *EndpointsController) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting endpoints config controller")

	if !cache.WaitForNamedCacheSync("endpoints config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnEndpointsSynced()")
		c.eventHandler.OnEndpointsSynced()
	}
}

func (c *EndpointsController) handleAddEndpoints(obj interface{}) {
	endpoints, ok := obj.(*corev1.Endpoints)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnEndpointsAdd")
		c.eventHandler.OnEndpointsAdd(endpoints)
	}
}

func (c *EndpointsController) handleUpdateEndpoints(oldObj, newObj interface{}) {
	oldEndpoints, ok := oldObj.(*corev1.Endpoints)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	endpoints, ok := newObj.(*corev1.Endpoints)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnEndpointsUpdate")
		c.eventHandler.OnEndpointsUpdate(oldEndpoints, endpoints)
	}
}

func (c *EndpointsController) handleDeleteEndpoints(obj interface{}) {
	endpoints, ok := obj.(*corev1.Endpoints)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if endpoints, ok = tombstone.Obj.(*corev1.Endpoints); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnEndpointsDelete")
		c.eventHandler.OnEndpointsDelete(endpoints)
	}
}
