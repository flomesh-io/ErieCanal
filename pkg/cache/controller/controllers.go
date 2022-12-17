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
	"github.com/flomesh-io/ErieCanal/pkg/controller"
	gwcontrollerv1alpha2 "github.com/flomesh-io/ErieCanal/pkg/controller/gateway/v1alpha2"
	gwcontrollerv1beta1 "github.com/flomesh-io/ErieCanal/pkg/controller/gateway/v1beta1"
)

type Controllers interface {
}

type LocalControllers struct {
	Service        *controller.ServiceController
	Endpoints      *controller.EndpointsController
	Ingressv1      *controller.Ingressv1Controller
	IngressClassv1 *controller.IngressClassv1Controller
	ServiceImport  *controller.ServiceImportController
	GatewayApi     *GatewayApiControllers
}

var _ Controllers = &LocalControllers{}

type RemoteControllers struct {
	ServiceExport *controller.ServiceExportController
}

var _ Controllers = &RemoteControllers{}

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
