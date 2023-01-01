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
	"fmt"
	gwcontrollerv1alpha2 "github.com/flomesh-io/ErieCanal/pkg/controller/gateway/v1alpha2"
	gwcontrollerv1beta1 "github.com/flomesh-io/ErieCanal/pkg/controller/gateway/v1beta1"
	"github.com/flomesh-io/ErieCanal/pkg/route"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Route , Ingress Route interface
type Route interface {
	String() string
	Headers() map[string]string
	Host() string
	Path() string
	Backend() ServicePortName
	Rewrite() []string
	SessionSticky() bool
	LBType() route.AlgoBalancer
	UpstreamSSLName() string
	UpstreamSSLCert() *route.CertificateSpec
	UpstreamSSLVerify() bool
	Certificate() *route.CertificateSpec
	IsTLS() bool
	IsWildcardHost() bool
	VerifyClient() bool
	VerifyDepth() int
	TrustedCA() *route.CertificateSpec
}

type ServicePortName struct {
	types.NamespacedName
	Port     string
	Protocol v1.Protocol
}

func (spn ServicePortName) String() string {
	return fmt.Sprintf("%s%s", spn.NamespacedName.String(), fmtPortName(spn.Port))
}

func fmtPortName(in string) string {
	if in == "" {
		return ""
	}
	return fmt.Sprintf(":%s", in)
}

type ServicePort interface {
	String() string
	Address() string
	Port() int
	Protocol() v1.Protocol
}

type Endpoint interface {
	String() string
	IP() string
	Port() (int, error)
	NodeName() string
	HostName() string
	ClusterInfo() string
	Equal(Endpoint) bool
}

type ServiceEndpoint struct {
	Endpoint        string
	ServicePortName ServicePortName
}

type GatewayApiControllers struct {
	V1beta1  *GatewayApiV1beta1Controllers
	V1alpha2 *GatewayApiV1alpha2Controllers
}

type GatewayApiV1beta1Controllers struct {
	Gateway      *gwcontrollerv1beta1.GatewayController
	GatewayClass *gwcontrollerv1beta1.GatewayClassController
	HTTPRoute    *gwcontrollerv1beta1.HTTPRouteController
}

type GatewayApiV1alpha2Controllers struct {
	ReferencePolicy *gwcontrollerv1alpha2.ReferencePolicyController
	TCPRoute        *gwcontrollerv1alpha2.TCPRouteController
	TLSRoute        *gwcontrollerv1alpha2.TLSRouteController
	UDPRoute        *gwcontrollerv1alpha2.UDPRouteController
}
