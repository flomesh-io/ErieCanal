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

package v1alpha1

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	clusterv1alpha1 "github.com/flomesh-io/ErieCanal/apis/cluster/v1alpha1"
	svcexpv1alpha1 "github.com/flomesh-io/ErieCanal/apis/serviceexport/v1alpha1"
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	conn "github.com/flomesh-io/ErieCanal/pkg/cluster"
	cctx "github.com/flomesh-io/ErieCanal/pkg/cluster/context"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/event"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"github.com/flomesh-io/ErieCanal/pkg/util"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metautil "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"time"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	k8sAPI      *kube.K8sAPI
	recorder    record.EventRecorder
	configStore *config.Store
	broker      *event.Broker
	certMgr     certificate.Manager
	backgrounds map[string]*connectorBackground
	mu          sync.Mutex
}

type connectorBackground struct {
	isInCluster bool
	context     cctx.ConnectorContext
	connector   conn.Connector
}

func New(
	client client.Client,
	api *kube.K8sAPI,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
	store *config.Store,
	broker *event.Broker,
	certMgr certificate.Manager,
	stop <-chan struct{},
) *ClusterReconciler {
	r := &ClusterReconciler{
		Client:      client,
		Scheme:      scheme,
		k8sAPI:      api,
		recorder:    recorder,
		configStore: store,
		broker:      broker,
		certMgr:     certMgr,
		backgrounds: make(map[string]*connectorBackground),
	}

	go r.processEvent(broker, stop)

	return r
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the Cluster instance
	cluster := &clusterv1alpha1.Cluster{}
	if err := r.Get(
		ctx,
		client.ObjectKey{Name: req.Name},
		cluster,
	); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			klog.V(3).Info("Cluster resource not found. Stopping the connector and remove the reference.")
			r.destroyConnector(cluster)

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		klog.Errorf("Failed to get Cluster, %#v", err)
		return ctrl.Result{}, err
	}

	mc := r.configStore.MeshConfig.GetConfig()

	result, err := r.deriveCodebases(mc)
	if err != nil {
		return result, err
	}

	key := cluster.Key()
	klog.V(5).Infof("Cluster key is %s", key)
	bg, exists := r.backgrounds[key]
	if exists && bg.context.SpecHash != util.SimpleHash(cluster.Spec) {
		klog.V(5).Infof("Background context of cluster [%s] exists, ")
		// exists and the spec changed, then stop it and start a new one
		if result, err = r.recreateConnector(ctx, bg, cluster, mc); err != nil {
			return result, err
		}
	} else if !exists {
		// doesn't exist, just create a new one
		if result, err = r.createConnector(ctx, cluster, mc); err != nil {
			return result, err
		}
	} else {
		klog.V(2).Infof("The connector %s already exists and the spec doesn't change", key)
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) deriveCodebases(mc *config.MeshConfig) (ctrl.Result, error) {
	repoClient := repo.NewRepoClient(mc.RepoRootURL())

	defaultServicesPath := mc.GetDefaultServicesPath()
	if err := repoClient.DeriveCodebase(defaultServicesPath, commons.DefaultServiceBasePath); err != nil {
		return ctrl.Result{RequeueAfter: 1 * time.Second}, err
	}

	defaultIngressPath := mc.GetDefaultIngressPath()
	if err := repoClient.DeriveCodebase(defaultIngressPath, commons.DefaultIngressBasePath); err != nil {
		return ctrl.Result{RequeueAfter: 1 * time.Second}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) createConnector(ctx context.Context, cluster *clusterv1alpha1.Cluster, mc *config.MeshConfig) (ctrl.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.newConnector(ctx, cluster, mc)
}

func (r *ClusterReconciler) recreateConnector(ctx context.Context, bg *connectorBackground, cluster *clusterv1alpha1.Cluster, mc *config.MeshConfig) (ctrl.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	close(bg.context.StopCh)
	delete(r.backgrounds, cluster.Key())

	return r.newConnector(ctx, cluster, mc)
}

func (r *ClusterReconciler) destroyConnector(cluster *clusterv1alpha1.Cluster) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := cluster.Key()
	if bg, exists := r.backgrounds[key]; exists {
		close(bg.context.StopCh)
		delete(r.backgrounds, key)
	}
}

func (r *ClusterReconciler) newConnector(ctx context.Context, cluster *clusterv1alpha1.Cluster, mc *config.MeshConfig) (ctrl.Result, error) {
	key := cluster.Key()

	kubeconfig, result, err := getKubeConfig(cluster)
	if err != nil {
		klog.Errorf("Failed to get kubeconfig for cluster %q: %s", cluster.Key(), err)
		return result, err
	}

	connCfg, err := r.connectorConfig(cluster, mc)
	if err != nil {
		return ctrl.Result{}, err
	}

	background := cctx.ConnectorContext{
		ClusterKey:      key,
		KubeConfig:      kubeconfig,
		ConnectorConfig: connCfg,
		SpecHash:        util.SimpleHash(cluster.Spec),
	}
	_, cancel := context.WithCancel(&background)
	stop := util.RegisterExitHandlers(cancel)
	background.Cancel = cancel
	background.StopCh = stop

	connector, err := conn.NewConnector(&background, r.broker, r.certMgr, 15*time.Minute)
	if err != nil {
		klog.Errorf("Failed to create connector for cluster %q: %s", cluster.Key(), err)
		return ctrl.Result{}, err
	}

	r.backgrounds[key] = &connectorBackground{
		isInCluster: cluster.Spec.IsInCluster,
		context:     background,
		connector:   connector,
	}

	success := true
	errorMsg := ""
	go func() {
		if err := connector.Run(stop); err != nil {
			success = false
			errorMsg = err.Error()
			klog.Errorf("Failed to run connector for cluster %q: %s", cluster.Key(), err)
			close(stop)
			delete(r.backgrounds, key)
		}
	}()

	if !cluster.Spec.IsInCluster {
		if success {
			return r.successJoinClusterSet(ctx, cluster, mc)
		} else {
			return r.failedJoinClusterSet(ctx, cluster, errorMsg)
		}
	}

	return ctrl.Result{}, nil
}

func getKubeConfig(cluster *clusterv1alpha1.Cluster) (*rest.Config, ctrl.Result, error) {
	if cluster.Spec.IsInCluster {
		kubeconfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, ctrl.Result{}, err
		}

		return kubeconfig, ctrl.Result{}, nil
	} else {
		return remoteKubeConfig(cluster)
	}
}

func remoteKubeConfig(cluster *clusterv1alpha1.Cluster) (*rest.Config, ctrl.Result, error) {
	// use the current context in kubeconfig
	kubeconfig, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (*api.Config, error) {
		cfg := api.Config{}
		if err := json.Unmarshal([]byte(cluster.Spec.Kubeconfig), &cfg); err != nil {
			return nil, err
		}

		return &cfg, nil
	})

	if err != nil {
		return nil, ctrl.Result{}, err
	}

	return kubeconfig, ctrl.Result{}, nil
}

func (r *ClusterReconciler) connectorConfig(cluster *clusterv1alpha1.Cluster, mc *config.MeshConfig) (*config.ConnectorConfig, error) {
	if cluster.Spec.IsInCluster {
		return config.NewConnectorConfig(
			mc.Cluster.Region,
			mc.Cluster.Zone,
			mc.Cluster.Group,
			mc.Cluster.Name,
			cluster.Spec.GatewayHost,
			cluster.Spec.GatewayPort,
			cluster.Spec.IsInCluster,
			"",
		)
	} else {
		return config.NewConnectorConfig(
			cluster.Spec.Region,
			cluster.Spec.Zone,
			cluster.Spec.Group,
			cluster.Name,
			cluster.Spec.GatewayHost,
			cluster.Spec.GatewayPort,
			cluster.Spec.IsInCluster,
			mc.Cluster.UID,
		)
	}
}

func (r *ClusterReconciler) processEvent(broker *event.Broker, stop <-chan struct{}) {
	msgBus := broker.GetMessageBus()
	svcExportCreatedCh := msgBus.Sub(string(event.ServiceExportCreated))
	defer broker.Unsub(msgBus, svcExportCreatedCh)

	for {
		// FIXME: refine it later

		select {
		case msg, ok := <-svcExportCreatedCh:
			mc := r.configStore.MeshConfig.GetConfig()
			// ONLY Control Plane takes care of the federation of service export/import
			if mc.IsManaged && mc.Cluster.ControlPlaneUID != "" && mc.Cluster.UID != mc.Cluster.ControlPlaneUID {
				klog.V(5).Infof("Ignore processing ServiceExportCreated event due to cluster is managed and not a control plane ...")
				continue
			}

			if !ok {
				klog.Warningf("Channel closed for ServiceExport")
				continue
			}
			klog.V(5).Infof("Received event ServiceExportCreated %v", msg)

			e, ok := msg.(event.Message)
			if !ok {
				klog.Errorf("Received unexpected message %T on channel, expected Message", e)
				continue
			}

			svcExportEvt, ok := e.NewObj.(*event.ServiceExportEvent)
			if !ok {
				klog.Errorf("Received unexpected object %T, expected *event.ServiceExportEvent", svcExportEvt)
				continue
			}

			// check ServiceExport Status, Invalid and Conflict ServiceExport is ignored
			export := svcExportEvt.ServiceExport
			if metautil.IsStatusConditionFalse(export.Status.Conditions, string(svcexpv1alpha1.ServiceExportValid)) {
				klog.Warningf("ServiceExport %#v is ignored due to Valid status is false", export)
				continue
			}
			if metautil.IsStatusConditionTrue(export.Status.Conditions, string(svcexpv1alpha1.ServiceExportConflict)) {
				klog.Warningf("ServiceExport %#v is ignored due to Conflict status is true", export)
				continue
			}

			r.processServiceExportCreatedEvent(svcExportEvt)
		case <-stop:
			klog.Infof("Received stop signal.")
			return
		}
	}
}

func (r *ClusterReconciler) processServiceExportCreatedEvent(svcExportEvt *event.ServiceExportEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	export := svcExportEvt.ServiceExport
	if r.isFirstTimeExport(svcExportEvt) {
		klog.V(5).Infof("[%s] ServiceExport %s/%s is exported first in the cluster set, will be accepted", svcExportEvt.Geo.Key(), export.Namespace, export.Name)
		r.acceptServiceExport(svcExportEvt)
	} else {
		valid, err := r.isValidServiceExport(svcExportEvt)
		if valid {
			klog.V(5).Infof("[%s] ServiceExport %s/%s is valid, will be accepted", svcExportEvt.Geo.Key(), export.Namespace, export.Name)
			r.acceptServiceExport(svcExportEvt)
		} else {
			klog.V(5).Infof("[%s] ServiceExport %s/%s is invalid, will be rejected", svcExportEvt.Geo.Key(), export.Namespace, export.Name)
			r.rejectServiceExport(svcExportEvt, err)
		}
	}
}

func (r *ClusterReconciler) isFirstTimeExport(event *event.ServiceExportEvent) bool {
	export := event.ServiceExport
	for _, bg := range r.backgrounds {
		if bg.isInCluster {
			continue
		}
		remoteConnector := bg.connector.(*conn.RemoteConnector)
		if remoteConnector.ServiceImportExists(export) {
			klog.Warningf("[%s] ServiceExport %s/%s exists in Cluster %s", event.Geo.Key(), export.Namespace, export.Name, bg.context.ClusterKey)
			return false
		}
	}

	return true
}

func (r *ClusterReconciler) isValidServiceExport(svcExportEvt *event.ServiceExportEvent) (bool, error) {
	export := svcExportEvt.ServiceExport
	for _, bg := range r.backgrounds {
		if bg.isInCluster {
			continue
		}

		connectorContext := bg.context
		if connectorContext.ClusterKey == svcExportEvt.ClusterKey() {
			// no need to test against itself
			continue
		}

		remoteConnector := bg.connector.(*conn.RemoteConnector)
		if err := remoteConnector.ValidateServiceExport(svcExportEvt.ServiceExport, svcExportEvt.Service); err != nil {
			klog.Warningf("[%s] ServiceExport %s/%s has conflict in Cluster %s", svcExportEvt.Geo.Key(), export.Namespace, export.Name, connectorContext.ClusterKey)
			return false, err
		}
	}

	return true, nil
}

func (r *ClusterReconciler) acceptServiceExport(svcExportEvt *event.ServiceExportEvent) {
	r.broker.Enqueue(
		event.Message{
			Kind:   event.ServiceExportAccepted,
			OldObj: nil,
			NewObj: svcExportEvt,
		},
	)
}

func (r *ClusterReconciler) rejectServiceExport(svcExportEvt *event.ServiceExportEvent, err error) {
	svcExportEvt.Error = err.Error()

	r.broker.Enqueue(
		event.Message{
			Kind:   event.ServiceExportRejected,
			OldObj: nil,
			NewObj: svcExportEvt,
		},
	)
}

func (r *ClusterReconciler) successJoinClusterSet(ctx context.Context, cluster *clusterv1alpha1.Cluster, mc *config.MeshConfig) (ctrl.Result, error) {
	metautil.SetStatusCondition(&cluster.Status.Conditions, metav1.Condition{
		Type:               string(clusterv1alpha1.ClusterManaged),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: cluster.Generation,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "Success",
		Message:            fmt.Sprintf("Cluster %s joined ClusterSet successfully.", cluster.Key()),
	})

	if err := r.Status().Update(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) failedJoinClusterSet(ctx context.Context, cluster *clusterv1alpha1.Cluster, err string) (ctrl.Result, error) {
	metautil.SetStatusCondition(&cluster.Status.Conditions, metav1.Condition{
		Type:               string(clusterv1alpha1.ClusterManaged),
		Status:             metav1.ConditionFalse,
		ObservedGeneration: cluster.Generation,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "Failed",
		Message:            fmt.Sprintf("Cluster %s failed to join ClusterSet: %s.", cluster.Key(), err),
	})

	if err := r.Status().Update(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1alpha1.Cluster{}).
		Owns(&corev1.Secret{}).
		Owns(&appv1.Deployment{}).
		Complete(r)
}
