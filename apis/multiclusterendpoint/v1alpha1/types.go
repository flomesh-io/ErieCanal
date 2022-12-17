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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// MultiClusterEndpointSpec defines the desired state of MultiClusterEndpoint
type MultiClusterEndpointSpec struct {
	ClusterKey string `json:"clusterKey"`

	Target Target `json:"target"`
}

type Target struct {
	Host string `json:"host"`
	IP   string `json:"ip"`
	Port int32  `json:"port"`
	Path string `json:"path"`
}

// MultiClusterEndpointStatus defines the observed state of MultiClusterEndpoint
type MultiClusterEndpointStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Age",type="date",priority=0,JSONPath=".metadata.creationTimestamp"

// MultiClusterEndpoint is the Schema for the MultiClusterEndpoints API
type MultiClusterEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MultiClusterEndpointSpec   `json:"spec,omitempty"`
	Status MultiClusterEndpointStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MultiClusterEndpointList contains a list of MultiClusterEndpoint
type MultiClusterEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MultiClusterEndpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MultiClusterEndpoint{}, &MultiClusterEndpointList{})
}
