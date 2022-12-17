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
	"github.com/flomesh-io/ErieCanal/apis/serviceimport/v1alpha1"
	"k8s.io/klog/v2"
)

func (c *LocalCache) OnServiceImportAdd(serviceImport *v1alpha1.ServiceImport) {
	c.OnServiceImportUpdate(nil, serviceImport)
}

func (c *LocalCache) OnServiceImportUpdate(oldServiceImport, serviceImport *v1alpha1.ServiceImport) {
	if c.serviceImportChanges.Update(oldServiceImport, serviceImport) && c.isInitialized() {
		klog.V(5).Infof("Detects ServiceImport change, syncing...")
		c.Sync()
	}
}

func (c *LocalCache) OnServiceImportDelete(serviceImport *v1alpha1.ServiceImport) {
	c.OnServiceImportUpdate(serviceImport, nil)
}

func (c *LocalCache) OnServiceImportSynced() {
	c.mu.Lock()
	c.serviceImportSynced = true
	c.setInitialized(c.servicesSynced && c.endpointsSynced && c.ingressesSynced && c.ingressClassesSynced)
	c.mu.Unlock()

	c.syncRoutes()
}
