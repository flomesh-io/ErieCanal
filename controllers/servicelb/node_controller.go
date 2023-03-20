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

package servicelb

import (
	"context"
	_ "embed"
	"github.com/flomesh-io/ErieCanal/controllers"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodeReconciler reconciles a Node object
type nodeReconciler struct {
	recorder record.EventRecorder
	cfg      *controllers.ReconcilerConfig
}

func NewNodeReconciler(rc *controllers.ReconcilerConfig) controllers.Reconciler {
	return &nodeReconciler{
		recorder: rc.Manager.GetEventRecorderFor("ServiceLB"),
		cfg:      rc,
	}
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Node object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *nodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the Node instance
	node := &corev1.Node{}
	if err := r.cfg.Client.Get(
		ctx,
		req.NamespacedName,
		node,
	); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			klog.V(3).Info("Node resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		klog.Errorf("Failed to get Node, %#v", err)
		return ctrl.Result{}, err
	}

	if _, ok := node.Labels[daemonsetNodeLabel]; !ok {
		return ctrl.Result{}, nil
	}

	if err := r.updateDaemonSets(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *nodeReconciler) updateDaemonSets(ctx context.Context) error {
	klog.V(5).Infof("Updating DaemonSets due to node labels change ...")

	daemonsets := &appv1.DaemonSetList{}
	if err := r.cfg.Client.List(
		ctx,
		daemonsets,
		client.InNamespace(corev1.NamespaceAll),
		client.MatchingLabels{
			nodeSelectorLabel: "false",
		},
	); err != nil {
		return err
	}

	for _, ds := range daemonsets.Items {
		ds.Spec.Template.Spec.NodeSelector = map[string]string{
			daemonsetNodeLabel: "true",
		}
		ds.Labels[nodeSelectorLabel] = "true"
		if err := r.cfg.Client.Update(ctx, &ds); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *nodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
