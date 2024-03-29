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
	"github.com/flomesh-io/ErieCanal/pkg/route"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceExportRule struct {
	// The port number of service
	PortNumber int32 `json:"portNumber,omitempty"`

	// Path is matched against the path of an incoming request. Currently it can
	// contain characters disallowed from the conventional "path" part of a URL
	// as defined by RFC 3986. Paths must begin with a '/' and must be present
	// when using PathType with value "Exact" or "Prefix".
	Path string `json:"path,omitempty"`

	// PathType determines the interpretation of the Path matching. PathType can
	// be one of the following values:
	// * Exact: Matches the URL path exactly.
	// * Prefix: Matches based on a URL path prefix split by '/'. Matching is
	//   done on a path element by element basis. A path element refers is the
	//   list of labels in the path split by the '/' separator. A request is a
	//   match for path p if every p is an element-wise prefix of p of the
	//   request path. Note that if the last element of the path is a substring
	//   of the last element in request path, it is not a match (e.g. /foo/bar
	//   matches /foo/bar/baz, but does not match /foo/barbaz).

	// +kubebuilder:validation:Enum=Exact;Prefix
	PathType *networkingv1.PathType `json:"pathType"`
}

type PathRewrite struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// ServiceExportSpec defines the desired state of ServiceExport
type ServiceExportSpec struct {
	// +optional
	// PathRewrite, it shares ONE rewrite rule for the same ServiceExport
	PathRewrite *PathRewrite `json:"pathRewrite,omitempty"`

	// +optional
	// Indicates if session sticky is  enabled
	SessionSticky bool `json:"sessionSticky,omitempty"`

	// +kubebuilder:default=RoundRobinLoadBalancer
	// +kubebuilder:validation:Enum=RoundRobinLoadBalancer;HashingLoadBalancer;LeastWorkLoadBalancer
	// +optional
	// The LoadBalancer Type applied to the Ingress Rules those created by the ServiceExport
	LoadBalancer route.AlgoBalancer `json:"loadBalancer,omitempty"`

	// +kubebuilder:validation:MinItems=1
	// The paths for accessing the service via Ingress controller
	Rules []ServiceExportRule `json:"rules,omitempty"`

	// +optional
	// If empty, service is exported to all managed clusters.
	// If not empty, service is exported to specified clusters,
	//  must be in format [region]/[zone]/[group]/[cluster]
	TargetClusters []string `json:"targetClusters,omitempty"`

	// +optional
	// The ServiceAccount associated with this service
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// ServiceExportStatus defines the observed state of ServiceExport
type ServiceExportStatus struct {
	// +optional
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// ServiceExportConditionType identifies a specific condition.
type ServiceExportConditionType string

const (
	// ServiceExportValid means that the service referenced by this
	// service export has been recognized as valid by controller.
	// This will be false if the service is found to be unexportable
	// (ExternalName, not found).
	ServiceExportValid ServiceExportConditionType = "Valid"
	// ServiceExportConflict means that there is a conflict between two
	// exports for the same Service. When "True", the condition message
	// should contain enough information to diagnose the conflict:
	// field(s) under contention, which cluster won, and why.
	// Users should not expect detailed per-cluster information in the
	// conflict message.
	ServiceExportConflict ServiceExportConditionType = "Conflict"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=sexp,scope=Namespaced
// +kubebuilder:printcolumn:name="Valid",type="string",priority=0,JSONPath=".status.conditions[?(@.type=='Valid')].status"
// +kubebuilder:printcolumn:name="Conflict",type="string",priority=0,JSONPath=".status.conditions[?(@.type=='Conflict')].status"
// +kubebuilder:printcolumn:name="Age",type="date",priority=0,JSONPath=".metadata.creationTimestamp"

// ServiceExport is the Schema for the ServiceExports API
type ServiceExport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceExportSpec   `json:"spec,omitempty"`
	Status ServiceExportStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceExportList contains a list of ServiceExport
type ServiceExportList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of endpoint slices
	// +listType=set
	Items []ServiceExport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceExport{}, &ServiceExportList{})
}
