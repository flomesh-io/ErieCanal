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

package cache

import (
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EndpointSlicesProcessor struct {
}

func (p *EndpointSlicesProcessor) Insert(obj interface{}, cache *GatewayCache) bool {
	eps, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		klog.Errorf("unexpected object type %T", obj)
		return false
	}

	owner := metav1.GetControllerOf(eps)
	if owner == nil {
		return false
	}

	svcKey := client.ObjectKey{Namespace: eps.Namespace, Name: owner.Name}
	_, found := cache.endpointslices[svcKey]
	if !found {
		cache.endpointslices[svcKey] = make(map[client.ObjectKey]bool)
	}
	cache.endpointslices[svcKey][objectKey(eps)] = true

	return cache.isRoutableService(svcKey)
}

func (p *EndpointSlicesProcessor) Delete(obj interface{}, cache *GatewayCache) bool {
	eps, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		klog.Errorf("unexpected object type %T", obj)
		return false
	}

	owner := metav1.GetControllerOf(eps)
	if owner == nil {
		return false
	}

	svcKey := client.ObjectKey{Namespace: eps.Namespace, Name: owner.Name}
	slices, found := cache.endpointslices[svcKey]
	if !found {
		return false
	}

	sliceKey := objectKey(eps)
	_, found = slices[sliceKey]
	delete(cache.endpointslices[svcKey], sliceKey)

	if len(cache.endpointslices[svcKey]) == 0 {
		delete(cache.endpointslices, svcKey)
	}

	return found
}
