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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Comma-separated minimum per-scope logging level of messages to output, in the form of <scope>:<level>,<scope>:<level>
// The control plane has different scopes depending on component, but can configure default log level across all components
// If empty, default scope and level will be used as configured in code
type LoggingConfiguration struct {
	// +kubebuilder:validation:Pattern="^([a-zA-Z]+:[a-zA-Z]+,?)+$"
	Level *string `json:"level,omitempty"`
}

// MeshPolicyConfiguration configures the default MeshPolicy resource
type MeshPolicyConfiguration struct {
	// MTLSMode sets the mesh-wide mTLS policy
	// +kubebuilder:validation:Enum="STRICT,PERMISSIVE,DISABLED"
	MTLSMode MTLSMode `json:"mtlsMode,omitempty"`
}

// IstiodConfiguration defines config options for Istiod
type IstiodConfiguration struct {
	Enabled             *bool `json:"enabled,omitempty"`
	MultiClusterSupport *bool `json:"multiClusterSupport,omitempty"`
}

// PilotConfiguration defines config options for Pilot
type PilotConfiguration struct {
	Enabled                             *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithHPA `json:",inline"`
	Sidecar                             *bool `json:"sidecar,omitempty"`
	TraceSampling                       int64 `json:"traceSampling,omitempty"`
	// If enabled, protocol sniffing will be used for outbound listeners whose port protocol is not specified or unsupported
	EnableProtocolSniffingOutbound *bool `json:"enableProtocolSniffingOutbound,omitempty"`
	// If enabled, protocol sniffing will be used for inbound listeners whose port protocol is not specified or unsupported
	EnableProtocolSniffingInbound *bool `json:"enableProtocolSniffingInbound,omitempty"`
	// Configure the certificate provider for control plane communication.
	// Currently, two providers are supported: "kubernetes" and "istiod".
	// As some platforms may not have kubernetes signing APIs,
	// Istiod is the default
	// +kubebuilder:validation:Enum="kubernetes,istiod"
	CertProvider PilotCertProviderType `json:"certProvider,omitempty"`

	// If present will be appended at the end of the initial/preconfigured container arguments
	AdditionalContainerArgs []string `json:"additionalContainerArgs,omitempty"`

	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`
}

// CertificateConfig configures DNS certificates provisioned through Chiron linked into Pilot
type CertificateConfig struct {
	SecretName *string  `json:"secretName,omitempty"`
	DNSNames   []string `json:"dnsNames,omitempty"`
}

// SidecarInjectorInitConfiguration defines options for init containers in the sidecar
type SidecarInjectorInitConfiguration struct {
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// CNIRepairConfiguration defines config for the repair CNI container
type CNIRepairConfiguration struct {
	Enabled             *bool   `json:"enabled,omitempty"`
	Hub                 *string `json:"hub,omitempty"`
	Tag                 *string `json:"tag,omitempty"`
	LabelPods           *bool   `json:"labelPods,omitempty"`
	DeletePods          *bool   `json:"deletePods,omitempty"`
	InitContainerName   *string `json:"initContainerName,omitempty"`
	BrokenPodLabelKey   *string `json:"brokenPodLabelKey,omitempty"`
	BrokenPodLabelValue *string `json:"brokenPodLabelValue,omitempty"`
}

// InitCNIConfiguration defines config for the sidecar proxy init CNI plugin
type InitCNIConfiguration struct {
	// If true, the privileged initContainer istio-init is not needed to perform the traffic redirect
	// settings for the istio-proxy
	Enabled *bool  `json:"enabled,omitempty"`
	Image   string `json:"image,omitempty"`
	// Must be the same as the environment’s --cni-bin-dir setting (kubelet parameter)
	BinDir string `json:"binDir,omitempty"`
	// Must be the same as the environment’s --cni-conf-dir setting (kubelet parameter)
	ConfDir string `json:"confDir,omitempty"`
	// List of namespaces to exclude from Istio pod check
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`
	// Logging level for CNI binary
	LogLevel string                 `json:"logLevel,omitempty"`
	Affinity *corev1.Affinity       `json:"affinity,omitempty"`
	Chained  *bool                  `json:"chained,omitempty"`
	Repair   CNIRepairConfiguration `json:"repair,omitempty"`
}

// SidecarInjectorConfiguration defines config options for SidecarInjector
type SidecarInjectorConfiguration struct {
	Enabled                                  *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithReplicas `json:",inline"`
	Init                                     SidecarInjectorInitConfiguration `json:"init,omitempty"`
	InitCNIConfiguration                     InitCNIConfiguration             `json:"initCNIConfiguration,omitempty"`
	// If true, sidecar injector will rewrite PodSpec for liveness
	// health check to redirect request to sidecar. This makes liveness check work
	// even when mTLS is enabled.
	RewriteAppHTTPProbe bool `json:"rewriteAppHTTPProbe,omitempty"`
	// This controls the 'policy' in the sidecar injector
	AutoInjectionPolicyEnabled *bool `json:"autoInjectionPolicyEnabled,omitempty"`
	// This controls whether the webhook looks for namespaces for injection enabled or disabled
	EnableNamespacesByDefault *bool `json:"enableNamespacesByDefault,omitempty"`
	// NeverInjectSelector: Refuses the injection on pods whose labels match this selector.
	// It's an array of label selectors, that will be OR'ed, meaning we will iterate
	// over it and stop at the first match
	// Takes precedence over AlwaysInjectSelector.
	NeverInjectSelector []metav1.LabelSelector `json:"neverInjectSelector,omitempty"`
	// AlwaysInjectSelector: Forces the injection on pods whose labels match this selector.
	// It's an array of label selectors, that will be OR'ed, meaning we will iterate
	// over it and stop at the first match
	AlwaysInjectSelector []metav1.LabelSelector `json:"alwaysInjectSelector,omitempty"`
	// injectedAnnotations are additional annotations that will be added to the pod spec after injection
	// This is primarily to support PSP annotations. For example, if you defined a PSP with the annotations:
	//
	// annotations:
	//   apparmor.security.beta.kubernetes.io/allowedProfileNames: runtime/default
	//   apparmor.security.beta.kubernetes.io/defaultProfileName: runtime/default
	//
	// The PSP controller would add corresponding annotations to the pod spec for each container. However, this happens before
	// the inject adds additional containers, so we must specify them explicitly here. With the above example, we could specify:
	// injectedAnnotations:
	//   container.apparmor.security.beta.kubernetes.io/istio-init: runtime/default
	//   container.apparmor.security.beta.kubernetes.io/istio-proxy: runtime/default
	InjectedAnnotations map[string]string `json:"injectedAnnotations,omitempty"`

	// If present will be appended at the end of the initial/preconfigured container arguments
	AdditionalContainerArgs []string `json:"additionalContainerArgs,omitempty"`

	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`

	// If present will be appended at the end of the initial/preconfigured container arguments
	InjectedContainerAdditionalArgs []string `json:"injectedContainerAdditionalArgs,omitempty"`

	// If present will be appended to the environment variables of the container
	InjectedContainerAdditionalEnvVars []corev1.EnvVar `json:"injectedContainerAdditionalEnvVars,omitempty"`
}

// ProxyWasmConfiguration defines config options for Envoy wasm
type ProxyWasmConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type GatewayConfiguration struct {
	MeshGatewayConfiguration `json:",inline"`
	Ports                    []corev1.ServicePort `json:"ports,omitempty"`
	Enabled                  *bool                `json:"enabled,omitempty"`
}

type BaseK8sResourceConfigurationWithHPAWithoutImage struct {
	ReplicaCount                 *int32 `json:"replicaCount,omitempty"`
	MinReplicas                  *int32 `json:"minReplicas,omitempty"`
	MaxReplicas                  *int32 `json:"maxReplicas,omitempty"`
	BaseK8sResourceConfiguration `json:",inline"`
}

type GatewaySDSConfiguration struct {
	Enabled   *bool                        `json:"enabled,omitempty"`
	Image     string                       `json:"image,omitempty"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

type K8sIngressConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
	// enableHttps will add port 443 on the ingress.
	// It REQUIRES that the certificates are installed  in the
	// expected secrets - enabling this option without certificates
	// will result in LDS rejection and the ingress will not work.
	EnableHttps *bool `json:"enableHttps,omitempty"`
}

// GatewaysConfiguration defines config options for Gateways
type GatewaysConfiguration struct {
	Enabled       *bool                   `json:"enabled,omitempty"`
	IngressConfig GatewayConfiguration    `json:"ingress,omitempty"`
	EgressConfig  GatewayConfiguration    `json:"egress,omitempty"`
	K8sIngress    K8sIngressConfiguration `json:"k8singress,omitempty"`
}

type EnvoyStatsD struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Host    string `json:"host,omitempty"`
	Port    int32  `json:"port,omitempty"`
}

type TLSSettings struct {
	// +kubebuilder:validation:Enum="DISABLE,SIMPLE,MUTUAL,ISTIO_MUTUAL"
	Mode              string   `json:"mode,omitempty"`
	ClientCertificate string   `json:"clientCertificate,omitempty"`
	PrivateKey        string   `json:"privateKey,omitempty"`
	CACertificates    string   `json:"caCertificates,omitempty"`
	SNI               string   `json:"sni,omitempty"`
	SubjectAltNames   []string `json:"subjectAltNames,omitempty"`
}

type TCPKeepalive struct {
	Probes   int32  `json:"probes,omitempty"`
	Time     string `json:"time,omitempty"`
	Interval string `json:"interval,omitempty"`
}

type EnvoyServiceCommonConfiguration struct {
	Enabled      *bool         `json:"enabled,omitempty"`
	Host         string        `json:"host,omitempty"`
	Port         int32         `json:"port,omitempty"`
	TLSSettings  *TLSSettings  `json:"tlsSettings,omitempty"`
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty"`
}

// ProxyConfiguration defines config options for Proxy
type ProxyConfiguration struct {
	Image string `json:"image,omitempty"`
	// Configures the access log for each sidecar.
	// Options:
	//   "" - disables access log
	//   "/dev/stdout" - enables access log
	// +kubebuilder:validation:Enum=",/dev/stdout"
	AccessLogFile *string `json:"accessLogFile,omitempty"`
	// Configure how and what fields are displayed in sidecar access log. Setting to
	// empty string will result in default log format.
	// If accessLogEncoding is TEXT, value will be used directly as the log format
	// example: "[%START_TIME%] %REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\n"
	// If AccessLogEncoding is JSON, value will be parsed as map[string]string
	// example: '{"start_time": "%START_TIME%", "req_method": "%REQ(:METHOD)%"}'
	AccessLogFormat *string `json:"accessLogFormat,omitempty"`
	// Configure the access log for sidecar to JSON or TEXT.
	// +kubebuilder:validation:Enum="JSON,TEXT"
	AccessLogEncoding *string `json:"accessLogEncoding,omitempty"`
	// If set to true, istio-proxy container will have privileged securityContext
	Privileged bool `json:"privileged,omitempty"`
	// If set, newly injected sidecars will have core dumps enabled.
	EnableCoreDump *bool `json:"enableCoreDump,omitempty"`
	// Image used to enable core dumps. This is only used, when "EnableCoreDump" is set to true.
	CoreDumpImage string `json:"coreDumpImage,omitempty"`
	// Log level for proxy, applies to gateways and sidecars. If left empty, "warning" is used.
	// Expected values are: trace|debug|info|warning|error|critical|off
	// +kubebuilder:validation:Enum="trace,debug,info,warning,error,critical,off"
	LogLevel string `json:"logLevel,omitempty"`
	// Per Component log level for proxy, applies to gateways and sidecars. If a component level is
	// not set, then the "LogLevel" will be used. If left empty, "misc:error" is used.
	ComponentLogLevel string `json:"componentLogLevel,omitempty"`
	// Configure the DNS refresh rate for Envoy cluster of type STRICT_DNS
	// This must be given it terms of seconds. For example, 300s is valid but 5m is invalid.
	// +kubebuilder:validation:Pattern="^[0-9]{1,5}s$"
	DNSRefreshRate string `json:"dnsRefreshRate,omitempty"`
	// cluster domain. Default value is "cluster.local"
	ClusterDomain string `json:"clusterDomain,omitempty"`

	EnvoyStatsD               EnvoyStatsD                     `json:"envoyStatsD,omitempty"`
	EnvoyMetricsService       EnvoyServiceCommonConfiguration `json:"envoyMetricsService,omitempty"`
	EnvoyAccessLogService     EnvoyServiceCommonConfiguration `json:"envoyAccessLogService,omitempty"`
	ProtocolDetectionTimeout  *string                         `json:"protocolDetectionTimeout,omitempty"`
	UseMetadataExchangeFilter *bool                           `json:"useMetadataExchangeFilter,omitempty"`

	Lifecycle corev1.Lifecycle `json:"lifecycle,omitempty"`

	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// IstioSpec defines the desired state of Istio
type IstioSpec struct {
	// Contains the intended Istio version validation:Pattern=^1.5
	Version IstioVersion `json:"version"`

	// Logging configurations
	Logging LoggingConfiguration `json:"logging,omitempty"`

	// MeshPolicy configures the default MeshPolicy resource
	MeshPolicy MeshPolicyConfiguration `json:"meshPolicy,omitempty"`

	// If set to true, and a given service does not have a corresponding DestinationRule configured,
	// or its DestinationRule does not have TLSSettings specified, Istio configures client side
	// TLS configuration automatically, based on the server side mTLS authentication policy and the
	// availability of sidecars.
	AutoMTLS *bool `json:"autoMtls,omitempty"`

	// IncludeIPRanges the range where to capture egress traffic
	IncludeIPRanges string `json:"includeIPRanges,omitempty"`

	// ExcludeIPRanges the range where not to capture egress traffic
	ExcludeIPRanges string `json:"excludeIPRanges,omitempty"`

	// List of namespaces to label with sidecar auto injection enabled
	AutoInjectionNamespaces []string `json:"autoInjectionNamespaces,omitempty"`

	// ControlPlaneSecurityEnabled control plane services are communicating through mTLS
	ControlPlaneSecurityEnabled bool `json:"controlPlaneSecurityEnabled,omitempty"`

	// Use the user-specified, secret volume mounted key and certs for Pilot and workloads.
	MountMtlsCerts *bool `json:"mountMtlsCerts,omitempty"`

	// DefaultResources are applied for all Istio components by default, can be overridden for each component
	DefaultResources *corev1.ResourceRequirements `json:"defaultResources,omitempty"`

	// Istiod configuration
	Istiod IstiodConfiguration `json:"istiod,omitempty"`

	// Pilot configuration options
	Pilot PilotConfiguration `json:"pilot,omitempty"`

	// Gateways configuration options
	Gateways GatewaysConfiguration `json:"gateways,omitempty"`

	// SidecarInjector configuration options
	SidecarInjector SidecarInjectorConfiguration `json:"sidecarInjector,omitempty"`

	// ProxyWasm configuration options
	ProxyWasm ProxyWasmConfiguration `json:"proxyWasm,omitempty"`

	// Proxy configuration options
	Proxy ProxyConfiguration `json:"proxy,omitempty"`

	// Proxy Init configuration options
	ProxyInit ProxyInitConfiguration `json:"proxyInit,omitempty"`

	// Whether to restrict the applications namespace the controller manages
	WatchOneNamespace bool `json:"watchOneNamespace,omitempty"`

	// Prior to Kubernetes v1.17.0 it was not allowed to use the system-cluster-critical and system-node-critical
	// PriorityClass outside of the kube-system namespace, so it is advised to create your own PriorityClass
	// and use its name here
	// On Kubernetes >=v1.17.0 it is possible to configure system-cluster-critical and
	// system-node-critical PriorityClass in order to make sure your Istio pods
	// will not be killed because of low priority class.
	// Refer to https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass
	// for more detail.
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// Use the Mesh Control Protocol (MCP) for configuring Mixer and Pilot. Requires galley.
	UseMCP *bool `json:"useMCP,omitempty"`

	// Set the default set of namespaces to which services, service entries, virtual services, destination rules should be exported to
	DefaultConfigVisibility string `json:"defaultConfigVisibility,omitempty"`

	// Whether or not to establish watches for adapter-specific CRDs
	WatchAdapterCRDs bool `json:"watchAdapterCRDs,omitempty"`

	// Enable pod disruption budget for the control plane, which is used to ensure Istio control plane components are gradually upgraded or recovered
	DefaultPodDisruptionBudget PDBConfiguration `json:"defaultPodDisruptionBudget,omitempty"`

	// Set the default behavior of the sidecar for handling outbound traffic from the application (ALLOW_ANY or REGISTRY_ONLY)
	OutboundTrafficPolicy OutboundTrafficPolicyConfiguration `json:"outboundTrafficPolicy,omitempty"`

	// Configuration for each of the supported tracers
	Tracing TracingConfiguration `json:"tracing,omitempty"`

	// ImagePullPolicy describes a policy for if/when to pull a container image
	// +kubebuilder:validation:Enum="Always,Never,IfNotPresent"
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// If set to true, the pilot and citadel mtls will be exposed on the
	// ingress gateway also the remote istios will be connected through gateways
	MeshExpansion *bool `json:"meshExpansion,omitempty"`

	// Set to true to connect two or more meshes via their respective
	// ingressgateway services when workloads in each cluster cannot directly
	// talk to one another. All meshes should be using Istio mTLS and must
	// have a shared root CA for this model to work.
	MultiMesh *bool `json:"multiMesh,omitempty"`

	// Istio CoreDNS provides DNS resolution for services in multi mesh setups
	IstioCoreDNS IstioCoreDNS `json:"istioCoreDNS,omitempty"`

	// Locality based load balancing distribution or failover settings.
	LocalityLB *LocalityLBConfiguration `json:"localityLB,omitempty"`

	// Should be set to the name of the cluster this installation will run in.
	// This is required for sidecar injection to properly label proxies
	ClusterName string `json:"clusterName,omitempty"`

	// Network defines the network this cluster belong to. This name
	// corresponds to the networks in the map of mesh networks.
	NetworkName string `json:"networkName,omitempty"`

	// Mesh ID means Mesh Identifier. It should be unique within the scope where
	// meshes will interact with each other, but it is not required to be
	// globally/universally unique.
	MeshID string `json:"meshID,omitempty"`

	// Mixerless telemetry configuration
	MixerlessTelemetry *MixerlessTelemetryConfiguration `json:"mixerlessTelemetry,omitempty"`

	//
	MeshNetworks *MeshNetworks `json:"meshNetworks,omitempty"`

	// The domain serves to identify the system with SPIFFE. (default "cluster.local")
	TrustDomain string `json:"trustDomain,omitempty"`

	//  The trust domain aliases represent the aliases of trust_domain.
	//  For example, if we have
	//  trustDomain: td1
	//  trustDomainAliases: [“td2”, "td3"]
	//  Any service with the identity "td1/ns/foo/sa/a-service-account", "td2/ns/foo/sa/a-service-account",
	//  or "td3/ns/foo/sa/a-service-account" will be treated the same in the Istio mesh.
	TrustDomainAliases []string `json:"trustDomainAliases,omitempty"`

	// Configures DNS certificates provisioned through Chiron linked into Pilot.
	// The DNS names in this file are all hard-coded; please ensure the namespaces
	// in dnsNames are consistent with those of your services.
	// Example:
	// certificates:
	//   - secretName: dns.istio-galley-service-account
	//     dnsNames: [istio-galley.istio-system.svc, istio-galley.istio-system]
	//   - secretName: dns.istio-sidecar-injector-service-account
	//     dnsNames: [istio-sidecar-injector.istio-system.svc, istio-sidecar-injector.istio-system]
	// +k8s:deepcopy-gen:interfaces=Certificates
	Certificates []CertificateConfig `json:"certificates,omitempty"`

	// Configure the policy for validating JWT.
	// Currently, two options are supported: "third-party-jwt" and "first-party-jwt".
	// +kubebuilder:validation:Enum="third-party-jwt,first-party-jwt"
	JWTPolicy JWTPolicyType `json:"jwtPolicy,omitempty"`

	// The customized CA address to retrieve certificates for the pods in the cluster.
	//CSR clients such as the Istio Agent and ingress gateways can use this to specify the CA endpoint.
	CAAddress string `json:"caAddress,omitempty"`
}

// IstioStatus defines the observed state of Istio
type IstioStatus struct {
	Status         ConfigState `json:"Status,omitempty"`
	GatewayAddress []string    `json:"GatewayAddress,omitempty"`
	ErrorMessage   string      `json:"ErrorMessage,omitempty"`
}

// +kubebuilder:object:root=true

// Istio is the Schema for the Istios API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.Status",description="Status of the resource"
// +kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.ErrorMessage",description="Error message"
// +kubebuilder:printcolumn:name="Ingress IPs",type="string",JSONPath=".status.GatewayAddress",description="Ingress gateway addresses of the resource"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Istio struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IstioSpec   `json:"spec,omitempty"`
	Status IstioStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IstioList contains a list of Istio
type IstioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Istio `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Istio{}, &IstioList{})
}
