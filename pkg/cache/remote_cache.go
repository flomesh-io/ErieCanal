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
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/cache/controller"
	conn "github.com/flomesh-io/ErieCanal/pkg/cluster/context"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	cachectrl "github.com/flomesh-io/ErieCanal/pkg/controller"
	"github.com/flomesh-io/ErieCanal/pkg/event"
	ecinformers "github.com/flomesh-io/ErieCanal/pkg/generated/informers/externalversions"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/util/async"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type RemoteCache struct {
	connectorConfig *config.ConnectorConfig
	k8sAPI          *kube.K8sAPI
	recorder        events.EventRecorder
	clusterCfg      *config.Store
	broker          *event.Broker

	mu sync.Mutex

	serviceExportSynced bool
	initialized         int32
	syncRunner          *async.BoundedFrequencyRunner

	controllers *controller.RemoteControllers
	broadcaster events.EventBroadcaster
}

func newRemoteCache(ctx context.Context, api *kube.K8sAPI, clusterCfg *config.Store, broker *event.Broker, resyncPeriod time.Duration) *RemoteCache {
	connectorCtx := ctx.(*conn.ConnectorContext)
	key := connectorCtx.ClusterKey
	formattedKey := strings.ReplaceAll(key, "/", "-")
	klog.Infof("Creating cache for Cluster [%s] ...", key)

	eventBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: api.Client.EventsV1()})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, fmt.Sprintf("erie-canal-cluster-connector-remote-%s", formattedKey))

	c := &RemoteCache{
		connectorConfig: connectorCtx.ConnectorConfig,
		k8sAPI:          api,
		recorder:        recorder,
		clusterCfg:      clusterCfg,
		broadcaster:     eventBroadcaster,
		broker:          broker,
	}

	ecInformerFactory := ecinformers.NewSharedInformerFactoryWithOptions(api.FlomeshClient, resyncPeriod)
	serviceExportController := cachectrl.NewServiceExportControllerWithEventHandler(
		ecInformerFactory.Serviceexport().V1alpha1().ServiceExports(),
		resyncPeriod,
		c,
	)

	c.controllers = &controller.RemoteControllers{
		ServiceExport: serviceExportController,
	}

	minSyncPeriod := 3 * time.Second
	syncPeriod := 30 * time.Second
	burstSyncs := 2
	runnerName := fmt.Sprintf("sync-runner-%s", formattedKey)
	c.syncRunner = async.NewBoundedFrequencyRunner(runnerName, c.syncManagedCluster, minSyncPeriod, syncPeriod, burstSyncs)

	return c
}

func (c *RemoteCache) setInitialized(value bool) {
	var initialized int32
	if value {
		initialized = 1
	}
	atomic.StoreInt32(&c.initialized, initialized)
}

func (c *RemoteCache) syncManagedCluster() {
	// Nothing to do for the time-being

	//c.mu.Lock()
	//defer c.mu.Unlock()
	klog.Infof("[%s] Syncing resources of managed clusters ...", c.connectorConfig.Key())
}

func (c *RemoteCache) Sync() {
	c.syncRunner.Run()
}

func (c *RemoteCache) SyncLoop(stopCh <-chan struct{}) {
	c.syncRunner.Loop(stopCh)
}

func (c *RemoteCache) GetBroadcaster() events.EventBroadcaster {
	return c.broadcaster
}

func (c *RemoteCache) GetControllers() controller.Controllers {
	return c.controllers
}

func (c *RemoteCache) GetRecorder() events.EventRecorder {
	return c.recorder
}
