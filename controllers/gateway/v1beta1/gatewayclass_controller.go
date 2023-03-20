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
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	gwv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sort"
	"time"
)

const (
	GatewayClassConditionStatusActive gwv1beta1.GatewayClassConditionType   = "Active"
	GatewayClassReasonActive          gwv1beta1.GatewayClassConditionReason = "Active"
	GatewayClassReasonInactive        gwv1beta1.GatewayClassConditionReason = "Inactive"
)

type gatewayClassReconciler struct {
	recorder record.EventRecorder
	cfg      *controllers.ReconcilerConfig
}

func NewGatewayClassReconciler(rc *controllers.ReconcilerConfig) controllers.Reconciler {
	return &gatewayClassReconciler{
		recorder: rc.Manager.GetEventRecorderFor("GatewayClass"),
		cfg:      rc,
	}
}

func (r *gatewayClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	gatewayClass := &gwv1beta1.GatewayClass{}
	if err := r.cfg.Client.Get(
		ctx,
		client.ObjectKey{Name: req.Name},
		gatewayClass,
	); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			klog.V(3).Info("GatewayClass resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		klog.Errorf("Failed to get GatewayClass, %#v", err)
		return ctrl.Result{}, err
	}

	gatewayClassList, err := r.cfg.K8sAPI.GatewayAPIClient.GatewayV1beta1().
		GatewayClasses().
		List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("failed list gatewayclasses: %s", err)
		return ctrl.Result{}, err
	}

	// Accept all GatewayClasses those ControllerName is flomesh.io/gateway-controller
	r.setAcceptedStatus(gatewayClassList, gatewayClass)
	if err := r.cfg.Client.Status().Update(ctx, gatewayClass); err != nil {
		return ctrl.Result{}, err
	}

	// If there's multiple GatewayClasses, the oldest is set to active and the rest are set to inactive
	for _, class := range r.setActiveStatus(gatewayClassList) {
		if err := r.cfg.Client.Status().Update(ctx, class); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *gatewayClassReconciler) setAcceptedStatus(list *gwv1beta1.GatewayClassList, gatewayClass *gwv1beta1.GatewayClass) {
	if gatewayClass.Spec.ControllerName == commons.GatewayController {
		r.setAccepted(gatewayClass)
	} else {
		r.setRejected(gatewayClass)
	}
}

func (r *gatewayClassReconciler) setActiveStatus(list *gwv1beta1.GatewayClassList) []*gwv1beta1.GatewayClass {
	acceptedClasses := make([]*gwv1beta1.GatewayClass, 0)
	for _, class := range list.Items {
		if isAcceptedGatewayClass(&class) {
			acceptedClasses = append(acceptedClasses, &class)
		}
	}

	sort.Slice(acceptedClasses, func(i, j int) bool {
		if acceptedClasses[i].CreationTimestamp.Time.Equal(acceptedClasses[j].CreationTimestamp.Time) {
			return acceptedClasses[i].Name < acceptedClasses[j].Name
		} else {
			return acceptedClasses[i].CreationTimestamp.Time.Before(acceptedClasses[j].CreationTimestamp.Time)
		}
	})

	statusChangedClasses := make([]*gwv1beta1.GatewayClass, 0)
	for i, class := range acceptedClasses {
		// ONLY the oldest GatewayClass is active
		if i == 0 {
			if !isActiveGatewayClass(class) {
				r.setActive(acceptedClasses[i])
				statusChangedClasses = append(statusChangedClasses, acceptedClasses[i])
			}
			continue
		}

		if isActiveGatewayClass(class) {
			r.setInactive(acceptedClasses[i])
			statusChangedClasses = append(statusChangedClasses, acceptedClasses[i])
		}
	}

	return statusChangedClasses
}

func (r *gatewayClassReconciler) setRejected(gatewayClass *gwv1beta1.GatewayClass) {
	metautil.SetStatusCondition(&gatewayClass.Status.Conditions, metav1.Condition{
		Type:               string(gwv1beta1.GatewayClassConditionStatusAccepted),
		Status:             metav1.ConditionFalse,
		ObservedGeneration: gatewayClass.Generation,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "Rejected",
		Message:            fmt.Sprintf("GatewayClass %q is rejected as ControllerName %q is not supported.", gatewayClass.Name, gatewayClass.Spec.ControllerName),
	})
}

func (r *gatewayClassReconciler) setAccepted(gatewayClass *gwv1beta1.GatewayClass) {
	metautil.SetStatusCondition(&gatewayClass.Status.Conditions, metav1.Condition{
		Type:               string(gwv1beta1.GatewayClassConditionStatusAccepted),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: gatewayClass.Generation,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             string(gwv1beta1.GatewayClassReasonAccepted),
		Message:            fmt.Sprintf("GatewayClass %q is accepted.", gatewayClass.Name),
	})
}

func (r *gatewayClassReconciler) setActive(gatewayClass *gwv1beta1.GatewayClass) {
	metautil.SetStatusCondition(&gatewayClass.Status.Conditions, metav1.Condition{
		Type:               string(GatewayClassConditionStatusActive),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: gatewayClass.Generation,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             string(GatewayClassReasonActive),
		Message:            fmt.Sprintf("GatewayClass %q is set to active.", gatewayClass.Name),
	})
}
func (r *gatewayClassReconciler) setInactive(gatewayClass *gwv1beta1.GatewayClass) {
	metautil.SetStatusCondition(&gatewayClass.Status.Conditions, metav1.Condition{
		Type:               string(GatewayClassConditionStatusActive),
		Status:             metav1.ConditionFalse,
		ObservedGeneration: gatewayClass.Generation,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             string(GatewayClassReasonInactive),
		Message:            fmt.Sprintf("GatewayClass %q is inactive as there's already an active GatewayClass.", gatewayClass.Name),
	})
}

func isAcceptedGatewayClass(gatewayClass *gwv1beta1.GatewayClass) bool {
	return metautil.IsStatusConditionTrue(gatewayClass.Status.Conditions, string(gwv1beta1.GatewayClassConditionStatusAccepted))
}

func isActiveGatewayClass(gatewayClass *gwv1beta1.GatewayClass) bool {
	return metautil.IsStatusConditionTrue(gatewayClass.Status.Conditions, string(GatewayClassConditionStatusActive))
}

func isEffectiveGatewayClass(gatewayClass *gwv1beta1.GatewayClass) bool {
	return isAcceptedGatewayClass(gatewayClass) && isActiveGatewayClass(gatewayClass)
}

// SetupWithManager sets up the controller with the Manager.
func (r *gatewayClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	gwclsPrct := predicate.NewPredicateFuncs(func(object client.Object) bool {
		gatewayClass, ok := object.(*gwv1beta1.GatewayClass)
		if !ok {
			klog.Infof("unexpected object type: %T", object)
			return false
		}

		return gatewayClass.Spec.ControllerName == commons.GatewayController
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&gwv1beta1.GatewayClass{}, builder.WithPredicates(gwclsPrct)).
		Complete(r)
}
