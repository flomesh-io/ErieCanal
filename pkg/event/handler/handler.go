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

package handler

import (
	"context"
	gw "github.com/flomesh-io/ErieCanal/pkg/gateway"
	"github.com/google/go-cmp/cmp"
	"k8s.io/kubernetes/pkg/util/async"
	"time"
)

type SyncFunc func()

type EventHandlerConfig struct {
	MinSyncPeriod time.Duration
	SyncPeriod    time.Duration
	BurstSyncs    int
	Cache         gw.Cache
	SyncFunc      SyncFunc
	StopCh        <-chan struct{}
}

type ErieCanalEventHandler struct {
	cache      gw.Cache
	syncRunner *async.BoundedFrequencyRunner
	stopCh     <-chan struct{}
}

func NewEventHandler(config EventHandlerConfig) EventHandler {
	if config.SyncFunc == nil {
		panic("SyncFunc is required")
	}

	handler := &ErieCanalEventHandler{
		cache:  config.Cache,
		stopCh: config.StopCh,
	}
	handler.syncRunner = async.NewBoundedFrequencyRunner("gateway-sync-runner", config.SyncFunc, config.MinSyncPeriod, config.SyncPeriod, config.BurstSyncs)

	return handler
}

func (e *ErieCanalEventHandler) OnAdd(obj interface{}) {
	if e.onChange(nil, obj) {
		e.Sync()
	}
}

func (e *ErieCanalEventHandler) OnUpdate(oldObj, newObj interface{}) {
	if e.onChange(oldObj, newObj) {
		e.Sync()
	}
}

func (e *ErieCanalEventHandler) OnDelete(obj interface{}) {
	if e.onChange(obj, nil) {
		e.Sync()
	}
}

func (e *ErieCanalEventHandler) onChange(oldObj, newObj interface{}) bool {
	if newObj == nil {
		return e.cache.Delete(oldObj)
	} else {
		if oldObj == nil {
			return e.cache.Insert(newObj)
		} else {
			if cmp.Equal(oldObj, newObj) {
				return false
			}

			del := e.cache.Delete(oldObj)
			ins := e.cache.Insert(newObj)

			return del || ins
		}
	}
}

func (e *ErieCanalEventHandler) Sync() {
	e.syncRunner.Run()
}

func (e *ErieCanalEventHandler) Start(ctx context.Context) error {
	e.syncRunner.Loop(e.stopCh)
	return nil
}
