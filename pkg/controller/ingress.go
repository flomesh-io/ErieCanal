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

type Ingressv1Handler interface {
	OnIngressv1Add(ingress *networkingv1.Ingress)
	OnIngressv1Update(oldIngress, ingress *networkingv1.Ingress)
	OnIngressv1Delete(ingress *networkingv1.Ingress)
	OnIngressv1Synced()
}

type Ingressv1Controller struct {
	Informer     cache.SharedIndexInformer
	Store        Ingressv1Store
	HasSynced    cache.InformerSynced
	Lister       networkingv1lister.IngressLister
	eventHandler Ingressv1Handler
}

type Ingressv1Store struct {
	cache.Store
}

func (l *Ingressv1Store) ByKey(key string) (*networkingv1.Ingress, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*networkingv1.Ingress), nil
}

func NewIngressv1ControllerWithEventHandler(ingressInformer networkingv1informers.IngressInformer, resyncPeriod time.Duration, handler Ingressv1Handler) *Ingressv1Controller {
	informer := ingressInformer.Informer()

	result := &Ingressv1Controller{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    ingressInformer.Lister(),
		Store: Ingressv1Store{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddIngress,
			UpdateFunc: result.handleUpdateIngress,
			DeleteFunc: result.handleDeleteIngress,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *Ingressv1Controller) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting ingress config controller")

	if !cache.WaitForNamedCacheSync("ingress v1 config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnIngressv1Synced()")
		c.eventHandler.OnIngressv1Synced()
	}
}

func (c *Ingressv1Controller) handleAddIngress(obj interface{}) {
	ing, ok := obj.(*networkingv1.Ingress)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnIngressv1Add")
		c.eventHandler.OnIngressv1Add(ing)
	}
}

func (c *Ingressv1Controller) handleUpdateIngress(oldObj, newObj interface{}) {
	oldIng, ok := oldObj.(*networkingv1.Ingress)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	ing, ok := newObj.(*networkingv1.Ingress)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnIngressv1Update")
		c.eventHandler.OnIngressv1Update(oldIng, ing)
	}
}

func (c *Ingressv1Controller) handleDeleteIngress(obj interface{}) {
	ing, ok := obj.(*networkingv1.Ingress)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if ing, ok = tombstone.Obj.(*networkingv1.Ingress); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnIngressv1Delete")
		c.eventHandler.OnIngressv1Delete(ing)
	}
}
