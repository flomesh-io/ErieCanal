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
	svcimpv1alpha1 "github.com/flomesh-io/ErieCanal/apis/serviceimport/v1alpha1"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceImportReconciler reconciles a ServiceImport object
type ServiceImportReconciler struct {
	client.Client
	K8sAPI                  *kube.K8sAPI
	Scheme                  *runtime.Scheme
	Recorder                record.EventRecorder
	ControlPlaneConfigStore *config.Store
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the ServiceImport closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceImport object against the actual ServiceImport state, and then
// perform operations to make the ServiceImport state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ServiceImportReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//mc := r.ControlPlaneConfigStore.MeshConfig.GetConfig()

	ServiceImport := &svcimpv1alpha1.ServiceImport{}
	if err := r.Get(
		ctx,
		client.ObjectKey{Name: req.Name, Namespace: req.Namespace},
		ServiceImport,
	); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			klog.V(3).Info("[ServiceImport] ServiceImport resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		klog.Errorf("Failed to get ServiceImport, %#v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceImportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&svcimpv1alpha1.ServiceImport{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Endpoints{}).
		Complete(r)
}
