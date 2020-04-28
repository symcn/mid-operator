package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GatewayType string

const (
	GatewayTypeIngress GatewayType = "ingress"
	GatewayTypeEgress  GatewayType = "egress"
)

// MeshGatewaySpec defines the desired state of MeshGateway
type MeshGatewaySpec struct {
	MeshGatewayConfiguration `json:",inline"`
	// +kubebuilder:validation:MinItems=1
	Ports []corev1.ServicePort `json:"ports"`
	Type  GatewayType          `json:"type"`
}

// MeshGatewayStatus defines the observed state of MeshGateway
type MeshGatewayStatus struct {
	Status         ConfigState `json:"Status,omitempty"`
	GatewayAddress []string    `json:"GatewayAddress,omitempty"`
	ErrorMessage   string      `json:"ErrorMessage,omitempty"`
}

// +kubebuilder:object:root=true

// MeshGateway is the Schema for the meshgateways API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="Type of the gateway"
// +kubebuilder:printcolumn:name="Service Type",type="string",JSONPath=".spec.serviceType",description="Type of the service"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.Status",description="Status of the resource"
// +kubebuilder:printcolumn:name="Ingress IPs",type="string",JSONPath=".status.GatewayAddress",description="Ingress gateway addresses of the resource"
// +kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.ErrorMessage",description="Error message"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:path=meshgateways,shortName=mgw
type MeshGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeshGatewaySpec   `json:"spec,omitempty"`
	Status MeshGatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MeshGatewayList contains a list of MeshGateway
type MeshGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeshGateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MeshGateway{}, &MeshGatewayList{})
}
