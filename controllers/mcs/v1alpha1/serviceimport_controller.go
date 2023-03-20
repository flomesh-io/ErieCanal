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
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

// serviceImportReconciler reconciles a ServiceImport object
type serviceImportReconciler struct {
	recorder record.EventRecorder
	cfg      *controllers.ReconcilerConfig
}

func NewServiceImportReconciler(rc *controllers.ReconcilerConfig) controllers.Reconciler {
	return &serviceImportReconciler{
		recorder: rc.Manager.GetEventRecorderFor("ServiceImport"),
		cfg:      rc,
	}
}

func (r *serviceImportReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	svcImport := &svcimpv1alpha1.ServiceImport{}
	if err := r.cfg.Client.Get(
		ctx,
		req.NamespacedName,
		svcImport,
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

	if shouldIgnoreImport(svcImport) {
		return ctrl.Result{}, nil
	}

	// Ensure the existence of the derived service
	if svcImport.Annotations[commons.MultiClusterDerivedServiceAnnotation] == "" {
		if svcImport.Annotations == nil {
			svcImport.Annotations = make(map[string]string)
		}

		svcImport.Annotations[commons.MultiClusterDerivedServiceAnnotation] = req.Name
		if err := r.cfg.Client.Update(ctx, svcImport); err != nil {
			return ctrl.Result{}, err
		}
		klog.Infof("Added annotation %s=%s", commons.MultiClusterDerivedServiceAnnotation, req.Name)

		return ctrl.Result{}, nil
	}

	svc, err := r.createOrGetDerivedService(ctx, svcImport)
	if err != nil {
		return ctrl.Result{}, err
	}

	if len(svcImport.Spec.IPs) == 0 {
		return ctrl.Result{}, nil
	}

	// update LoadBalancer status with provided ClusterSetIPs
	ingress := make([]corev1.LoadBalancerIngress, 0)
	for _, ip := range svcImport.Spec.IPs {
		ingress = append(ingress, corev1.LoadBalancerIngress{
			IP: ip,
		})
	}

	svc.Status = corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: ingress,
		},
	}

	if err := r.cfg.Client.Status().Update(ctx, svc); err != nil {
		return ctrl.Result{}, err
	}

	// TODO: create/update/delete EndpointSlice and add required annotations/labels
	//for _, p := range svcImport.Spec.Ports {
	//    for _, ep := range p.Endpoints {
	//        ep.Target.
	//    }
	//}

	return ctrl.Result{}, nil
}

func (r *serviceImportReconciler) createOrGetDerivedService(ctx context.Context, svcImport *svcimpv1alpha1.ServiceImport) (*corev1.Service, error) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: svcImport.Namespace,
			Name:      svcImport.Name,
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:       svcImport.Name,
					Kind:       serviceImportKind,
					APIVersion: svcimpv1alpha1.SchemeGroupVersion.String(),
					UID:        svcImport.UID,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeClusterIP,
			Ports: servicePorts(svcImport),
		},
	}

	// just create to avoid concurrent write
	if err := r.cfg.Client.Create(ctx, svc); err != nil {
		if errors.IsAlreadyExists(err) {
			if err = r.cfg.Client.Get(
				ctx,
				types.NamespacedName{Namespace: svcImport.Namespace, Name: svcImport.Name},
				svc,
			); err != nil {
				return nil, err
			}

			if isAlreadyOwnerOfService(svcImport, svc.OwnerReferences) {
				return svc, nil
			}

			svc.OwnerReferences = append(svc.OwnerReferences, metav1.OwnerReference{
				Name:       svcImport.Name,
				Kind:       serviceImportKind,
				APIVersion: svcimpv1alpha1.SchemeGroupVersion.String(),
				UID:        svcImport.UID,
			})

			if err = r.cfg.Client.Update(ctx, svc); err != nil {
				return nil, err
			}

			return svc, nil
		}

		return nil, err
	}

	klog.Infof("Created service %s/%s", svc.Namespace, svc.Name)

	return svc, nil
}

func servicePorts(svcImport *svcimpv1alpha1.ServiceImport) []corev1.ServicePort {
	ports := make([]corev1.ServicePort, len(svcImport.Spec.Ports))
	for i, p := range svcImport.Spec.Ports {
		ports[i] = corev1.ServicePort{
			Name:        p.Name,
			Protocol:    p.Protocol,
			Port:        p.Port,
			AppProtocol: p.AppProtocol,
		}
	}
	return ports
}

func shouldIgnoreImport(svcImport *svcimpv1alpha1.ServiceImport) bool {
	if svcImport.DeletionTimestamp != nil {
		return true
	}

	if svcImport.Spec.Type != svcimpv1alpha1.ClusterSetIP {
		return true
	}

	return false
}

func isAlreadyOwnerOfService(svcImport *svcimpv1alpha1.ServiceImport, svcOwnerRefs []metav1.OwnerReference) bool {
	for _, ref := range svcOwnerRefs {
		if ref.APIVersion == svcimpv1alpha1.SchemeGroupVersion.String() && ref.Kind == serviceImportKind {
			return ref.Name == svcImport.Name
		}
	}

	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *serviceImportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&svcimpv1alpha1.ServiceImport{}).
		Complete(r)
}
