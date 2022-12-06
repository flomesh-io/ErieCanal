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

package cluster

import (
	"context"
	"fmt"
	clusterv1alpha1 "github.com/flomesh-io/ErieCanal/apis/cluster/v1alpha1"
	flomeshadmission "github.com/flomesh-io/ErieCanal/pkg/admission"
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/config"
	"github.com/flomesh-io/ErieCanal/pkg/kube"
	"github.com/pkg/errors"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	"net"
)

const (
	kind      = "Cluster"
	groups    = "flomesh.io"
	resources = "clusters"
	versions  = "v1alpha1"

	mwPath = commons.ClusterMutatingWebhookPath
	mwName = "mcluster.kb.flomesh.io"
	vwPath = commons.ClusterValidatingWebhookPath
	vwName = "vcluster.kb.flomesh.io"
)

func RegisterWebhooks(webhookSvcNs, webhookSvcName string, caBundle []byte) {
	rule := flomeshadmission.NewRule(
		[]admissionregv1.OperationType{admissionregv1.Create, admissionregv1.Update},
		[]string{groups},
		[]string{versions},
		[]string{resources},
	)

	mutatingWebhook := flomeshadmission.NewMutatingWebhook(
		mwName,
		webhookSvcNs,
		webhookSvcName,
		mwPath,
		caBundle,
		nil,
		[]admissionregv1.RuleWithOperations{rule},
	)

	validatingWebhook := flomeshadmission.NewValidatingWebhook(
		vwName,
		webhookSvcNs,
		webhookSvcName,
		vwPath,
		caBundle,
		nil,
		[]admissionregv1.RuleWithOperations{rule},
	)

	flomeshadmission.RegisterMutatingWebhook(mwName, mutatingWebhook)
	flomeshadmission.RegisterValidatingWebhook(vwName, validatingWebhook)
}

type ClusterDefaulter struct {
	k8sAPI      *kube.K8sAPI
	configStore *config.Store
}

func NewDefaulter(k8sAPI *kube.K8sAPI, configStore *config.Store) *ClusterDefaulter {
	return &ClusterDefaulter{
		k8sAPI:      k8sAPI,
		configStore: configStore,
	}
}

func (w *ClusterDefaulter) RuntimeObject() runtime.Object {
	return &clusterv1alpha1.Cluster{}
}

func (w *ClusterDefaulter) SetDefaults(obj interface{}) {
	c, ok := obj.(*clusterv1alpha1.Cluster)
	if !ok {
		return
	}

	klog.V(5).Infof("Default Webhook, name=%s", c.Name)
	klog.V(4).Infof("Before setting default values, spec=%#v", c.Spec)

	meshConfig := w.configStore.MeshConfig.GetConfig()

	if meshConfig == nil {
		return
	}

	// for InCluster connector, it's name is always 'local'
	//if c.Spec.IsInCluster {
	//	c.Name = "local"
	//}
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}

	if c.Spec.IsInCluster {
		c.Labels[commons.MultiClustersConnectorMode] = "local"
	} else {
		c.Labels[commons.MultiClustersConnectorMode] = "remote"
	}

	klog.V(4).Infof("After setting default values, spec=%#v", c.Spec)
}

type ClusterValidator struct {
	k8sAPI *kube.K8sAPI
}

func (w *ClusterValidator) RuntimeObject() runtime.Object {
	return &clusterv1alpha1.Cluster{}
}

func (w *ClusterValidator) ValidateCreate(obj interface{}) error {
	cluster, ok := obj.(*clusterv1alpha1.Cluster)
	if !ok {
		return nil
	}

	if cluster.Spec.IsInCluster {
		// There can be ONLY ONE Cluster of InCluster mode
		clusterList, err := w.k8sAPI.FlomeshClient.
			ClusterV1alpha1().
			Clusters().
			List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to list Clusters, %#v", err)
			return err
		}

		numOfInCluster := 0
		for _, c := range clusterList.Items {
			if c.Spec.IsInCluster {
				numOfInCluster++
			}
		}
		if numOfInCluster >= 1 {
			errMsg := fmt.Sprintf("there're %d InCluster resources, should ONLY have exact ONE", numOfInCluster)
			klog.Errorf(errMsg)
			return errors.New(errMsg)
		}
	}

	return doValidation(obj)
}

func (w *ClusterValidator) ValidateUpdate(oldObj, obj interface{}) error {
	oldCluster, ok := oldObj.(*clusterv1alpha1.Cluster)
	if !ok {
		return nil
	}

	cluster, ok := obj.(*clusterv1alpha1.Cluster)
	if !ok {
		return nil
	}

	if oldCluster.Spec.IsInCluster != cluster.Spec.IsInCluster {
		return errors.New("cannot update an immutable field: spec.IsInCluster")
	}

	return doValidation(obj)
}

func (w *ClusterValidator) ValidateDelete(obj interface{}) error {
	return nil
}

func NewValidator(k8sAPI *kube.K8sAPI) *ClusterValidator {
	return &ClusterValidator{
		k8sAPI: k8sAPI,
	}
}

func doValidation(obj interface{}) error {
	c, ok := obj.(*clusterv1alpha1.Cluster)
	if !ok {
		return nil
	}

	if c.Labels == nil || c.Labels[commons.MultiClustersConnectorMode] == "" {
		return fmt.Errorf("missing required label 'multicluster.flomesh.io/connector-mode'")
	}

	connectorMode := c.Labels[commons.MultiClustersConnectorMode]
	switch connectorMode {
	case "local", "remote":
		klog.V(5).Infof("multicluster.flomesh.io/connector-mode=%s", connectorMode)
	default:
		return fmt.Errorf("invalid value %q for label multicluster.flomesh.io/connector-mode, must be either 'local' or 'remote'", connectorMode)
	}

	if c.Spec.IsInCluster {
		if connectorMode == "remote" {
			return fmt.Errorf("label and spec doesn't match: multicluster.flomesh.io/connector-mode=remote, spec.IsInCluster=true")
		}

		return nil
	} else {
		if connectorMode == "local" {
			return fmt.Errorf("label and spec doesn't match: multicluster.flomesh.io/connector-mode=local, spec.IsInCluster=false")
		}

		host := c.Spec.GatewayHost
		if host == "" {
			return errors.New("GatewayHost is required in OutCluster mode")
		}

		if c.Spec.Kubeconfig == "" {
			return fmt.Errorf("kubeconfig must be set in OutCluster mode")
		}

		//if c.Name == "local" {
		//	return errors.New("Cluster Name 'local' is reserved for InCluster Mode ONLY, please change the cluster name")
		//}

		isDNSName := false
		if ipErrs := validation.IsValidIPv4Address(field.NewPath(""), host); len(ipErrs) > 0 {
			// Not IPv4 address
			klog.Warningf("%q is NOT a valid IPv4 address: %v", host, ipErrs)
			if dnsErrs := validation.IsDNS1123Subdomain(host); len(dnsErrs) > 0 {
				// Not valid DNS domain name
				return fmt.Errorf("invalid DNS name %q: %v", host, dnsErrs)
			} else {
				// is DNS name
				isDNSName = true
			}
		}

		var gwIPv4 net.IP
		if isDNSName {
			ipAddr, err := net.ResolveIPAddr("ip4", host)
			if err != nil {
				return fmt.Errorf("%q cannot be resolved to IP", host)
			}
			klog.Infof("%q is resolved to IP: %s", host, ipAddr.IP)
			gwIPv4 = ipAddr.IP.To4()
		} else {
			gwIPv4 = net.ParseIP(host).To4()
		}

		if gwIPv4 == nil {
			return fmt.Errorf("%q cannot be resolved to a IPv4 address", host)
		}

		if gwIPv4 != nil && (gwIPv4.IsLoopback() || gwIPv4.IsUnspecified()) {
			return fmt.Errorf("gateway Host %s is resolved to Loopback IP or Unspecified", host)
		}

		port := int(c.Spec.GatewayPort)
		if errs := validation.IsValidPortNum(port); len(errs) > 0 {
			return fmt.Errorf("invalid port number %d: %v", c.Spec.GatewayPort, errs)
		}
	}

	return nil
}
