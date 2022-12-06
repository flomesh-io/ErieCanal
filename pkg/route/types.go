/*
 * MIT License
 *
 * Copyright (c) since 2021,  flomesh.io Authors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
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
