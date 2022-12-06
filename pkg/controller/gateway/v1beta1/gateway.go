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
