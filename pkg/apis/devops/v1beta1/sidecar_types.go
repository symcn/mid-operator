/*
Copyright 2020 The symcn authors.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SidecarSpec defines the desired state of Sidecar
type SidecarSpec struct {
	Foo string `json:"foo,omitempty"`
}

// SidecarStatus defines the observed state of Sidecar
type SidecarStatus struct {
}

// +kubebuilder:object:root=true

// Sidecar is the Schema for the sidecars API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=sd
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. "
type Sidecar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SidecarSpec   `json:"spec,omitempty"`
	Status SidecarStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SidecarList contains a list of Sidecar
type SidecarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sidecar `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sidecar{}, &SidecarList{})
}
