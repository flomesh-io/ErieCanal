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
