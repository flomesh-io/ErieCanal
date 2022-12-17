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
	"github.com/flomesh-io/ErieCanal/pkg/cache/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"
	utilcache "k8s.io/kubernetes/pkg/proxy/util"
	utilnet "k8s.io/utils/net"
	"net"
	"reflect"
	"strconv"
	"sync"
)

type BaseEndpointInfo struct {
	Endpoint string
	Nodename string
	Hostname string
	Cluster  string
}

var _ Endpoint = &BaseEndpointInfo{}

func (info *BaseEndpointInfo) String() string {
	return info.Endpoint
}

func (info *BaseEndpointInfo) IP() string {
	return utilcache.IPPart(info.Endpoint)
}

func (info *BaseEndpointInfo) Port() (int, error) {
	return utilcache.PortPart(info.Endpoint)
}

func (info *BaseEndpointInfo) NodeName() string {
	return info.Nodename
}

func (info *BaseEndpointInfo) HostName() string {
	return info.Hostname
}

func (info *BaseEndpointInfo) ClusterInfo() string {
	return info.Cluster
}

func (info *BaseEndpointInfo) Equal(other Endpoint) bool {
	return info.String() == other.String()
}

func newBaseEndpointInfo(IP string, port int, nodename string, hostname string) *BaseEndpointInfo {
	return &BaseEndpointInfo{
		Endpoint: net.JoinHostPort(IP, strconv.Itoa(port)),
		Nodename: nodename,
		Hostname: hostname,
	}
}

type enrichEndpointFunc func(info *BaseEndpointInfo) Endpoint

type EndpointChangeTracker struct {
	lock               sync.Mutex
	items              map[types.NamespacedName]*endpointsChange
	enrichEndpointInfo enrichEndpointFunc
	recorder           events.EventRecorder
	controllers        *controller.LocalControllers
}

func NewEndpointChangeTracker(enrichEndpointInfo enrichEndpointFunc, recorder events.EventRecorder, controllers *controller.LocalControllers) *EndpointChangeTracker {
	return &EndpointChangeTracker{
		items:              make(map[types.NamespacedName]*endpointsChange),
		enrichEndpointInfo: enrichEndpointInfo,
		recorder:           recorder,
		controllers:        controllers,
	}
}

func (ect *EndpointChangeTracker) Update(previous, current *corev1.Endpoints) bool {
	endpoints := current
	if endpoints == nil {
		endpoints = previous
	}
	if endpoints == nil {
		return false
	}

	namespacedName := types.NamespacedName{Namespace: endpoints.Namespace, Name: endpoints.Name}

	ect.lock.Lock()
	defer ect.lock.Unlock()

	change, exists := ect.items[namespacedName]
	if !exists {
		change = &endpointsChange{}
		change.previous = ect.endpointsToEndpointsMap(previous)
		ect.items[namespacedName] = change
	}

	change.current = ect.endpointsToEndpointsMap(current)

	if reflect.DeepEqual(change.previous, change.current) {
		delete(ect.items, namespacedName)
	} else {
		for spn, eps := range change.current {
			klog.V(2).Infof("Service port %s updated: %d endpoints", spn, len(eps))
		}
	}

	return len(ect.items) > 0
}

func (ect *EndpointChangeTracker) checkoutChanges() []*endpointsChange {
	ect.lock.Lock()
	defer ect.lock.Unlock()

	changes := []*endpointsChange{}
	for _, change := range ect.items {
		changes = append(changes, change)
	}
	ect.items = make(map[types.NamespacedName]*endpointsChange)
	return changes
}

type endpointsChange struct {
	previous EndpointsMap
	current  EndpointsMap
}

func (em EndpointsMap) Update(changes *EndpointChangeTracker) {
	em.apply(changes)
}

type EndpointsMap map[ServicePortName][]Endpoint

func (ect *EndpointChangeTracker) endpointsToEndpointsMap(endpoints *corev1.Endpoints) EndpointsMap {
	if endpoints == nil {
		return nil
	}

	endpointsMap := make(EndpointsMap)
	for i := range endpoints.Subsets {
		ss := &endpoints.Subsets[i]
		for i := range ss.Ports {
			port := &ss.Ports[i]
			if port.Port == 0 {
				klog.Warningf("ignoring invalid endpoint port %s", port.Name)
				continue
			}
			svcPortName := ServicePortName{
				NamespacedName: types.NamespacedName{Namespace: endpoints.Namespace, Name: endpoints.Name},
				Port:           port.Name,
				Protocol:       port.Protocol,
			}
			for i := range ss.Addresses {
				addr := &ss.Addresses[i]
				if addr.IP == "" {
					klog.Warningf("ignoring invalid endpoint port %s with empty host", port.Name)
					continue
				}

				// ONLY supports IPv4
				if !utilnet.IsIPv4String(addr.IP) {
					continue
				}

				klog.V(5).Infof("Address = %#v", addr)

				baseEndpointInfo := newBaseEndpointInfo(addr.IP, int(port.Port), nodename(addr), addr.Hostname)
				if ect.enrichEndpointInfo != nil {
					endpointsMap[svcPortName] = append(endpointsMap[svcPortName], ect.enrichEndpointInfo(baseEndpointInfo))
				} else {
					endpointsMap[svcPortName] = append(endpointsMap[svcPortName], baseEndpointInfo)
				}
			}

			klog.V(3).Infof("Setting endpoints for %q to %#v", svcPortName, formatEndpointsList(endpointsMap[svcPortName]))
		}
	}

	return endpointsMap
}

func nodename(addr *corev1.EndpointAddress) string {
	if addr == nil {
		return ""
	}

	nodeName := ""
	if addr.NodeName != nil {
		nodeName = *addr.NodeName
	}

	return nodeName
}

func formatEndpointsList(endpoints []Endpoint) []string {
	var formattedList []string
	for _, ep := range endpoints {
		formattedList = append(formattedList, ep.String())
	}
	return formattedList
}

func (em EndpointsMap) apply(ect *EndpointChangeTracker) {
	if ect == nil {
		return
	}

	changes := ect.checkoutChanges()
	for _, change := range changes {
		em.unmerge(change.previous)
		em.merge(change.current)
	}
}

func (em EndpointsMap) merge(other EndpointsMap) {
	for svcPortName := range other {
		em[svcPortName] = other[svcPortName]
	}
}

func (em EndpointsMap) unmerge(other EndpointsMap) {
	for svcPortName := range other {
		delete(em, svcPortName)
	}
}
