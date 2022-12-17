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
	"fmt"
	nsigv1alpha1 "github.com/flomesh-io/ErieCanal/apis/namespacedingress/v1alpha1"
	"github.com/flomesh-io/ErieCanal/pkg/certificate"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/helm"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/flomesh-io/ErieCanal/pkg/repo"
	"github.com/flomesh-io/ErieCanal/pkg/util/tls"
	ghodssyaml "github.com/ghodss/yaml"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/strvals"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var (
	//go:embed chart.tgz
	chartSource []byte
)

// NamespacedIngressReconciler reconciles a NamespacedIngress object
type NamespacedIngressReconciler struct {
	client.Client
	K8sAPI                  *kube.K8sAPI
	Scheme                  *runtime.Scheme
	Recorder                record.EventRecorder
	ControlPlaneConfigStore *config.Store
	CertMgr                 certificate.Manager
}

type namespacedIngressValues struct {
	NamespacedIngress *nsigv1alpha1.NamespacedIngress `json:"nsig,omitempty"`
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the NamespacedIngress closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NamespacedIngress object against the actual NamespacedIngress state, and then
// perform operations to make the NamespacedIngress state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *NamespacedIngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	mc := r.ControlPlaneConfigStore.MeshConfig.GetConfig()

	klog.Infof("[NSIG] Ingress Enabled = %t, Namespaced Ingress = %t", mc.Ingress.Enabled, mc.Ingress.Namespaced)
	if !mc.Ingress.Enabled || !mc.Ingress.Namespaced {
		klog.Warning("Ingress is not enabled or Ingress mode is not Namespace, ignore processing NamespacedIngress...")
		return ctrl.Result{}, nil
	}

	nsig := &nsigv1alpha1.NamespacedIngress{}
	if err := r.Get(
		ctx,
		client.ObjectKey{Name: req.Name, Namespace: req.Namespace},
		nsig,
	); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			klog.V(3).Info("[NSIG] NamespacedIngress resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		klog.Errorf("Failed to get NamespacedIngress, %#v", err)
		return ctrl.Result{}, err
	}

	ctrlResult, err := r.deriveCodebases(nsig, mc)
	if err != nil {
		return ctrlResult, err
	}

	ctrlResult, err = r.updateConfig(nsig, mc)
	if err != nil {
		return ctrlResult, err
	}

	releaseName := fmt.Sprintf("namespaced-ingress-%s", nsig.Namespace)
	if ctrlResult, err = helm.RenderChart(releaseName, nsig, chartSource, mc, r.Client, r.Scheme, resolveValues); err != nil {
		return ctrlResult, err
	}

	return ctrl.Result{}, nil
}

func resolveValues(object metav1.Object, mc *config.MeshConfig) (map[string]interface{}, error) {
	nsig, ok := object.(*nsigv1alpha1.NamespacedIngress)
	if !ok {
		return nil, fmt.Errorf("object %v is not type of nsigv1alpha1.NamespacedIngress", object)
	}

	klog.V(5).Infof("[NSIG] Resolving Values ...")

	nsigBytes, err := ghodssyaml.Marshal(&namespacedIngressValues{NamespacedIngress: nsig})
	if err != nil {
		return nil, fmt.Errorf("convert NamespacedIngress to yaml, err = %#v", err)
	}
	klog.V(5).Infof("\n\nNSIG VALUES YAML:\n\n\n%s\n\n", string(nsigBytes))
	nsigValues, err := chartutil.ReadValues(nsigBytes)
	if err != nil {
		return nil, err
	}

	finalValues := nsigValues.AsMap()

	overrides := []string{
		"ec.ingress.namespaced=true",
		fmt.Sprintf("ec.image.repository=%s", mc.Images.Repository),
		fmt.Sprintf("ec.namespace=%s", config.GetErieCanalNamespace()),
	}

	for _, ov := range overrides {
		if err := strvals.ParseInto(ov, finalValues); err != nil {
			return nil, err
		}
	}

	return finalValues, nil
}

func (r *NamespacedIngressReconciler) deriveCodebases(nsig *nsigv1alpha1.NamespacedIngress, mc *config.MeshConfig) (ctrl.Result, error) {
	repoClient := repo.NewRepoClient(mc.RepoRootURL())

	ingressPath := mc.NamespacedIngressCodebasePath(nsig.Namespace)
	parentPath := mc.IngressCodebasePath()
	if err := repoClient.DeriveCodebase(ingressPath, parentPath); err != nil {
		return ctrl.Result{RequeueAfter: 1 * time.Second}, err
	}

	return ctrl.Result{}, nil
}

func (r *NamespacedIngressReconciler) updateConfig(nsig *nsigv1alpha1.NamespacedIngress, mc *config.MeshConfig) (ctrl.Result, error) {
	if mc.Ingress.Namespaced && nsig.Spec.TLS.Enabled {
		repoClient := repo.NewRepoClient(mc.RepoRootURL())
		basepath := mc.NamespacedIngressCodebasePath(nsig.Namespace)

		if nsig.Spec.TLS.SSLPassthrough.Enabled {
			// SSL passthrough
			err := tls.UpdateSSLPassthrough(
				basepath,
				repoClient,
				nsig.Spec.TLS.SSLPassthrough.Enabled,
				nsig.Spec.TLS.SSLPassthrough.UpstreamPort,
			)
			if err != nil {
				return ctrl.Result{RequeueAfter: 1 * time.Second}, err
			}
		} else {
			// TLS offload
			err := tls.IssueCertForIngress(basepath, repoClient, r.CertMgr, mc)
			if err != nil {
				return ctrl.Result{RequeueAfter: 1 * time.Second}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespacedIngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nsigv1alpha1.NamespacedIngress{}).
		Owns(&corev1.Service{}).
		Owns(&appv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}
