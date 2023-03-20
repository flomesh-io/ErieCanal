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

package controllers

import (
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	mcsevent "github.com/flomesh-io/ErieCanal/pkg/mcs/event"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type ReconcilerConfig struct {
	Manager            manager.Manager
	ConfigStore        *config.Store
	K8sAPI             *kube.K8sAPI
	CertificateManager certificate.Manager
	RepoClient         *repo.PipyRepoClient
	Broker             *mcsevent.Broker
	client.Client
	Scheme *runtime.Scheme
}

type Reconciler interface {
	SetupWithManager(mgr ctrl.Manager) error
}
