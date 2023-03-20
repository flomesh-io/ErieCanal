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
	"github.com/flomesh-io/ErieCanal/controllers"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// endpointSliceReconciler reconciles an EndpointSlice object
type endpointSliceReconciler struct {
	recorder record.EventRecorder
	cfg      *controllers.ReconcilerConfig
}

func NewEndpointSliceReconciler(rc *controllers.ReconcilerConfig) controllers.Reconciler {
	return &endpointSliceReconciler{
		recorder: rc.Manager.GetEventRecorderFor("EndpointSlice"),
		cfg:      rc,
	}
}

func (r *endpointSliceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	epSlice := &discoveryv1.EndpointSlice{}
	if err := r.cfg.Client.Get(ctx, req.NamespacedName, epSlice); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if shouldIgnoreEndpointSlice(epSlice) {
		return ctrl.Result{}, nil
	}

	// Ensure the EndpointSlice is labelled to match the ServiceImport's derived
	// Service.
	serviceName := derivedName(types.NamespacedName{Namespace: epSlice.Namespace, Name: epSlice.Labels[commons.MultiClusterLabelServiceName]})
	if epSlice.Labels[discoveryv1.LabelServiceName] == serviceName {
		return ctrl.Result{}, nil
	}

	epSlice.Labels[discoveryv1.LabelServiceName] = serviceName
	epSlice.Labels[commons.MultiClusterLabelServiceName] = serviceName
	if err := r.cfg.Client.Update(ctx, epSlice); err != nil {
		return ctrl.Result{}, err
	}

	klog.Info("added label", discoveryv1.LabelServiceName, serviceName)

	return ctrl.Result{}, nil
}

func shouldIgnoreEndpointSlice(epSlice *discoveryv1.EndpointSlice) bool {
	if epSlice.DeletionTimestamp != nil {
		return true
	}

	if epSlice.Labels[commons.MultiClusterLabelServiceName] == "" {
		return true
	}

	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *endpointSliceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&discoveryv1.EndpointSlice{}).
		Complete(r)
}
