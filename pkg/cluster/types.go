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

package cluster

import (
	"context"
	"github.com/flomesh-io/ErieCanal/pkg/cache"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/event"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
)

type Connector interface {
	Run(stopCh <-chan struct{}) error
}

type LocalConnector struct {
	context    context.Context
	k8sAPI     *kube.K8sAPI
	cache      cache.Cache
	clusterCfg *config.Store
	broker     *event.Broker
}

type RemoteConnector struct {
	context    context.Context
	k8sAPI     *kube.K8sAPI
	cache      cache.Cache
	clusterCfg *config.Store
	broker     *event.Broker
}
