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
	"context"
	"github.com/flomesh-io/ErieCanal/pkg/cache/controller"
	conn "github.com/flomesh-io/ErieCanal/pkg/cluster/context"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/event"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"k8s.io/client-go/tools/events"
	"time"
)

type Cache interface {
	Sync()
	SyncLoop(stopCh <-chan struct{})
	GetBroadcaster() events.EventBroadcaster
	GetControllers() controller.Controllers
	GetRecorder() events.EventRecorder
}

func NewCache(ctx context.Context, api *kube.K8sAPI, clusterCfg *config.Store, broker *event.Broker, resyncPeriod time.Duration) Cache {
	connectorCtx := ctx.(*conn.ConnectorContext)

	if connectorCtx.ConnectorConfig.IsInCluster() {
		return newLocalCache(ctx, api, clusterCfg, broker, resyncPeriod)
	} else {
		return newRemoteCache(ctx, api, clusterCfg, broker, resyncPeriod)
	}
}
