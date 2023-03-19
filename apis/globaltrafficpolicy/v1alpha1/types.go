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

type LoadBalancerType string

const (
	ActiveActiveLbType LoadBalancerType = "ActiveActive"
	LocalityLbType     LoadBalancerType = "Locality"
	FailOverLbType     LoadBalancerType = "FailOver"
)

type TrafficTarget struct {
	// Format: [region]/[zone]/[group]/[cluster]
	ClusterKey string `json:"clusterKey"`

	// +optional
	Weight *int `json:"weight,omitempty"`
}

// GlobalTrafficPolicySpec defines the desired state of GlobalTrafficPolicy
type GlobalTrafficPolicySpec struct {
	// +kubebuilder:default=Locality
	// +kubebuilder:validation:Enum=Locality;ActiveActive;FailOver
	// Type of global load distribution
	LbType LoadBalancerType `json:"lbType"`

	// +optional
	Targets []TrafficTarget `json:"targets,omitempty"`
}

// GlobalTrafficPolicyStatus defines the observed state of GlobalTrafficPolicy
type GlobalTrafficPolicyStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=gtp,scope=Namespaced
// +kubebuilder:printcolumn:name="LB Type",type="string",priority=0,JSONPath=".spec.lbType"
// +kubebuilder:printcolumn:name="Age",type="date",priority=0,JSONPath=".metadata.creationTimestamp"

// GlobalTrafficPolicy is the Schema for the GlobalTrafficPolicys API
type GlobalTrafficPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalTrafficPolicySpec   `json:"spec,omitempty"`
	Status GlobalTrafficPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GlobalTrafficPolicyList contains a list of GlobalTrafficPolicy
type GlobalTrafficPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalTrafficPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GlobalTrafficPolicy{}, &GlobalTrafficPolicyList{})
}
