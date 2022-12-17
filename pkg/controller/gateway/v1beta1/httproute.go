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
