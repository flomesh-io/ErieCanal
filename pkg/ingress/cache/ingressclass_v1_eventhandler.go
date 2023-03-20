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
	ingresspipy "github.com/flomesh-io/ErieCanal/pkg/ingress"
	networkingv1 "k8s.io/api/networking/v1"
)

func (c *Cache) OnIngressClassv1Add(class *networkingv1.IngressClass) {
	c.updateDefaultIngressClass(class, class.Name)
}

func (c *Cache) OnIngressClassv1Update(oldClass, class *networkingv1.IngressClass) {
	if oldClass.ResourceVersion == class.ResourceVersion {
		return
	}

	c.updateDefaultIngressClass(class, class.Name)
}

func (c *Cache) OnIngressClassv1Delete(class *networkingv1.IngressClass) {
	// if the default IngressClass is deleted, set the DefaultIngressClass variable to empty
	c.updateDefaultIngressClass(class, ingresspipy.NoDefaultIngressClass)
}

func (c *Cache) OnIngressClassv1Synced() {
	c.mu.Lock()
	c.ingressClassesSynced = true
	c.setInitialized(c.ingressesSynced && c.servicesSynced && c.endpointsSynced && c.serviceImportSynced)
	c.mu.Unlock()

	c.syncRoutes()
}

func (c *Cache) updateDefaultIngressClass(class *networkingv1.IngressClass, className string) {
	isDefault, ok := class.GetAnnotations()[ingresspipy.IngressClassAnnotationKey]
	if ok && isDefault == "true" {
		ingresspipy.DefaultIngressClass = className
	}
}
