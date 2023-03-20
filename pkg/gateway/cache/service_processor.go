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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type ServicesProcessor struct{}

func (p *ServicesProcessor) Insert(obj interface{}, cache *GatewayCache) bool {
	svc, ok := obj.(*corev1.Service)
	if !ok {

		klog.Errorf("unexpected object type %T", obj)
		return false
	}

	key := objectKey(svc)
	cache.services[key] = true

	return cache.isRoutableService(key)
}

func (p *ServicesProcessor) Delete(obj interface{}, cache *GatewayCache) bool {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		klog.Errorf("unexpected object type %T", obj)
		return false
	}

	key := objectKey(svc)
	_, found := cache.services[key]
	delete(cache.services, key)

	return found
}
