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

package event

import (
	svcexpv1alpha1 "github.com/flomesh-io/ErieCanal/apis/serviceexport/v1alpha1"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	corev1 "k8s.io/api/core/v1"
)

type EventType string

const (
	ServiceExportCreated  EventType = "service.export.created"
	ServiceExportDeleted  EventType = "service.export.deleted"
	ServiceExportAccepted EventType = "service.export.accepted"
	ServiceExportRejected EventType = "service.export.rejected"
)

type Message struct {
	Kind   EventType
	OldObj interface{}
	NewObj interface{}
}

//type GeoInfo struct {
//	Region  string
//	Zone    string
//	Group   string
//	Cluster string
//}

type ServiceExportEvent struct {
	Geo           *config.ConnectorConfig
	ServiceExport *svcexpv1alpha1.ServiceExport
	Service       *corev1.Service
	Error         string
	//Data          map[string]interface{}
}

func (e *ServiceExportEvent) ClusterKey() string {
	return e.Geo.Key()
}

//func NewServiceExportMessage(eventType EventType, geo *config.ConnectorConfig, serviceExport *svcexpv1alpha1.ServiceExport, svc *corev1.Service, data map[string]interface{}) *Message {
//	obj := ServiceExportEvent{Geo: geo, ServiceExport: serviceExport, Service: svc, Data: data}
//
//	switch eventType {
//	case ServiceExportAccepted, ServiceExportCreated, ServiceExportRejected:
//		return &Message{
//			Kind:   eventType,
//			OldObj: nil,
//			NewObj: obj,
//		}
//	case ServiceExportDeleted:
//		return &Message{
//			Kind:   eventType,
//			OldObj: obj,
//			NewObj: nil,
//		}
//	}
//
//	return nil
//}
