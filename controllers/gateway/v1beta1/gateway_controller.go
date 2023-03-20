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

package v1beta1

import (
	"context"
	"fmt"
	"github.com/flomesh-io/ErieCanal/controllers"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"k8s.io/apimachinery/pkg/api/errors"
	metautil "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	"sort"
	"time"
)

type gatewayReconciler struct {
	recorder record.EventRecorder
	cfg      *controllers.ReconcilerConfig
}

func NewGatewayReconciler(rc *controllers.ReconcilerConfig) controllers.Reconciler {
	return &gatewayReconciler{
		recorder: rc.Manager.GetEventRecorderFor("Gateway"),
		cfg:      rc,
	}
}

func (r *gatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	gateway := &gwv1beta1.Gateway{}
	if err := r.cfg.Client.Get(
		ctx,
		req.NamespacedName,
		gateway,
	); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			klog.V(3).Info("Gateway resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		klog.Errorf("Failed to get Gateway, %#v", err)
		return ctrl.Result{}, err
	}

	var gatewayClasses gwv1beta1.GatewayClassList
	if err := r.cfg.Client.List(ctx, &gatewayClasses); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to list gateway classes: %s", err)
	}

	var effectiveGatewayClass *gwv1beta1.GatewayClass
	for idx, cls := range gatewayClasses.Items {
		if isEffectiveGatewayClass(&cls) {
			effectiveGatewayClass = &gatewayClasses.Items[idx]
		}
	}

	if effectiveGatewayClass == nil {
		klog.Warningf("No effective GatewayClass, ignore processing Gateway resource %s.", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	// If current Gateway is linked to effective GatewayClass
	if string(gateway.Spec.GatewayClassName) == effectiveGatewayClass.Name {
		// 1. List all Gateways in the namespace whose GatewayClass is current effective class
		var gatewayList gwv1beta1.GatewayList
		if err := r.cfg.Client.List(ctx, &gatewayList, client.InNamespace(gateway.Namespace)); err != nil {
			klog.Errorf("Failed to list all gateways in namespace %s: %s", gateway.Namespace, err)
			return ctrl.Result{}, err
		}

		// 2. Find the oldest Gateway in the namespace, if CreateTimestamp is equal, then sort by alphabet order asc.
		validGateways := make([]*gwv1beta1.Gateway, 0)
		invalidGateways := make([]*gwv1beta1.Gateway, 0)
		for _, gw := range gatewayList.Items {
			if string(gw.Spec.GatewayClassName) == effectiveGatewayClass.Name {
				validGateways = append(validGateways, &gw)
			} else {
				invalidGateways = append(invalidGateways, &gw)
			}
		}

		sort.Slice(validGateways, func(i, j int) bool {
			if validGateways[i].CreationTimestamp.Time.Equal(validGateways[j].CreationTimestamp.Time) {
				return validGateways[i].Name < validGateways[j].Name
			} else {
				return validGateways[i].CreationTimestamp.Time.Before(validGateways[j].CreationTimestamp.Time)
			}
		})

		// 3. Set the oldest as Accepted and the rest are unaccepted
		statusChangedGateways := make([]*gwv1beta1.Gateway, 0)
		for i := range validGateways {
			if i == 0 {
				if !isAcceptedGateway(validGateways[i]) {
					r.setAccepted(validGateways[i])
					statusChangedGateways = append(statusChangedGateways, validGateways[i])
				}
			} else {
				if isAcceptedGateway(validGateways[i]) {
					r.setUnaccepted(validGateways[i])
					statusChangedGateways = append(statusChangedGateways, validGateways[i])
				}
			}
		}

		// in case of effective GatewayClass changed
		for i := range invalidGateways {
			if isAcceptedGateway(invalidGateways[i]) {
				r.setUnaccepted(invalidGateways[i])
				statusChangedGateways = append(statusChangedGateways, invalidGateways[i])
			}
		}

		// 4. update status
		for _, gw := range statusChangedGateways {
			if err := r.cfg.Client.Status().Update(ctx, gw); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *gatewayReconciler) setAccepted(gateway *gwv1beta1.Gateway) {
	metautil.SetStatusCondition(&gateway.Status.Conditions, metav1.Condition{
		Type:               string(gwv1beta1.GatewayConditionAccepted),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: gateway.Generation,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             string(gwv1beta1.GatewayReasonAccepted),
		Message:            fmt.Sprintf("Gateway %s/%s is accepted.", gateway.Namespace, gateway.Name),
	})
}

func (r *gatewayReconciler) setUnaccepted(gateway *gwv1beta1.Gateway) {
	metautil.SetStatusCondition(&gateway.Status.Conditions, metav1.Condition{
		Type:               string(gwv1beta1.GatewayConditionAccepted),
		Status:             metav1.ConditionFalse,
		ObservedGeneration: gateway.Generation,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "Unaccepted",
		Message:            fmt.Sprintf("Gateway %s/%s is not accepted as it's not the oldest one in namespace %q.", gateway.Namespace, gateway.Name, gateway.Namespace),
	})
}

func isAcceptedGateway(gateway *gwv1beta1.Gateway) bool {
	return metautil.IsStatusConditionTrue(gateway.Status.Conditions, string(gwv1beta1.GatewayConditionAccepted))
}

// SetupWithManager sets up the controller with the Manager.
func (r *gatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gwv1beta1.Gateway{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			gateway, ok := obj.(*gwv1beta1.Gateway)
			if !ok {
				klog.Errorf("unexpected object type %T", obj)
				return false
			}

			gatewayClass, err := r.cfg.K8sAPI.GatewayAPIClient.
				GatewayV1beta1().
				GatewayClasses().
				Get(context.TODO(), string(gateway.Spec.GatewayClassName), metav1.GetOptions{})
			if err != nil {
				klog.Errorf("failed to get gatewayclass %s", gateway.Spec.GatewayClassName)
				return false
			}

			if gatewayClass.Spec.ControllerName != commons.GatewayController {
				klog.Warningf("class controller of Gateway %s/%s is not %s", gateway.Namespace, gateway.Name, commons.GatewayController)
				return false
			}

			return true
		}))).
		Watches(
			&source.Kind{Type: &gwv1beta1.GatewayClass{}},
			handler.EnqueueRequestsFromMapFunc(r.gatewayClassToGateways),
			builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
				gatewayClass, ok := obj.(*gwv1beta1.GatewayClass)
				if !ok {
					klog.Errorf("unexpected object type: %T", obj)
					return false
				}

				return gatewayClass.Spec.ControllerName == commons.GatewayController
			})),
		).
		Complete(r)
}

func (r *gatewayReconciler) gatewayClassToGateways(obj client.Object) []reconcile.Request {
	gatewayClass, ok := obj.(*gwv1beta1.GatewayClass)
	if !ok {
		klog.Errorf("unexpected object type: %T", obj)
		return nil
	}

	if isEffectiveGatewayClass(gatewayClass) {
		var gateways gwv1beta1.GatewayList
		if err := r.cfg.Client.List(context.TODO(), &gateways); err != nil {
			klog.Error("error listing gateways: %s", err)
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

	return nil
}
