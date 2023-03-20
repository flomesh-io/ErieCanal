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
	"github.com/flomesh-io/ErieCanal/controllers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// serviceReconciler reconciles a Service object
type serviceReconciler struct {
	recorder record.EventRecorder
	cfg      *controllers.ReconcilerConfig
}

func NewServiceReconciler(rc *controllers.ReconcilerConfig) controllers.Reconciler {
	return &serviceReconciler{
		recorder: rc.Manager.GetEventRecorderFor("Service"),
		cfg:      rc,
	}
}

func (r *serviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	svc := &corev1.Service{}
	if err := r.cfg.Client.Get(ctx, req.NamespacedName, svc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// service is being deleted
	if svc.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	importName := serviceImportOwner(svc.OwnerReferences)
	// If ServiceImport name is empty, stands for it's not MCS or has not been linked to ServiceImport yet
	if importName == "" {
		return ctrl.Result{}, nil
	}

	svcImport := &svcimpv1alpha1.ServiceImport{}
	if err := r.cfg.Client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: importName}, svcImport); err != nil {
		return ctrl.Result{}, err
	}

	if len(svcImport.Spec.IPs) > 0 {
		return ctrl.Result{}, nil
	}

	svcImport.Spec.IPs = []string{svc.Spec.ClusterIP}
	if err := r.cfg.Client.Update(ctx, svcImport); err != nil {
		return ctrl.Result{}, err
	}
	klog.Infof("Updated ServiceImport %s/%s, ClusterIP: %s", req.Namespace, importName, svc.Spec.ClusterIP)

	return ctrl.Result{}, nil
}

func serviceImportOwner(refs []metav1.OwnerReference) string {
	for _, ref := range refs {
		if ref.APIVersion == svcimpv1alpha1.SchemeGroupVersion.String() && ref.Kind == serviceImportKind {
			return ref.Name
		}
	}
	return ""
}

// SetupWithManager sets up the controller with the Manager.
func (r *serviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}
