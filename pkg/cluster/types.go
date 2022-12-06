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
