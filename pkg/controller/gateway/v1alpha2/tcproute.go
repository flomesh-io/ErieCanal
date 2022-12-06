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
