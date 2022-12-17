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
