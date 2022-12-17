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

type ReferencePolicyHandler interface {
	OnReferencePolicyAdd(referencePolicy *gwv1alpha2.ReferencePolicy)
	OnReferencePolicyUpdate(oldReferencePolicy, referencePolicy *gwv1alpha2.ReferencePolicy)
	OnReferencePolicyDelete(referencePolicy *gwv1alpha2.ReferencePolicy)
	OnReferencePolicySynced()
}

type ReferencePolicyController struct {
	Informer     cache.SharedIndexInformer
	Store        ReferencePolicyStore
	HasSynced    cache.InformerSynced
	Lister       gwlisterv1alpha2.ReferencePolicyLister
	eventHandler ReferencePolicyHandler
}

type ReferencePolicyStore struct {
	cache.Store
}

func (l *ReferencePolicyStore) ByKey(key string) (*gwv1alpha2.ReferencePolicy, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*gwv1alpha2.ReferencePolicy), nil
}

func NewReferencePolicyControllerWithEventHandler(referencePolicyInformer gwinformerv1alpha2.ReferencePolicyInformer, resyncPeriod time.Duration, handler ReferencePolicyHandler) *ReferencePolicyController {
	informer := referencePolicyInformer.Informer()

	result := &ReferencePolicyController{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    referencePolicyInformer.Lister(),
		Store: ReferencePolicyStore{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddReferencePolicy,
			UpdateFunc: result.handleUpdateReferencePolicy,
			DeleteFunc: result.handleDeleteReferencePolicy,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *ReferencePolicyController) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting ReferencePolicy config controller")

	if !cache.WaitForNamedCacheSync("ReferencePolicy config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnReferencePolicySynced()")
		c.eventHandler.OnReferencePolicySynced()
	}
}

func (c *ReferencePolicyController) handleAddReferencePolicy(obj interface{}) {
	referencePolicy, ok := obj.(*gwv1alpha2.ReferencePolicy)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnReferencePolicyAdd")
		c.eventHandler.OnReferencePolicyAdd(referencePolicy)
	}
}

func (c *ReferencePolicyController) handleUpdateReferencePolicy(oldObj, newObj interface{}) {
	oldReferencePolicy, ok := oldObj.(*gwv1alpha2.ReferencePolicy)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	referencePolicy, ok := newObj.(*gwv1alpha2.ReferencePolicy)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnReferencePolicyUpdate")
		c.eventHandler.OnReferencePolicyUpdate(oldReferencePolicy, referencePolicy)
	}
}

func (c *ReferencePolicyController) handleDeleteReferencePolicy(obj interface{}) {
	referencePolicy, ok := obj.(*gwv1alpha2.ReferencePolicy)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if referencePolicy, ok = tombstone.Obj.(*gwv1alpha2.ReferencePolicy); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}
	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnReferencePolicyDelete")
		c.eventHandler.OnReferencePolicyDelete(referencePolicy)
	}
}
