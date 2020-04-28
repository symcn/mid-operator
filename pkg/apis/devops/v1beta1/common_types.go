package v1beta1

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

type ConfigState string

const (
	Created         ConfigState = "Created"
	ReconcileFailed ConfigState = "ReconcileFailed"
	Reconciling     ConfigState = "Reconciling"
	Available       ConfigState = "Available"
	Unmanaged       ConfigState = "Unmanaged"
)

type MeshNetworkEndpoint struct {
	FromCIDR     string `json:"fromCidr,omitempty"`
	FromRegistry string `json:"fromRegistry,omitempty"`
}

type MeshNetworkGateway struct {
	Address string `json:"address"`
	Port    uint   `json:"port"`
}

type MeshNetwork struct {
	Endpoints []MeshNetworkEndpoint `json:"endpoints,omitempty"`
	Gateways  []MeshNetworkGateway  `json:"gateways,omitempty"`
}

type MeshNetworks struct {
	Networks map[string]MeshNetwork `json:"networks"`
}

func (s *IstioSpec) SetMeshNetworks(networks *MeshNetworks) *IstioSpec {
	s.meshNetworks = networks
	return s
}

func (s *IstioSpec) GetMeshNetworks() *MeshNetworks {
	return s.meshNetworks
}

func (s *IstioSpec) GetMeshNetworksHash() string {
	hash := ""
	j, err := json.Marshal(s.meshNetworks)
	if err != nil {
		return hash
	}

	hash = fmt.Sprintf("%x", md5.Sum(j))

	return hash
}

const supportedIstioMinorVersionRegex = "^1.5"

// IstioVersion stores the intended Istio version
type IstioVersion string

type MTLSMode string

const (
	STRICT     MTLSMode = "STRICT"
	PERMISSIVE MTLSMode = "PERMISSIVE"
	DISABLED   MTLSMode = "DISABLED"
)

type PilotCertProviderType string

const (
	PilotCertProviderTypeKubernetes PilotCertProviderType = "kubernetes"
	PilotCertProviderTypeIstiod     PilotCertProviderType = "istiod"
)

type JWTPolicyType string

const (
	JWTPolicyThirdPartyJWT JWTPolicyType = "third-party-jwt"
	JWTPolicyFirstPartyJWT JWTPolicyType = "first-party-jwt"
)

// BaseK8sResourceConfiguration defines basic K8s resource spec configurations
type BaseK8sResourceConfiguration struct {
	Resources      *corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector   map[string]string            `json:"nodeSelector,omitempty"`
	Affinity       *corev1.Affinity             `json:"affinity,omitempty"`
	Tolerations    []corev1.Toleration          `json:"tolerations,omitempty"`
	PodAnnotations map[string]string            `json:"podAnnotations,omitempty"`
}

type BaseK8sResourceConfigurationWithImage struct {
	Image                        *string `json:"image,omitempty"`
	BaseK8sResourceConfiguration `json:",inline"`
}

type BaseK8sResourceConfigurationWithReplicas struct {
	ReplicaCount                          *int32 `json:"replicaCount,omitempty"`
	BaseK8sResourceConfigurationWithImage `json:",inline"`
}

type BaseK8sResourceConfigurationWithHPA struct {
	MinReplicas                              *int32 `json:"minReplicas,omitempty"`
	MaxReplicas                              *int32 `json:"maxReplicas,omitempty"`
	BaseK8sResourceConfigurationWithReplicas `json:",inline"`
}

// Describes how traffic originating in the 'from' zone is
// distributed over a set of 'to' zones. Syntax for specifying a zone is
// {region}/{zone} and terminal wildcards are allowed on any
// segment of the specification. Examples:
// * - matches all localities
// us-west/* - all zones and sub-zones within the us-west region
type LocalityLBDistributeConfiguration struct {
	// Originating locality, '/' separated, e.g. 'region/zone'.
	From string `json:"from,omitempty"`
	// Map of upstream localities to traffic distribution weights. The sum of
	// all weights should be == 100. Any locality not assigned a weight will
	// receive no traffic.
	To map[string]uint32 `json:"to,omitempty"`
}

// Specify the traffic failover policy across regions. Since zone
// failover is supported by default this only needs to be specified for
// regions when the operator needs to constrain traffic failover so that
// the default behavior of failing over to any endpoint globally does not
// apply. This is useful when failing over traffic across regions would not
// improve service health or may need to be restricted for other reasons
// like regulatory controls.
type LocalityLBFailoverConfiguration struct {
	// Originating region.
	From string `json:"from,omitempty"`
	// Destination region the traffic will fail over to when endpoints in
	// the 'from' region becomes unhealthy.
	To string `json:"to,omitempty"`
}

// Locality-weighted load balancing allows administrators to control the
// distribution of traffic to endpoints based on the localities of where the
// traffic originates and where it will terminate.
type LocalityLBConfiguration struct {
	// If set to true, locality based load balancing will be enabled
	Enabled *bool `json:"enabled,omitempty"`
	// Optional: only one of distribute or failover can be set.
	// Explicitly specify loadbalancing weight across different zones and geographical locations.
	// Refer to [Locality weighted load balancing](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/load_balancing/locality_weight)
	// If empty, the locality weight is set according to the endpoints number within it.
	Distribute []*LocalityLBDistributeConfiguration `json:"distribute,omitempty"`
	// Optional: only failover or distribute can be set.
	// Explicitly specify the region traffic will land on when endpoints in local region becomes unhealthy.
	// Should be used together with OutlierDetection to detect unhealthy endpoints.
	// Note: if no OutlierDetection specified, this will not take effect.
	Failover []*LocalityLBFailoverConfiguration `json:"failover,omitempty"`
}

// IstioCoreDNS
type IstioCoreDNS struct {
	Enabled                                  *bool `json:"enabled,omitempty"`
	BaseK8sResourceConfigurationWithReplicas `json:",inline"`
	PluginImage                              string `json:"pluginImage,omitempty"`
}

// PDBConfiguration holds Pod Disruption Budget related config options
type PDBConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type OutboundTrafficPolicyConfiguration struct {
	// +kubebuilder:validation:Enum="ALLOW_ANY,REGISTRY_ONLY"
	Mode string `json:"mode,omitempty"`
}

// ProxyInitConfiguration defines config options for Proxy Init containers
type ProxyInitConfiguration struct {
	Image string `json:"image,omitempty"`
}

type TracerType string

const (
	TracerTypeZipkin      TracerType = "zipkin"
	TracerTypeLightstep   TracerType = "lightstep"
	TracerTypeDatadog     TracerType = "datadog"
	TracerTypeStackdriver TracerType = "stackdriver"
)

// Configuration for Envoy to send trace data to Zipkin/Jaeger.
type ZipkinConfiguration struct {
	// Host:Port for reporting trace data in zipkin format. If not specified, will default to zipkin service (port 9411) in the same namespace as the other istio components.
	Address string `json:"address,omitempty"`
	// TLS setting for Zipkin endpoint.
	TLSSettings *TLSSettings `json:"tlsSettings,omitempty"`
}

// Configuration for Envoy to send trace data to Lightstep
type LightstepConfiguration struct {
	// the <host>:<port> of the satellite pool
	Address string `json:"address,omitempty"`
	// required for sending data to the pool
	AccessToken string `json:"accessToken,omitempty"`
	// specifies whether data should be sent with TLS
	Secure bool `json:"secure,omitempty"`
	// the path to the file containing the cacert to use when verifying TLS. If secure is true, this is
	// required. If a value is specified then a secret called "lightstep.cacert" must be created in the destination
	// namespace with the key matching the base of the provided cacertPath and the value being the cacert itself.
	CacertPath string `json:"cacertPath,omitempty"`
}

// Configuration for Envoy to send trace data to Datadog
type DatadogConfiugration struct {
	// Host:Port for submitting traces to the Datadog agent.
	Address string `json:"address,omitempty"`
}

type StrackdriverConfiguration struct {
	// enables trace output to stdout.
	Debug *bool `json:"debug,omitempty"`
	// The global default max number of attributes per span.
	MaxNumberOfAttributes *int32 `json:"maxNumberOfAttributes,omitempty"`
	// The global default max number of annotation events per span.
	MaxNumberOfAnnotations *int32 `json:"maxNumberOfAnnotations,omitempty"`
	// The global default max number of message events per span.
	MaxNumberOfMessageEvents *int32 `json:"maxNumberOfMessageEvents,omitempty"`
}

//
type TracingConfiguration struct {
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Enum="zipkin,lightstep,datadog,stackdriver"
	Tracer       TracerType                `json:"tracer,omitempty"`
	Zipkin       ZipkinConfiguration       `json:"zipkin,omitempty"`
	Lightstep    LightstepConfiguration    `json:"lightstep,omitempty"`
	Datadog      DatadogConfiugration      `json:"datadog,omitempty"`
	Strackdriver StrackdriverConfiguration `json:"stackdriver,omitempty"`
}

type MeshGatewayConfiguration struct {
	BaseK8sResourceConfigurationWithHPAWithoutImage `json:",inline"`
	Labels                                          map[string]string `json:"labels,omitempty"`
	// +kubebuilder:validation:Enum="ClusterIP,NodePort,LoadBalancer"
	ServiceType        corev1.ServiceType `json:"serviceType,omitempty"`
	LoadBalancerIP     string             `json:"loadBalancerIP,omitempty"`
	ServiceAnnotations map[string]string  `json:"serviceAnnotations,omitempty"`
	ServiceLabels      map[string]string  `json:"serviceLabels,omitempty"`

	RequestedNetworkView string `json:"requestedNetworkView,omitempty"`
	// If present will be appended to the environment variables of the container
	AdditionalEnvVars []corev1.EnvVar `json:"additionalEnvVars,omitempty"`
}
