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

package route

import (
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"net"
)

type RouteBase struct {
	// Region,
	Region string `json:"region"`
	// Zone,
	Zone string `json:"zone"`
	// Group,
	Group string `json:"group"`
	// Cluster,
	Cluster string `json:"cluster"`

	GatewayHost string `json:"gatewayHost"`
	GatewayIP   net.IP `json:"gatewayIP"`
	GatewayPort int32  `json:"gatewayPort"`
}
type IngressRoute struct {
	RouteBase `json:",inline"`
	// Hash
	Hash string `json:"hash" hash:"ignore"`
	// Routes
	Routes []IngressRouteEntry `json:"routes" hash:"set"`
}

type IngressRouteEntry struct {
	Host        string            `json:"host,omitempty"`
	Path        string            `json:"path,omitempty"`
	ServiceName string            `json:"serviceName,omitempty"`
	Rewrite     []string          `json:"rewrite,omitempty"`
	Sticky      bool              `json:"sticky,omitempty"`
	Balancer    repo.AlgoBalancer `json:"balancer,omitempty"`
	Upstreams   []EndpointEntry   `json:"upstreams,omitempty" hash:"set"`
}

type EndpointEntry struct {
	// IP is the entry's IP.  The IP address protocol corresponds to the HashFamily of IPSet.
	// All entries' IP addresses in the same ip set has same the protocol, IPv4 or IPv6.
	IP string `json:"ip,omitempty"`
	// Port is the entry's Port.
	Port int `json:"port,omitempty"`
	// Protocol is the entry's Protocol.  The protocols of entries in the same ip set are all
	// the same.  The accepted protocols are TCP, UDP and SCTP.
	Protocol string `json:"protocol,omitempty"`
}

type ServiceRoute struct {
	RouteBase `json:",inline"`
	// Hash
	Hash   string              `json:"hash" hash:"ignore"`
	Routes []ServiceRouteEntry `json:"routes" hash:"set"`
}

type ServiceRouteEntry struct {
	// Name, the name of the service
	Name string `json:"name"`
	// Namespace, the namespace of the service, it has value no matter in cluster/out cluster, but will only be used for in-cluster
	Namespace string `json:"namespace"`
	// Targets
	Targets []Target `json:"targets" hash:"set"`
	// PortName
	PortName string `json:"portName,omitempty"`
	//// ExternalPath, it's for out-cluster access, combined with address, can be empty if it's not exposed by ingress
	//ExternalPath string `json:"externalPath,omitempty"`
	//// Export
	//Export bool `json:"export,omitempty"`
	//// ExportName
	//ExportName string `json:"exportName,omitempty"`
}

type Target struct {
	// Address can be IP address if in the same cluster, or ingress address for out cluster route
	Address string `json:"address"`
	// Tag, reserved placeholder for futher features
	Tags map[string]string `json:"tags,omitempty" hash:"set"`
}
