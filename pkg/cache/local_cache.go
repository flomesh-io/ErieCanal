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

package cache

import (
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/flomesh-io/ErieCanal/pkg/cache/controller"
	conn "github.com/flomesh-io/ErieCanal/pkg/cluster/context"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	cachectrl "github.com/flomesh-io/ErieCanal/pkg/controller"
	"github.com/flomesh-io/ErieCanal/pkg/event"
	eriecanalinformers "github.com/flomesh-io/ErieCanal/pkg/generated/informers/externalversions"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	routepkg "github.com/flomesh-io/ErieCanal/pkg/route"
	"github.com/flomesh-io/ErieCanal/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/events"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/util/async"
	"sync"
	"sync/atomic"
	"time"
)

type LocalCache struct {
	connectorConfig *config.ConnectorConfig
	k8sAPI          *kube.K8sAPI
	recorder        events.EventRecorder
	clusterCfg      *config.Store
	broker          *event.Broker

	serviceChanges       *ServiceChangeTracker
	endpointsChanges     *EndpointChangeTracker
	ingressChanges       *IngressChangeTracker
	serviceImportChanges *ServiceImportChangeTracker

	serviceMap               ServiceMap
	endpointsMap             EndpointsMap
	ingressMap               IngressMap
	serviceImportMap         ServiceImportMap
	multiClusterEndpointsMap MultiClusterEndpointsMap

	mu sync.Mutex

	endpointsSynced      bool
	servicesSynced       bool
	ingressesSynced      bool
	ingressClassesSynced bool
	serviceImportSynced  bool
	initialized          int32

	syncRunner *async.BoundedFrequencyRunner
	repoClient *repo.PipyRepoClient

	controllers *controller.LocalControllers
	broadcaster events.EventBroadcaster

	ingressRoutesVersion string
	serviceRoutesVersion string
}

func newLocalCache(ctx context.Context, api *kube.K8sAPI, clusterCfg *config.Store, broker *event.Broker, resyncPeriod time.Duration) *LocalCache {
	connectorCtx := ctx.(*conn.ConnectorContext)
	eventBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: api.Client.EventsV1()})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, "erie-canal-cluster-connector-local")
	mc := clusterCfg.MeshConfig.GetConfig()

	c := &LocalCache{
		connectorConfig:          connectorCtx.ConnectorConfig,
		k8sAPI:                   api,
		recorder:                 recorder,
		clusterCfg:               clusterCfg,
		serviceMap:               make(ServiceMap),
		serviceImportMap:         make(ServiceImportMap),
		endpointsMap:             make(EndpointsMap),
		ingressMap:               make(IngressMap),
		multiClusterEndpointsMap: make(MultiClusterEndpointsMap),
		repoClient:               repo.NewRepoClient(mc.RepoRootURL()),
		broadcaster:              eventBroadcaster,
		broker:                   broker,
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(api.Client, resyncPeriod)
	serviceController := cachectrl.NewServiceControllerWithEventHandler(
		informerFactory.Core().V1().Services(),
		resyncPeriod,
		c,
	)
	endpointsController := cachectrl.NewEndpointsControllerWithEventHandler(
		informerFactory.Core().V1().Endpoints(),
		resyncPeriod,
		c,
	)
	ingressClassV1Controller := cachectrl.NewIngressClassv1ControllerWithEventHandler(
		informerFactory.Networking().V1().IngressClasses(),
		resyncPeriod,
		c,
	)
	ingressV1Controller := cachectrl.NewIngressv1ControllerWithEventHandler(
		informerFactory.Networking().V1().Ingresses(),
		resyncPeriod,
		c,
	)

	erieCanalInformerFactory := eriecanalinformers.NewSharedInformerFactoryWithOptions(api.FlomeshClient, resyncPeriod)
	serviceImortController := cachectrl.NewServiceImportControllerWithEventHandler(
		erieCanalInformerFactory.Serviceimport().V1alpha1().ServiceImports(),
		resyncPeriod,
		c,
	)

	c.controllers = &controller.LocalControllers{
		Service:        serviceController,
		Endpoints:      endpointsController,
		Ingressv1:      ingressV1Controller,
		IngressClassv1: ingressClassV1Controller,
		ServiceImport:  serviceImortController,
	}

	c.serviceChanges = NewServiceChangeTracker(enrichServiceInfo, recorder, c.controllers)
	c.serviceImportChanges = NewServiceImportChangeTracker(enrichServiceImportInfo, nil, recorder, c.controllers)
	c.endpointsChanges = NewEndpointChangeTracker(nil, recorder, c.controllers)
	c.ingressChanges = NewIngressChangeTracker(api, c.controllers, recorder, enrichIngressInfo)

	// FIXME: make it configurable
	minSyncPeriod := 3 * time.Second
	syncPeriod := 30 * time.Second
	burstSyncs := 2
	c.syncRunner = async.NewBoundedFrequencyRunner("sync-runner-local", c.syncRoutes, minSyncPeriod, syncPeriod, burstSyncs)

	return c
}

func (c *LocalCache) GetControllers() controller.Controllers {
	return c.controllers
}

func (c *LocalCache) GetBroadcaster() events.EventBroadcaster {
	return c.broadcaster
}

func (c *LocalCache) GetRecorder() events.EventRecorder {
	return c.recorder
}

func (c *LocalCache) setInitialized(value bool) {
	var initialized int32
	if value {
		initialized = 1
	}
	atomic.StoreInt32(&c.initialized, initialized)
}

func (c *LocalCache) isInitialized() bool {
	return atomic.LoadInt32(&c.initialized) > 0
}

func (c *LocalCache) Sync() {
	c.syncRunner.Run()
}

// SyncLoop runs periodic work.  This is expected to run as a goroutine or as the main loop of the app.  It does not return.
func (c *LocalCache) SyncLoop(stopCh <-chan struct{}) {
	c.syncRunner.Loop(stopCh)
}

func (c *LocalCache) syncRoutes() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.serviceMap.Update(c.serviceChanges)
	klog.V(5).Infof("Service Map: %#v", c.serviceMap)

	c.serviceImportMap.Update(c.serviceImportChanges)
	klog.V(5).Infof("ServiceImport Map: %#v", c.serviceImportMap)

	c.multiClusterEndpointsMap.Update(c.serviceImportChanges)
	klog.V(5).Infof("MultiCluster Endpoints Map: %#v", c.multiClusterEndpointsMap)

	c.endpointsMap.Update(c.endpointsChanges)
	klog.V(5).Infof("Endpoints Map: %#v", c.endpointsMap)

	c.ingressMap.Update(c.ingressChanges)
	klog.V(5).Infof("Ingress Map: %#v", c.ingressMap)

	klog.V(3).InfoS("Start syncing rules ...")

	r := routepkg.RouteBase{
		Region:      c.connectorConfig.Region(),
		Zone:        c.connectorConfig.Zone(),
		Group:       c.connectorConfig.Group(),
		Cluster:     c.connectorConfig.Name(),
		GatewayHost: c.connectorConfig.GatewayHost(),
		GatewayIP:   c.connectorConfig.GatewayIP(),
		GatewayPort: c.connectorConfig.GatewayPort(),
	}

	mc := c.clusterCfg.MeshConfig.GetConfig()

	ingressRoutes := c.buildIngressRoutes(r)
	klog.V(5).Infof("Ingress Routes:\n %#v", ingressRoutes)
	if c.ingressRoutesVersion != ingressRoutes.Hash {
		klog.V(5).Infof("Ingress Routes changed, old hash=%q, new hash=%q", c.ingressRoutesVersion, ingressRoutes.Hash)
		//c.ingressRoutesVersion = ingressRoutes.Hash
		//go c.aggregatorClient.PostIngresses(ingressRoutes)
		batches := ingressBatches(ingressRoutes, mc)
		if batches != nil {
			go func() {
				if err := c.repoClient.Batch(batches); err != nil {
					klog.Errorf("Sync ingress routes to repo failed: %s", err)
					return
				}

				klog.V(5).Infof("Updating ingress routes version ...")
				c.ingressRoutesVersion = ingressRoutes.Hash
			}()
		}
	}

	serviceRoutes := c.buildServiceRoutes(r)
	klog.V(5).Infof("Service Routes:\n %#v", serviceRoutes)
	if c.serviceRoutesVersion != serviceRoutes.Hash {
		klog.V(5).Infof("Service Routes changed, old hash=%q, new hash=%q", c.serviceRoutesVersion, serviceRoutes.Hash)
		//c.serviceRoutesVersion = serviceRoutes.Hash
		//go c.aggregatorClient.PostServices(serviceRoutes)
		batches := serviceBatches(serviceRoutes, mc)
		if batches != nil {
			go func() {
				if err := c.repoClient.Batch(batches); err != nil {
					klog.Errorf("Sync service routes to repo failed: %s", err)
					return
				}

				klog.V(5).Infof("Updating service routes version ...")
				c.serviceRoutesVersion = serviceRoutes.Hash
			}()
		}
	}
}

func (c *LocalCache) buildIngressRoutes(r routepkg.RouteBase) routepkg.IngressRoute {
	ingressRoutes := routepkg.IngressRoute{
		RouteBase: r,
		//Hash:      hash,
		Routes: []routepkg.IngressRouteEntry{},
	}

	for svcName, route := range c.ingressMap {
		ir := routepkg.IngressRouteEntry{
			Host:        route.Host(),
			Path:        route.Path(),
			ServiceName: svcName.String(),
			Rewrite:     route.Rewrite(),
			Sticky:      route.SessionSticky(),
			Balancer:    route.LBType(),
			Upstreams:   []routepkg.EndpointEntry{},
		}
		for _, e := range c.endpointsMap[svcName] {
			ep, ok := e.(*BaseEndpointInfo)
			if !ok {
				klog.ErrorS(nil, "Failed to cast BaseEndpointInfo", "endpoint", e.String())
				continue
			}

			epIP := ep.IP()
			epPort, err := ep.Port()
			// Error parsing this endpoint has been logged. Skip to next endpoint.
			if epIP == "" || err != nil {
				continue
			}
			entry := routepkg.EndpointEntry{
				IP:   epIP,
				Port: epPort,
				//Protocol: protocol,
			}
			ir.Upstreams = append(ir.Upstreams, entry)
		}

		ingressRoutes.Routes = append(ingressRoutes.Routes, ir)
	}

	ingressRoutes.Hash = util.SimpleHash(ingressRoutes)

	return ingressRoutes
}

func ingressBatches(ingressRoutes routepkg.IngressRoute, mc *config.MeshConfig) []repo.Batch {
	batch := repo.Batch{
		Basepath: mc.GetDefaultIngressPath(),
		Items:    []repo.BatchItem{},
	}

	// Generate router.json
	router := repo.Router{Routes: repo.RouterEntry{}}
	// Generate balancer.json
	balancer := repo.Balancer{Services: repo.BalancerEntry{}}

	for _, r := range ingressRoutes.Routes {
		// router
		router.Routes[routerKey(r)] = repo.ServiceInfo{Service: r.ServiceName, Rewrite: r.Rewrite}

		// balancer
		balancer.Services[r.ServiceName] = upstream(r)
	}

	batch.Items = append(batch.Items, ingressBatchItems(router, balancer)...)
	if len(batch.Items) > 0 {
		return []repo.Batch{batch}
	}

	return nil
}

func (c *LocalCache) buildServiceRoutes(r routepkg.RouteBase) routepkg.ServiceRoute {
	// Build  rules for each service.
	serviceRoutes := routepkg.ServiceRoute{
		RouteBase: r,
		Routes:    []routepkg.ServiceRouteEntry{},
	}

	svcNames := mapset.NewSet[ServicePortName]()
	for svcName := range c.serviceMap {
		svcNames.Add(svcName)
	}
	for svcName := range c.serviceImportMap {
		svcNames.Add(svcName)
	}

	for _, svcName := range svcNames.ToSlice() {
		svc, exists := c.serviceMap[svcName]
		if exists {
			svcInfo, ok := svc.(*serviceInfo)
			if ok {
				sr := routepkg.ServiceRouteEntry{
					Name:      svcInfo.svcName.Name,
					Namespace: svcInfo.svcName.Namespace,
					Targets:   make([]routepkg.Target, 0),
					PortName:  svcInfo.portName,
				}

				switch svcInfo.Type {
				case corev1.ServiceTypeClusterIP:
					for _, ep := range c.endpointsMap[svcName] {
						sr.Targets = append(sr.Targets, routepkg.Target{
							Address: ep.String(),
							Tags: map[string]string{
								"Node": ep.NodeName(),
								"Host": ep.HostName(),
							}},
						)
					}
					serviceRoutes.Routes = append(serviceRoutes.Routes, sr)
				case corev1.ServiceTypeExternalName:
					sr.Targets = append(sr.Targets, routepkg.Target{
						Address: svcInfo.Address(),
						Tags:    map[string]string{}},
					)
					serviceRoutes.Routes = append(serviceRoutes.Routes, sr)
				}
			} else {
				klog.ErrorS(nil, "Failed to cast serviceInfo", "svcName", svcName.String())
			}
		}

		svcImp, exists := c.serviceImportMap[svcName]
		if exists {
			svcImpInfo, ok := svcImp.(*serviceImportInfo)
			if ok {
				sr := routepkg.ServiceRouteEntry{
					Name:      svcImpInfo.svcName.Name,
					Namespace: svcImpInfo.svcName.Namespace,
					Targets:   make([]routepkg.Target, 0),
					PortName:  svcImpInfo.portName,
				}

				for _, ep := range c.multiClusterEndpointsMap[svcName] {
					sr.Targets = append(sr.Targets, routepkg.Target{
						Address: ep.String(),
						Tags: map[string]string{
							"Cluster": ep.ClusterInfo(),
						}},
					)
				}

				serviceRoutes.Routes = append(serviceRoutes.Routes, sr)
			}
		}
	}
	serviceRoutes.Hash = util.SimpleHash(serviceRoutes)

	return serviceRoutes
}

func serviceBatches(serviceRoutes routepkg.ServiceRoute, mc *config.MeshConfig) []repo.Batch {
	registry := repo.ServiceRegistry{Services: repo.ServiceRegistryEntry{}}

	for _, route := range serviceRoutes.Routes {
		serviceName := servicePortName(route)
		registry.Services[serviceName] = append(registry.Services[serviceName], addresses(route)...)
	}

	batch := repo.Batch{
		Basepath: mc.GetDefaultServicesPath(),
		Items:    []repo.BatchItem{},
	}

	item := repo.BatchItem{
		Path:     "/config",
		Filename: "registry.json",
		Content:  registry,
	}

	batch.Items = append(batch.Items, item)
	if len(batch.Items) > 0 {
		return []repo.Batch{batch}
	}

	return nil
}

func routerKey(r routepkg.IngressRouteEntry) string {
	return fmt.Sprintf("%s%s", r.Host, r.Path)
}

func upstream(r routepkg.IngressRouteEntry) repo.Upstream {
	return repo.Upstream{
		Balancer: r.Balancer,
		Sticky:   r.Sticky,
		Targets:  transformTargets(r.Upstreams),
	}
}

func transformTargets(endpoints []routepkg.EndpointEntry) []string {
	if len(endpoints) == 0 {
		return []string{}
	}

	targets := sets.String{}
	for _, ep := range endpoints {
		targets.Insert(fmt.Sprintf("%s:%d", ep.IP, ep.Port))
	}

	return targets.List()
}

func ingressBatchItems(router repo.Router, balancer repo.Balancer) []repo.BatchItem {
	routerItem := repo.BatchItem{
		Path:     "/config",
		Filename: "router.json",
		Content:  router,
	}
	balancerItem := repo.BatchItem{
		Path:     "/config",
		Filename: "balancer.json",
		Content:  balancer,
	}
	return []repo.BatchItem{routerItem, balancerItem}
}

func servicePortName(route routepkg.ServiceRouteEntry) string {
	return fmt.Sprintf("%s/%s%s", route.Namespace, route.Name, fmtPortName(route.PortName))
}

func addresses(route routepkg.ServiceRouteEntry) []string {
	result := make([]string, 0)
	for _, target := range route.Targets {
		result = append(result, target.Address)
	}

	return result
}
