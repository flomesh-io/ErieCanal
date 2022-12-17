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

package v1beta1

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwinformerv1beta1 "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions/apis/v1beta1"
	gwlisterv1beta1 "sigs.k8s.io/gateway-api/pkg/client/listers/apis/v1beta1"
	"time"
)

type GatewayClassHandler interface {
	OnGatewayClassAdd(gatewayClass *gwv1alpha2.GatewayClass)
	OnGatewayClassUpdate(oldGatewayClass, gatewayClass *gwv1alpha2.GatewayClass)
	OnGatewayClassDelete(gatewayClass *gwv1alpha2.GatewayClass)
	OnGatewayClassSynced()
}

type GatewayClassController struct {
	Informer     cache.SharedIndexInformer
	Store        GatewayClassStore
	HasSynced    cache.InformerSynced
	Lister       gwlisterv1beta1.GatewayClassLister
	eventHandler GatewayClassHandler
}

type GatewayClassStore struct {
	cache.Store
}

func (l *GatewayClassStore) ByKey(key string) (*gwv1alpha2.GatewayClass, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*gwv1alpha2.GatewayClass), nil
}

func NewGatewayClassControllerWithEventHandler(gatewayClassInformer gwinformerv1beta1.GatewayClassInformer, resyncPeriod time.Duration, handler GatewayClassHandler) *GatewayClassController {
	informer := gatewayClassInformer.Informer()

	result := &GatewayClassController{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    gatewayClassInformer.Lister(),
		Store: GatewayClassStore{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddGatewayClass,
			UpdateFunc: result.handleUpdateGatewayClass,
			DeleteFunc: result.handleDeleteGatewayClass,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *GatewayClassController) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting GatewayClass config controller")

	if !cache.WaitForNamedCacheSync("GatewayClass config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnGatewayClassSynced()")
		c.eventHandler.OnGatewayClassSynced()
	}
}

func (c *GatewayClassController) handleAddGatewayClass(obj interface{}) {
	gatewayClass, ok := obj.(*gwv1alpha2.GatewayClass)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnGatewayClassAdd")
		c.eventHandler.OnGatewayClassAdd(gatewayClass)
	}
}

func (c *GatewayClassController) handleUpdateGatewayClass(oldObj, newObj interface{}) {
	oldGatewayClass, ok := oldObj.(*gwv1alpha2.GatewayClass)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	gatewayClass, ok := newObj.(*gwv1alpha2.GatewayClass)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnGatewayClassUpdate")
		c.eventHandler.OnGatewayClassUpdate(oldGatewayClass, gatewayClass)
	}
}

func (c *GatewayClassController) handleDeleteGatewayClass(obj interface{}) {
	gatewayClass, ok := obj.(*gwv1alpha2.GatewayClass)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if gatewayClass, ok = tombstone.Obj.(*gwv1alpha2.GatewayClass); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}
	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnGatewayClassDelete")
		c.eventHandler.OnGatewayClassDelete(gatewayClass)
	}
}
