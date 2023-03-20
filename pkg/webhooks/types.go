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
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	mcsevent "github.com/flomesh-io/ErieCanal/pkg/mcs/event"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type WebhookObject interface {
	RuntimeObject() runtime.Object
}

type Defaulter interface {
	WebhookObject
	SetDefaults(obj interface{})
}

type Validator interface {
	WebhookObject
	ValidateCreate(obj interface{}) error
	ValidateUpdate(oldObj, obj interface{}) error
	ValidateDelete(obj interface{}) error
}

type Register interface {
	GetWebhooks() ([]admissionregv1.MutatingWebhook, []admissionregv1.ValidatingWebhook)
	GetHandlers() map[string]http.Handler
}

type RegisterConfig struct {
	Manager            manager.Manager
	ConfigStore        *config.Store
	K8sAPI             *kube.K8sAPI
	CertificateManager certificate.Manager
	RepoClient         *repo.PipyRepoClient
	Broker             *mcsevent.Broker
	WebhookSvcNs       string
	WebhookSvcName     string
	CaBundle           []byte
}
