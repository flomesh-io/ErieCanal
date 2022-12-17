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

package v1beta1

import (
	"context"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gwv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	GatewayV1alpha2Controller = "flomesh.io/gateway-v1alpha2-controller"
)

type GatewayReconciler struct {
	client.Client
	K8sAPI   *kube.K8sAPI
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gwv1beta1.Gateway{}).
		Watches(
			&source.Kind{Type: &gwv1beta1.GatewayClass{}},
			handler.EnqueueRequestsFromMapFunc(r.gatewayClassToGateways),
			builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
				gatewayClass, ok := obj.(*gwv1beta1.GatewayClass)
				if !ok {
					klog.Infof("unexpected object type: %T", obj)
					return false
				}

				return gatewayClass.Spec.ControllerName == GatewayV1alpha2Controller
			})),
		).
		Complete(r)
}

func (r *GatewayReconciler) gatewayClassToGateways(gatewayClass client.Object) []reconcile.Request {
	var gateways gwv1beta1.GatewayList
	if err := r.Client.List(context.Background(), &gateways); err != nil {
		klog.Error("error listing gateways")
		return nil
	}

	var reconciles []reconcile.Request
	for _, gw := range gateways.Items {
		if string(gw.Spec.GatewayClassName) == gatewayClass.GetName() {
			reconciles = append(reconciles, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: gw.Namespace,
					Name:      gw.Name,
				},
			})
		}
	}

	return reconciles
}
