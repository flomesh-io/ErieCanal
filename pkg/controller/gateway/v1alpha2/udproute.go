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
