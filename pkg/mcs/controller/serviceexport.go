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
	svcexpv1alpha1 "github.com/flomesh-io/ErieCanal/apis/serviceexport/v1alpha1"
	svcexpv1alpha1informers "github.com/flomesh-io/ErieCanal/pkg/generated/informers/externalversions/serviceexport/v1alpha1"
	svcexpv1alpha1lister "github.com/flomesh-io/ErieCanal/pkg/generated/listers/serviceexport/v1alpha1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"time"
)

type ServiceExportHandler interface {
	OnServiceExportAdd(serviceExport *svcexpv1alpha1.ServiceExport)
	OnServiceExportUpdate(oldServiceExport, serviceExport *svcexpv1alpha1.ServiceExport)
	OnServiceExportDelete(serviceExport *svcexpv1alpha1.ServiceExport)
	OnServiceExportSynced()
}

type ServiceExportController struct {
	Informer     cache.SharedIndexInformer
	Store        ServiceExportStore
	HasSynced    cache.InformerSynced
	Lister       svcexpv1alpha1lister.ServiceExportLister
	eventHandler ServiceExportHandler
}

type ServiceExportStore struct {
	cache.Store
}

func (l *ServiceExportStore) ByKey(key string) (*svcexpv1alpha1.ServiceExport, error) {
	s, exists, err := l.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no object matching key %q in local store", key)
	}
	return s.(*svcexpv1alpha1.ServiceExport), nil
}

func NewServiceExportControllerWithEventHandler(serviceExportInformer svcexpv1alpha1informers.ServiceExportInformer, resyncPeriod time.Duration, handler ServiceExportHandler) *ServiceExportController {
	informer := serviceExportInformer.Informer()

	result := &ServiceExportController{
		HasSynced: informer.HasSynced,
		Informer:  informer,
		Lister:    serviceExportInformer.Lister(),
		Store: ServiceExportStore{
			Store: informer.GetStore(),
		},
	}

	informer.AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddServiceExport,
			UpdateFunc: result.handleUpdateServiceExport,
			DeleteFunc: result.handleDeleteServiceExport,
		},
		resyncPeriod,
	)

	if handler != nil {
		result.eventHandler = handler
	}

	return result
}

func (c *ServiceExportController) Run(stopCh <-chan struct{}) {
	klog.InfoS("Starting ServiceExport config controller")

	if !cache.WaitForNamedCacheSync("ServiceExport config", stopCh, c.HasSynced) {
		return
	}

	if c.eventHandler != nil {
		klog.V(3).Info("Calling handler.OnServiceExportSynced()")
		c.eventHandler.OnServiceExportSynced()
	}
}

func (c *ServiceExportController) handleAddServiceExport(obj interface{}) {
	export, ok := obj.(*svcexpv1alpha1.ServiceExport)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnServiceExportAdd")
		c.eventHandler.OnServiceExportAdd(export)
	}
}

func (c *ServiceExportController) handleUpdateServiceExport(oldObj, newObj interface{}) {
	oldExport, ok := oldObj.(*svcexpv1alpha1.ServiceExport)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	export, ok := newObj.(*svcexpv1alpha1.ServiceExport)
	if !ok {
		runtime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnServiceExportUpdate")
		c.eventHandler.OnServiceExportUpdate(oldExport, export)
	}
}

func (c *ServiceExportController) handleDeleteServiceExport(obj interface{}) {
	export, ok := obj.(*svcexpv1alpha1.ServiceExport)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
		if export, ok = tombstone.Obj.(*svcexpv1alpha1.ServiceExport); !ok {
			runtime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
			return
		}
	}

	if c.eventHandler != nil {
		klog.V(4).Info("Calling handler.OnServiceExportDelete")
		c.eventHandler.OnServiceExportDelete(export)
	}
}
