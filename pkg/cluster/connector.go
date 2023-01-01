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
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/cache"
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	conn "github.com/flomesh-io/ErieCanal/pkg/cluster/context"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/event"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/version"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"time"
)

func NewConnector(ctx context.Context, broker *event.Broker, certMgr certificate.Manager, resyncPeriod time.Duration) (Connector, error) {
	connectorCtx := ctx.(*conn.ConnectorContext)

	k8sAPI, err := kube.NewAPIForConfig(connectorCtx.KubeConfig, 30*time.Second)
	if err != nil {
		return nil, err
	}

	if !version.IsSupportedK8sVersion(k8sAPI) {
		err := fmt.Errorf("kubernetes server version %s is not supported, requires at least %s",
			version.ServerVersion.String(), version.MinK8sVersion.String())
		klog.Error(err)

		return nil, err
	}

	// checks if ErieCanal is installed in the cluster, this's a MUST otherwise it doesn't work
	_, err = k8sAPI.Client.AppsV1().
		Deployments(config.GetErieCanalNamespace()).
		Get(context.TODO(), commons.ManagerDeploymentName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Error("ErieCanal Control Plane is not installed or not in a proper state, please check it.")
			return nil, err
		}

		klog.Errorf("Get ErieCanal manager component %s/%s error: %s", config.GetErieCanalNamespace(), commons.ManagerDeploymentName, err)
		return nil, err
	}

	clusterCfg := config.NewStore(k8sAPI)
	connectorCache := cache.NewCache(connectorCtx, k8sAPI, clusterCfg, broker, certMgr, resyncPeriod)

	if connectorCtx.ConnectorConfig.IsInCluster() {
		return &LocalConnector{
			context:    connectorCtx,
			k8sAPI:     k8sAPI,
			cache:      connectorCache,
			clusterCfg: clusterCfg,
			broker:     broker,
		}, nil
	} else {
		return &RemoteConnector{
			context:    connectorCtx,
			k8sAPI:     k8sAPI,
			cache:      connectorCache,
			clusterCfg: clusterCfg,
			broker:     broker,
		}, nil
	}
}
