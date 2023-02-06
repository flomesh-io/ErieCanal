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
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
)

func (c *LocalCache) OnIngressv1Add(ingress *networkingv1.Ingress) {
	c.onIngressUpdate(nil, ingress)
}

func (c *LocalCache) OnIngressv1Update(oldIngress, ingress *networkingv1.Ingress) {
	c.onIngressUpdate(oldIngress, ingress)
}

func (c *LocalCache) OnIngressv1Delete(ingress *networkingv1.Ingress) {
	c.onIngressUpdate(ingress, nil)
}

func (c *LocalCache) OnIngressv1Synced() {
	c.mu.Lock()
	c.ingressesSynced = true
	c.setInitialized(c.servicesSynced && c.endpointsSynced && c.serviceImportSynced && c.ingressClassesSynced)
	c.mu.Unlock()

	c.syncRoutes()
}

func (c *LocalCache) onIngressUpdate(oldIngress, ingress *networkingv1.Ingress) {
	// ONLY update ingress after IngressClass, svc & ep are synced
	if c.ingressChanges.Update(oldIngress, ingress) && c.isInitialized() {
		klog.V(5).Infof("Detects ingress change, syncing...")
		c.Sync()
	}
}
