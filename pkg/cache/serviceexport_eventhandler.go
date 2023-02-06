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
	"context"
	"fmt"
	svcexpv1alpha1 "github.com/flomesh-io/ErieCanal/apis/serviceexport/v1alpha1"
	"github.com/flomesh-io/ErieCanal/pkg/event"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *RemoteCache) OnServiceExportAdd(export *svcexpv1alpha1.ServiceExport) {
	klog.V(5).Infof("[%s] OnServiceExportAdd: %#v", c.connectorConfig.Key(), export)

	c.OnUpdate(nil, export)
}

func (c *RemoteCache) OnServiceExportUpdate(oldExport, export *svcexpv1alpha1.ServiceExport) {
	klog.V(5).Infof("[%s] OnServiceExportUpdate: %#v", c.connectorConfig.Key(), export)

	if oldExport.ResourceVersion == export.ResourceVersion {
		klog.Warningf("[%s] OnServiceExportUpdate %s is ignored as ResourceVersion doesn't change", client.ObjectKeyFromObject(export), c.connectorConfig.Key())
		return
	}

	c.OnUpdate(oldExport, export)
}

func (c *RemoteCache) OnUpdate(oldExport, export *svcexpv1alpha1.ServiceExport) {
	mc := c.clusterCfg.MeshConfig.GetConfig()
	if !mc.IsManaged {
		klog.Warningf("[%s] Cluster is not managed, ignore processing ServiceExport %s", c.connectorConfig.Key(), client.ObjectKeyFromObject(export))
		return
	}

	svc, err := c.getService(export)
	if err != nil {
		klog.Errorf("[%s] Ignore processing ServiceExport %s", c.connectorConfig.Key(), client.ObjectKeyFromObject(export))
		return
	}

	c.broker.Enqueue(
		event.Message{
			Kind:   event.ServiceExportCreated,
			OldObj: nil,
			NewObj: &event.ServiceExportEvent{
				Geo:           c.connectorConfig,
				ServiceExport: export,
				Service:       svc,
			},
		},
	)
}

func (c *RemoteCache) OnServiceExportDelete(export *svcexpv1alpha1.ServiceExport) {
	klog.V(5).Infof("[%s] OnServiceExportDelete: %#v", c.connectorConfig.Key(), export)

	mc := c.clusterCfg.MeshConfig.GetConfig()
	if !mc.IsManaged {
		klog.Warningf("[%s] Cluster is not managed, ignore processing ServiceExport %s", c.connectorConfig.Key(), client.ObjectKeyFromObject(export))
		return
	}

	svc, err := c.getService(export)
	if err != nil {
		klog.Errorf("[%s] Ignore processing ServiceExport %s due to get service failed", c.connectorConfig.Key(), client.ObjectKeyFromObject(export))
		return
	}

	c.broker.Enqueue(
		event.Message{
			Kind:   event.ServiceExportDeleted,
			NewObj: nil,
			OldObj: &event.ServiceExportEvent{
				Geo:           c.connectorConfig,
				ServiceExport: export,
				Service:       svc,
				//Data:          make(map[string]interface{}),
			},
		},
	)
}

func (c *RemoteCache) OnServiceExportSynced() {
	c.mu.Lock()
	c.serviceExportSynced = true
	c.setInitialized(c.serviceExportSynced)
	c.mu.Unlock()

	c.syncManagedCluster()
}

func (c *RemoteCache) getService(export *svcexpv1alpha1.ServiceExport) (*corev1.Service, error) {
	klog.V(5).Infof("[%s] Getting service %s/%s", c.connectorConfig.Key(), export.Namespace, export.Name)

	svc, err := c.k8sAPI.Client.CoreV1().
		Services(export.Namespace).
		Get(context.TODO(), export.Name, metav1.GetOptions{})

	if err != nil {
		klog.Errorf("[%s] Failed to get svc %s/%s, %s", c.connectorConfig.Key(), export.Namespace, export.Name, err)
		return nil, err
	}

	if svc.Spec.Type == corev1.ServiceTypeExternalName {
		msg := fmt.Sprintf("[%s] ExternalName service %s/%s cannot be exported", c.connectorConfig.Key(), export.Namespace, export.Name)
		klog.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}

	return svc, nil
}
