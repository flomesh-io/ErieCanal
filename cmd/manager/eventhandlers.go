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

package main

import (
	"context"
	"fmt"
	svcimpv1alpha1 "github.com/flomesh-io/ErieCanal/apis/serviceimport/v1alpha1"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/event/handler"
	gwcache "github.com/flomesh-io/ErieCanal/pkg/gateway/cache"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	rtcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"time"
)

func (c *ManagerConfig) GetResourceEventHandler() handler.EventHandler {

	gatewayCache := gwcache.NewGatewayCache(gwcache.GatewayCacheConfig{
		Client: c.manager.GetClient(),
		Cache:  c.manager.GetCache(),
	})

	return handler.NewEventHandler(handler.EventHandlerConfig{
		MinSyncPeriod: 5 * time.Second,
		SyncPeriod:    30 * time.Second,
		BurstSyncs:    5,
		Cache:         gatewayCache,
		SyncFunc:      gatewayCache.BuildConfigs,
		StopCh:        c.stopCh,
	})

}

func (c *ManagerConfig) RegisterEventHandlers() error {
	// FIXME: make it configurable
	resyncPeriod := 15 * time.Minute

	configHandler := config.NewConfigurationHandler(
		config.NewFlomeshConfigurationHandler(
			c.manager.GetClient(),
			c.k8sAPI,
			c.configStore,
			c.certificateManager,
		),
	)

	if err := informOnResource(&corev1.ConfigMap{}, configHandler, c.manager.GetCache(), resyncPeriod); err != nil {
		klog.Errorf("failed to create informer for configmaps: %s", err)
		return err
	}

	mc := c.configStore.MeshConfig.GetConfig()
	if mc.IsGatewayApiEnabled() {
		if c.eventHandler == nil {
			return fmt.Errorf("GatewayAPI is enabled, but no valid EventHanlder is provided")
		}

		for name, r := range map[string]client.Object{
			"namespaces":     &corev1.Namespace{},
			"services":       &corev1.Service{},
			"serviceimports": &svcimpv1alpha1.ServiceImport{},
			"endpointslices": &discoveryv1.EndpointSlice{},
			"gatewayclasses": &gwv1beta1.GatewayClass{},
			"gateways":       &gwv1beta1.Gateway{},
			"httproutes":     &gwv1beta1.HTTPRoute{},
		} {
			if err := informOnResource(r, c.eventHandler, c.manager.GetCache(), resyncPeriod); err != nil {
				klog.Errorf("failed to create informer for %s: %s", name, err)
				return err
			}
		}

		ctx := context.TODO()

		if !c.manager.GetCache().WaitForCacheSync(ctx) {
			err := fmt.Errorf("informer cache failed to sync")
			klog.Error(err)
			return err
		}

		//go func() {
		//    if err := c.eventHandler.Start(ctx); err != nil {
		//        panic(err)
		//    }
		//}()

	}

	return nil
}

func informOnResource(obj client.Object, handler cache.ResourceEventHandler, cache rtcache.Cache, resyncPeriod time.Duration) error {
	informer, err := cache.GetInformer(context.TODO(), obj)
	if err != nil {
		return err
	}

	if handler != nil {
		_, err = informer.AddEventHandlerWithResyncPeriod(handler, resyncPeriod)
		return err
	}

	return nil
}
