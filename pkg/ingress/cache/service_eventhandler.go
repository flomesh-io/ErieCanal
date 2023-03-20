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

func (c *Cache) OnServiceAdd(service *corev1.Service) {
	c.OnServiceUpdate(nil, service)
}

func (c *Cache) OnServiceUpdate(oldService, service *corev1.Service) {
	if c.serviceChanges.Update(oldService, service) && c.isInitialized() {
		klog.V(5).Infof("Detects service change, syncing...")
		c.Sync()
	}
}

func (c *Cache) OnServiceDelete(service *corev1.Service) {
	c.OnServiceUpdate(service, nil)
}

func (c *Cache) OnServiceSynced() {
	c.mu.Lock()
	c.servicesSynced = true
	c.setInitialized(c.serviceImportSynced && c.endpointsSynced && c.ingressesSynced && c.ingressClassesSynced)
	c.mu.Unlock()

	c.syncRoutes()
}
