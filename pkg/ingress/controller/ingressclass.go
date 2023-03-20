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

package controller

import (
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	networkingv1informers "k8s.io/client-go/informers/networking/v1"
	networkingv1lister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"time"
)

type IngressClassv1Handler interface {
	OnIngressClassv1Add(ingressClass *networkingv1.IngressClass)
	OnIngressClassv1Update(oldIngressClass, ingressClass *networkingv1.IngressClass)
	OnIngressClassv1Delete(ingressClass *networkingv1.IngressClass)
	OnIngressClassv1Synced()
}

type IngressClassv1Controller struct {
	Informer     cache.SharedIndexInformer
	Store        IngressClassv1Store
	HasSynced    cache.InformerSynced
	Lister       networkingv1lister.IngressClassLister
	eventHandler IngressClassv1Handler
}

type IngressClassv1Store struct {
	cache.Store
}

func (l *IngressClassv1Store) ByKey(key string) (*networkingv1.IngressClass, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*networkingv1.IngressClass), nil
}

func NewIngressClassv1ControllerWithEventHandler(ingressClassInformer networkingv1informers.IngressClassInformer, resyncPeriod time.Duration, handler IngressClassv1Handler) *IngressClassv1Controller {
	informer := ingressClassInformer.Informer()

	result := &IngressClassv1Controller{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    ingressClassInformer.Lister(),
		Store: IngressClassv1Store{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddIngressClass,
			UpdateFunc: result.handleUpdateIngressClass,
			DeleteFunc: result.handleDeleteIngressClass,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *IngressClassv1Controller) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting ingress class config controller")

	if !cache.WaitForNamedCacheSync("ingress class v1 config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnIngressClassv1Synced()")
		c.eventHandler.OnIngressClassv1Synced()
	}
}

func (c *IngressClassv1Controller) handleAddIngressClass(obj interface{}) {
	class, ok := obj.(*networkingv1.IngressClass)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnIngressClassv1Add")
		c.eventHandler.OnIngressClassv1Add(class)
	}
}

func (c *IngressClassv1Controller) handleUpdateIngressClass(oldObj, newObj interface{}) {
	oldClass, ok := oldObj.(*networkingv1.IngressClass)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	class, ok := newObj.(*networkingv1.IngressClass)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnIngressClassv1Update")
		c.eventHandler.OnIngressClassv1Update(oldClass, class)
	}
}

func (c *IngressClassv1Controller) handleDeleteIngressClass(obj interface{}) {
	class, ok := obj.(*networkingv1.IngressClass)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if class, ok = tombstone.Obj.(*networkingv1.IngressClass); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnIngressClassv1Delete")
		c.eventHandler.OnIngressClassv1Delete(class)
	}
}
