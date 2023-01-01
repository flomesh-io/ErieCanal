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

//type RouteBase struct {
//	// Region,
//	Region string `json:"region"`
//	// Zone,
//	Zone string `json:"zone"`
//	// Group,
//	Group string `json:"group"`
//	// Cluster,
//	Cluster string `json:"cluster"`
//
//	GatewayHost string `json:"gatewayHost"`
//	GatewayIP   net.IP `json:"gatewayIP"`
//	GatewayPort int32  `json:"gatewayPort"`
//}

type IngressData struct {
	//RouteBase `json:",inline"`
	// Hash
	Hash string `json:"hash" hash:"ignore"`
	// Routes
	Routes []IngressRouteSpec `json:"routes" hash:"set"`
}

type IngressRouteSpec struct {
	RouterSpec   `json:",inline"`
	BalancerSpec `json:",inline"`
	TLSSpec      `json:",inline"`
}

type RouterSpec struct {
	Host    string   `json:"-"`
	Path    string   `json:"-"`
	Service string   `json:"service,omitempty"`
	Rewrite []string `json:"rewrite,omitempty"`
}

type BalancerSpec struct {
	Sticky   bool          `json:"sticky,omitempty"`
	Balancer AlgoBalancer  `json:"balancer,omitempty"`
	Upstream *UpstreamSpec `json:"upstream,omitempty"`
}

type UpstreamSpec struct {
	SSLName   string             `json:"sslName,omitempty"`
	SSLCert   *CertificateSpec   `json:"sslCert,omitempty"`
	SSLVerify bool               `json:"sslVerify,omitempty"`
	Endpoints []UpstreamEndpoint `json:"endpoints,omitempty" hash:"set"`
}

type TLSSpec struct {
	IsTLS          bool             `json:"isTLS,omitempty"`
	IsWildcardHost bool             `json:"isWildcardHost,omitempty"`
	VerifyClient   bool             `json:"verifyClient,omitempty"`
	VerifyDepth    int              `json:"verifyDepth,omitempty"`
	Certificate    *CertificateSpec `json:"certificate,omitempty"`
	TrustedCA      *CertificateSpec `json:"trustedCA,omitempty"`
}

type CertificateSpec struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
	CA   string `json:"ca,omitempty"`
}

type UpstreamEndpoint struct {
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
	//RouteBase `json:",inline"`
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
}

type Target struct {
	// Address can be IP address if in the same cluster, or ingress address for out cluster route
	Address string `json:"address"`
	// Tag, reserved placeholder for futher features
	Tags map[string]string `json:"tags,omitempty" hash:"set"`
}

type IngressConfig struct {
	TrustedCAs     []string `json:"trustedCAs"`
	TLSConfig      `json:",inline"`
	RouterConfig   `json:",inline"`
	BalancerConfig `json:",inline"`
}

type TLSConfig struct {
	Certificates map[string]TLSSpec `json:"certificates"`
}

type RouterConfig struct {
	Routes map[string]RouterSpec `json:"routes"`
}

type BalancerConfig struct {
	Services map[string]BalancerSpec `json:"services"`
}

type AlgoBalancer string

const (
	RoundRobinLoadBalancer AlgoBalancer = "round-robin"
	HashingLoadBalancer    AlgoBalancer = "hashing"
	LeastWorkLoadBalancer  AlgoBalancer = "least-work"
)
