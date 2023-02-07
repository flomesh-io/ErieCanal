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

package webhooks

import (
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/cluster"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/cm"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/gateway"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/gatewayclass"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/globaltrafficpolicy"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/httproute"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/ingress"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/namespacedingress"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/serviceexport"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/serviceimport"
)

func RegisterWebhooks(webhookSvcNs, webhookSvcName string, caBundle []byte) {
	cluster.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	cm.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	namespacedingress.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	serviceexport.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	serviceimport.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	globaltrafficpolicy.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	ingress.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
}

func RegisterGatewayApiWebhooks(webhookSvcNs, webhookSvcName string, caBundle []byte) {
	gateway.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	gatewayclass.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	//referencepolicy.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	httproute.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	//tcproute.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	//tlsroute.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
	//udproute.RegisterWebhooks(webhookSvcNs, webhookSvcName, caBundle)
}
