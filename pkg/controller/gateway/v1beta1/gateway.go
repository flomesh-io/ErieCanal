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

type GatewayHandler interface {
	OnGatewayAdd(gateway *gwv1alpha2.Gateway)
	OnGatewayUpdate(oldGateway, gateway *gwv1alpha2.Gateway)
	OnGatewayDelete(gateway *gwv1alpha2.Gateway)
	OnGatewaySynced()
}

type GatewayController struct {
	Informer     cache.SharedIndexInformer
	Store        GatewayStore
	HasSynced    cache.InformerSynced
	Lister       gwlisterv1beta1.GatewayLister
	eventHandler GatewayHandler
}

type GatewayStore struct {
	cache.Store
}

func (l *GatewayStore) ByKey(key string) (*gwv1alpha2.Gateway, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*gwv1alpha2.Gateway), nil
}

func NewGatewayControllerWithEventHandler(gatewayInformer gwinformerv1beta1.GatewayInformer, resyncPeriod time.Duration, handler GatewayHandler) *GatewayController {
	informer := gatewayInformer.Informer()

	result := &GatewayController{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    gatewayInformer.Lister(),
		Store: GatewayStore{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddGateway,
			UpdateFunc: result.handleUpdateGateway,
			DeleteFunc: result.handleDeleteGateway,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *GatewayController) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting Gateway config controller")

	if !cache.WaitForNamedCacheSync("Gateway config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnGatewaySynced()")
		c.eventHandler.OnGatewaySynced()
	}
}

func (c *GatewayController) handleAddGateway(obj interface{}) {
	gateway, ok := obj.(*gwv1alpha2.Gateway)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnGatewayAdd")
		c.eventHandler.OnGatewayAdd(gateway)
	}
}

func (c *GatewayController) handleUpdateGateway(oldObj, newObj interface{}) {
	oldGateway, ok := oldObj.(*gwv1alpha2.Gateway)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	gateway, ok := newObj.(*gwv1alpha2.Gateway)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnGatewayUpdate")
		c.eventHandler.OnGatewayUpdate(oldGateway, gateway)
	}
}

func (c *GatewayController) handleDeleteGateway(obj interface{}) {
	gateway, ok := obj.(*gwv1alpha2.Gateway)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if gateway, ok = tombstone.Obj.(*gwv1alpha2.Gateway); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}
	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnGatewayDelete")
		c.eventHandler.OnGatewayDelete(gateway)
	}
}
