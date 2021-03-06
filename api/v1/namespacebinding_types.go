/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	Config = "config"
)

// NamespaceBindingSpec defines the desired state of NamespaceBinding
type NamespaceBindingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Selector is a label selector, which is used to select the namespace to be set their parent.
	Selector string `json:"selector"`

	// Parent is the parent of the selected namespaces.
	Parent string `json:"parent"`

	// Interval is the reconciler execution interval (default is 30 seconds).
	Interval *int64 `json:"interval,omitempty"`

	Exclusions []Exclusion `json:"exclusions,omitempty"`
}

type Exclusion struct {
	Value string `json:"value"`
}

// NamespaceBindingStatus defines the observed state of NamespaceBinding
type NamespaceBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	LastExecutionTime *metav1.Time `json:"lastExecutionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// NamespaceBinding is the Schema for the namespacebindings API
type NamespaceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespaceBindingSpec   `json:"spec,omitempty"`
	Status NamespaceBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NamespaceBindingList contains a list of NamespaceBinding
type NamespaceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespaceBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NamespaceBinding{}, &NamespaceBindingList{})
}
