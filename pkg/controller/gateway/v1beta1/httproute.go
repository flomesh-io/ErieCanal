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

type HTTPRouteHandler interface {
	OnHTTPRouteAdd(httpRoute *gwv1alpha2.HTTPRoute)
	OnHTTPRouteUpdate(oldHttpRoute, httpRoute *gwv1alpha2.HTTPRoute)
	OnHTTPRouteDelete(httpRoute *gwv1alpha2.HTTPRoute)
	OnHTTPRouteSynced()
}

type HTTPRouteController struct {
	Informer     cache.SharedIndexInformer
	Store        HTTPRouteStore
	HasSynced    cache.InformerSynced
	Lister       gwlisterv1beta1.HTTPRouteLister
	eventHandler HTTPRouteHandler
}

type HTTPRouteStore struct {
	cache.Store
}

func (l *HTTPRouteStore) ByKey(key string) (*gwv1alpha2.HTTPRoute, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*gwv1alpha2.HTTPRoute), nil
}

func NewHTTPRouteControllerWithEventHandler(httpRouteInformer gwinformerv1beta1.HTTPRouteInformer, resyncPeriod time.Duration, handler HTTPRouteHandler) *HTTPRouteController {
	informer := httpRouteInformer.Informer()

	result := &HTTPRouteController{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    httpRouteInformer.Lister(),
		Store: HTTPRouteStore{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddHTTPRoute,
			UpdateFunc: result.handleUpdateHTTPRoute,
			DeleteFunc: result.handleDeleteHTTPRoute,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *HTTPRouteController) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting HTTPRoute config controller")

	if !cache.WaitForNamedCacheSync("HTTPRoute config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnHTTPRouteSynced()")
		c.eventHandler.OnHTTPRouteSynced()
	}
}

func (c *HTTPRouteController) handleAddHTTPRoute(obj interface{}) {
	httpRoute, ok := obj.(*gwv1alpha2.HTTPRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnHTTPRouteAdd")
		c.eventHandler.OnHTTPRouteAdd(httpRoute)
	}
}

func (c *HTTPRouteController) handleUpdateHTTPRoute(oldObj, newObj interface{}) {
	oldHttpRoute, ok := oldObj.(*gwv1alpha2.HTTPRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	httpRoute, ok := newObj.(*gwv1alpha2.HTTPRoute)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnHTTPRouteUpdate")
		c.eventHandler.OnHTTPRouteUpdate(oldHttpRoute, httpRoute)
	}
}

func (c *HTTPRouteController) handleDeleteHTTPRoute(obj interface{}) {
	httpRoute, ok := obj.(*gwv1alpha2.HTTPRoute)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if httpRoute, ok = tombstone.Obj.(*gwv1alpha2.HTTPRoute); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}
	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnHTTPRouteDelete")
		c.eventHandler.OnHTTPRouteDelete(httpRoute)
	}
}
