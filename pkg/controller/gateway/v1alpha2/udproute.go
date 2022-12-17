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

package v1alpha2

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwinformerv1alpha2 "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions/apis/v1alpha2"
	gwlisterv1alpha2 "sigs.k8s.io/gateway-api/pkg/client/listers/apis/v1alpha2"
	"time"
)

type UDPRouteHandler interface {
	OnUDPRouteAdd(udpRoute *gwv1alpha2.UDPRoute)
	OnUDPRouteUpdate(oldUdpRoute, udpRoute *gwv1alpha2.UDPRoute)
	OnUDPRouteDelete(udpRoute *gwv1alpha2.UDPRoute)
	OnUDPRouteSynced()
}

type UDPRouteController struct {
	Informer     cache.SharedIndexInformer
	Store        UDPRouteStore
	HasSynced    cache.InformerSynced
	Lister       gwlisterv1alpha2.UDPRouteLister
	eventHandler UDPRouteHandler
}

type UDPRouteStore struct {
	cache.Store
}

func (l *UDPRouteStore) ByKey(key string) (*gwv1alpha2.UDPRoute, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*gwv1alpha2.UDPRoute), nil
}

func NewUDPRouteControllerWithEventHandler(udpRouteInformer gwinformerv1alpha2.UDPRouteInformer, resyncPeriod time.Duration, handler UDPRouteHandler) *UDPRouteController {
	informer := udpRouteInformer.Informer()

	result := &UDPRouteController{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    udpRouteInformer.Lister(),
		Store: UDPRouteStore{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddUDPRoute,
			UpdateFunc: result.handleUpdateUDPRoute,
			DeleteFunc: result.handleDeleteUDPRoute,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *UDPRouteController) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting UDPRoute config controller")

	if !cache.WaitForNamedCacheSync("UDPRoute config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnUDPRouteSynced()")
		c.eventHandler.OnUDPRouteSynced()
	}
}

func (c *UDPRouteController) handleAddUDPRoute(obj interface{}) {
	udpRoute, ok := obj.(*gwv1alpha2.UDPRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnUDPRouteAdd")
		c.eventHandler.OnUDPRouteAdd(udpRoute)
	}
}

func (c *UDPRouteController) handleUpdateUDPRoute(oldObj, newObj interface{}) {
	oldUdpRoute, ok := oldObj.(*gwv1alpha2.UDPRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	udpRoute, ok := newObj.(*gwv1alpha2.UDPRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnUDPRouteUpdate")
		c.eventHandler.OnUDPRouteUpdate(oldUdpRoute, udpRoute)
	}
}

func (c *UDPRouteController) handleDeleteUDPRoute(obj interface{}) {
	udpRoute, ok := obj.(*gwv1alpha2.UDPRoute)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if udpRoute, ok = tombstone.Obj.(*gwv1alpha2.UDPRoute); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}
	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnUDPRouteDelete")
		c.eventHandler.OnUDPRouteDelete(udpRoute)
	}
}
