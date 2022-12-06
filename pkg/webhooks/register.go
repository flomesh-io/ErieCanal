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

package webhooks

import (
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/cluster"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/cm"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/gateway"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/gatewayclass"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/globaltrafficpolicy"
	"github.com/flomesh-io/ErieCanal/pkg/webhooks/httproute"
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
