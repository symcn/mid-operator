package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SignCert struct {
	CA    []byte `json:"ca,omitempty"`
	Root  []byte `json:"root,omitempty"`
	Key   []byte `json:"key,omitempty"`
	Chain []byte `json:"chain,omitempty"`
}

type IstioService struct {
	Name          string               `json:"name"`
	LabelSelector string               `json:"labelSelector,omitempty"`
	IPs           []string             `json:"podIPs,omitempty"`
	Ports         []corev1.ServicePort `json:"ports,omitempty"`
}

// RemoteIstioSpec defines the desired state of RemoteIstio
type RemoteIstioSpec struct {
	// IncludeIPRanges the range where to capture egress traffic
	IncludeIPRanges string `json:"includeIPRanges,omitempty"`

	// ExcludeIPRanges the range where not to capture egress traffic
	ExcludeIPRanges string `json:"excludeIPRanges,omitempty"`

	// EnabledServices the Istio component services replicated to remote side
	EnabledServices []IstioService `json:"enabledServices"`

	// List of namespaces to label with sidecar auto injection enabled
	AutoInjectionNamespaces []string `json:"autoInjectionNamespaces,omitempty"`

	// DefaultResources are applied for all Istio components by default, can be overridden for each component
	DefaultResources *corev1.ResourceRequirements `json:"defaultResources,omitempty"`

	// SidecarInjector configuration options
	SidecarInjector SidecarInjectorConfiguration `json:"sidecarInjector,omitempty"`

	// Proxy configuration options
	Proxy ProxyConfiguration `json:"proxy,omitempty"`

	// Proxy Init configuration options
	ProxyInit ProxyInitConfiguration `json:"proxyInit,omitempty"`

	SignCert SignCert `json:"signCert,omitempty"`
}

// RemoteIstioStatus defines the observed state of RemoteIstio
type RemoteIstioStatus struct {
	Status         ConfigState `json:"Status,omitempty"`
	GatewayAddress []string    `json:"GatewayAddress,omitempty"`
	ErrorMessage   string      `json:"ErrorMessage,omitempty"`
}

// +kubebuilder:object:root=true

// RemoteIstio is the Schema for the remoteistios API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.Status",description="Status of the resource"
// +kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.ErrorMessage",description="Error message"
// +kubebuilder:printcolumn:name="Ingress IPs",type="string",JSONPath=".status.GatewayAddress",description="Ingress gateway addresses of the resource"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RemoteIstio struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RemoteIstioSpec   `json:"spec,omitempty"`
	Status RemoteIstioStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RemoteIstioList contains a list of RemoteIstio
type RemoteIstioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteIstio `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RemoteIstio{}, &RemoteIstioList{})
}
