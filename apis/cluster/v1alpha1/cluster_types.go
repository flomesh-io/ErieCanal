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
	"github.com/flomesh-io/ErieCanal/pkg/commons"
	"github.com/flomesh-io/ErieCanal/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// +kubebuilder:default=false
	// +optional

	// IsInCluster, it indicates the Cluster resource is used for connecting local cluster
	//  or a remote cluster.
	IsInCluster bool `json:"isInCluster"`

	// +kubebuilder:default=default
	// +optional

	// Region, the locality information of this cluster
	Region string `json:"region,omitempty"`

	// +kubebuilder:default=default
	// +optional

	// Zone, the locality information of this cluster
	Zone string `json:"zone,omitempty"`

	// +kubebuilder:default=default
	// +optional

	// Group, the locality information of this cluster
	Group string `json:"group,omitempty"`

	// +optional

	// GatewayHost, the Full Qualified Domain Name or IP of the gateway/ingress of this cluster
	// If it's an IP address, only IPv4 is supported
	GatewayHost string `json:"gatewayHost,omitempty"`

	// +kubebuilder:default=80
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +optional

	// The port number of the gateway
	GatewayPort int32 `json:"gatewayPort,omitempty"`

	// FIXME: temp solution, should NOT store this as plain text.
	//  consider use cli to add cluster to control plane, import kubeconfig
	//  and create a Secret with proper SA to store it as bytes

	// +optional

	// Kubeconfig, The kubeconfig of the cluster you want to connnect to
	// This's not needed if ClusterMode is InCluster, it will use InCluster
	// config
	Kubeconfig string `json:"kubeconfig,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// +optional
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// ClusterConditionType identifies a specific condition.
type ClusterConditionType string

const (
	// ClusterManaged means that the cluster has joined the CLusterSet successfully
	//  and is managed by Control Plane.
	ClusterManaged ClusterConditionType = "Managed"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Region",type="string",priority=0,JSONPath=".spec.region"
// +kubebuilder:printcolumn:name="Zone",type="string",priority=0,JSONPath=".spec.zone"
// +kubebuilder:printcolumn:name="Group",type="string",priority=0,JSONPath=".spec.group"
// +kubebuilder:printcolumn:name="Gateway Host",type="string",priority=0,JSONPath=".spec.gatewayHost"
// +kubebuilder:printcolumn:name="Gateway Port",type="integer",priority=0,JSONPath=".spec.gatewayPort"
// +kubebuilder:printcolumn:name="Managed",type="string",priority=0,JSONPath=".status.conditions[?(@.type=='Managed')].status"
// +kubebuilder:printcolumn:name="Managed Age",type="date",priority=0,JSONPath=".status.conditions[?(@.type=='Managed')].lastTransitionTime"
// +kubebuilder:printcolumn:name="Age",type="date",priority=0,JSONPath=".metadata.creationTimestamp"

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}

func (c *Cluster) Key() string {
	return util.EvaluateTemplate(commons.ClusterIDTemplate, struct {
		Region  string
		Zone    string
		Group   string
		Cluster string
	}{
		Region:  c.Spec.Region,
		Zone:    c.Spec.Zone,
		Group:   c.Spec.Group,
		Cluster: c.Name,
	})
}
