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

type TCPRouteHandler interface {
	OnTCPRouteAdd(tcpRoute *gwv1alpha2.TCPRoute)
	OnTCPRouteUpdate(oldTcpRoute, tcpRoute *gwv1alpha2.TCPRoute)
	OnTCPRouteDelete(tcpRoute *gwv1alpha2.TCPRoute)
	OnTCPRouteSynced()
}

type TCPRouteController struct {
	Informer     cache.SharedIndexInformer
	Store        TCPRouteStore
	HasSynced    cache.InformerSynced
	Lister       gwlisterv1alpha2.TCPRouteLister
	eventHandler TCPRouteHandler
}

type TCPRouteStore struct {
	cache.Store
}

func (l *TCPRouteStore) ByKey(key string) (*gwv1alpha2.TCPRoute, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*gwv1alpha2.TCPRoute), nil
}

func NewTCPRouteControllerWithEventHandler(tcpRouteInformer gwinformerv1alpha2.TCPRouteInformer, resyncPeriod time.Duration, handler TCPRouteHandler) *TCPRouteController {
	informer := tcpRouteInformer.Informer()

	result := &TCPRouteController{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    tcpRouteInformer.Lister(),
		Store: TCPRouteStore{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddTCPRoute,
			UpdateFunc: result.handleUpdateTCPRoute,
			DeleteFunc: result.handleDeleteTCPRoute,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *TCPRouteController) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting TCPRoute config controller")

	if !cache.WaitForNamedCacheSync("TCPRoute config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnTCPRouteSynced()")
		c.eventHandler.OnTCPRouteSynced()
	}
}

func (c *TCPRouteController) handleAddTCPRoute(obj interface{}) {
	tcpRoute, ok := obj.(*gwv1alpha2.TCPRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnTCPRouteAdd")
		c.eventHandler.OnTCPRouteAdd(tcpRoute)
	}
}

func (c *TCPRouteController) handleUpdateTCPRoute(oldObj, newObj interface{}) {
	oldTcpRoute, ok := oldObj.(*gwv1alpha2.TCPRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	tcpRoute, ok := newObj.(*gwv1alpha2.TCPRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnTCPRouteUpdate")
		c.eventHandler.OnTCPRouteUpdate(oldTcpRoute, tcpRoute)
	}
}

func (c *TCPRouteController) handleDeleteTCPRoute(obj interface{}) {
	tcpRoute, ok := obj.(*gwv1alpha2.TCPRoute)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if tcpRoute, ok = tombstone.Obj.(*gwv1alpha2.TCPRoute); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}
	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnTCPRouteDelete")
		c.eventHandler.OnTCPRouteDelete(tcpRoute)
	}
}
