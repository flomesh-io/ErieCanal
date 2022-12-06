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

type TLSRouteHandler interface {
	OnTLSRouteAdd(tlsRoute *gwv1alpha2.TLSRoute)
	OnTLSRouteUpdate(oldTlsRoute, tlsRoute *gwv1alpha2.TLSRoute)
	OnTLSRouteDelete(tlsRoute *gwv1alpha2.TLSRoute)
	OnTLSRouteSynced()
}

type TLSRouteController struct {
	Informer     cache.SharedIndexInformer
	Store        TLSRouteStore
	HasSynced    cache.InformerSynced
	Lister       gwlisterv1alpha2.TLSRouteLister
	eventHandler TLSRouteHandler
}

type TLSRouteStore struct {
	cache.Store
}

func (l *TLSRouteStore) ByKey(key string) (*gwv1alpha2.TLSRoute, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*gwv1alpha2.TLSRoute), nil
}

func NewTLSRouteControllerWithEventHandler(tlsRouteInformer gwinformerv1alpha2.TLSRouteInformer, resyncPeriod time.Duration, handler TLSRouteHandler) *TLSRouteController {
	informer := tlsRouteInformer.Informer()

	result := &TLSRouteController{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    tlsRouteInformer.Lister(),
		Store: TLSRouteStore{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddTLSRoute,
			UpdateFunc: result.handleUpdateTLSRoute,
			DeleteFunc: result.handleDeleteTLSRoute,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *TLSRouteController) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting TLSRoute config controller")

	if !cache.WaitForNamedCacheSync("TLSRoute config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnTLSRouteSynced()")
		c.eventHandler.OnTLSRouteSynced()
	}
}

func (c *TLSRouteController) handleAddTLSRoute(obj interface{}) {
	tlsRoute, ok := obj.(*gwv1alpha2.TLSRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnTLSRouteAdd")
		c.eventHandler.OnTLSRouteAdd(tlsRoute)
	}
}

func (c *TLSRouteController) handleUpdateTLSRoute(oldObj, newObj interface{}) {
	oldTlsRoute, ok := oldObj.(*gwv1alpha2.TLSRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	tlsRoute, ok := newObj.(*gwv1alpha2.TLSRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnTLSRouteUpdate")
		c.eventHandler.OnTLSRouteUpdate(oldTlsRoute, tlsRoute)
	}
}

func (c *TLSRouteController) handleDeleteTLSRoute(obj interface{}) {
	tlsRoute, ok := obj.(*gwv1alpha2.TLSRoute)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if tlsRoute, ok = tombstone.Obj.(*gwv1alpha2.TLSRoute); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}
	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnTLSRouteDelete")
		c.eventHandler.OnTLSRouteDelete(tlsRoute)
	}
}
