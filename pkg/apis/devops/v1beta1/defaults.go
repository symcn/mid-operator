package v1beta1

import (
	"fmt"

	"github.com/symcn/mid-operator/pkg/utils"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	defaultImageHub                   = "docker.io/istio"
	defaultImageVersion               = "1.5.2"
	defaultLogLevel                   = "default:info"
	defaultMeshPolicy                 = PERMISSIVE
	defaultPilotImage                 = defaultImageHub + "/" + "pilot" + ":" + defaultImageVersion
	defaultSidecarInjectorImage       = defaultImageHub + "/" + "sidecar_injector" + ":" + defaultImageVersion
	defaultProxyImage                 = defaultImageHub + "/" + "proxyv2" + ":" + defaultImageVersion
	defaultProxyInitImage             = defaultImageHub + "/" + "proxyv2" + ":" + defaultImageVersion
	defaultProxyCoreDumpImage         = "busybox"
	defaultInitCNIImage               = defaultImageHub + "/" + "install-cni:" + defaultImageVersion
	defaultCoreDNSImage               = "coredns/coredns:1.6.2"
	defaultCoreDNSPluginImage         = defaultImageHub + "/coredns-plugin:0.2-istio-1.1"
	defaultIncludeIPRanges            = "*"
	defaultReplicaCount               = 1
	defaultMinReplicas                = 1
	defaultMaxReplicas                = 5
	defaultTraceSampling              = 1.0
	defaultIngressGatewayServiceType  = apiv1.ServiceTypeLoadBalancer
	defaultEgressGatewayServiceType   = apiv1.ServiceTypeClusterIP
	outboundTrafficPolicyAllowAny     = "ALLOW_ANY"
	defaultZipkinAddress              = "zipkin.%s:9411"
	defaultInitCNIBinDir              = "/opt/cni/bin"
	defaultInitCNIConfDir             = "/etc/cni/net.d"
	defaultInitCNILogLevel            = "info"
	defaultInitCNIContainerName       = "istio-validation"
	defaultInitCNIBrokenPodLabelKey   = "cni.istio.io/uninitialized"
	defaultInitCNIBrokenPodLabelValue = "true"
	defaultImagePullPolicy            = "IfNotPresent"
	defaultEnvoyAccessLogFile         = "/dev/stdout"
	defaultEnvoyAccessLogFormat       = ""
	defaultEnvoyAccessLogEncoding     = "TEXT"
	defaultClusterName                = "Kubernetes"
	defaultNetworkName                = "local-network"
)

var defaultResources = &apiv1.ResourceRequirements{
	Requests: apiv1.ResourceList{
		apiv1.ResourceCPU: resource.MustParse("10m"),
	},
}

var defaultProxyResources = &apiv1.ResourceRequirements{
	Requests: apiv1.ResourceList{
		apiv1.ResourceCPU:    resource.MustParse("100m"),
		apiv1.ResourceMemory: resource.MustParse("128Mi"),
	},
	Limits: apiv1.ResourceList{
		apiv1.ResourceCPU:    resource.MustParse("2000m"),
		apiv1.ResourceMemory: resource.MustParse("1024Mi"),
	},
}

var defaultInitResources = &apiv1.ResourceRequirements{
	Requests: apiv1.ResourceList{
		apiv1.ResourceCPU:    resource.MustParse("10m"),
		apiv1.ResourceMemory: resource.MustParse("10Mi"),
	},
	Limits: apiv1.ResourceList{
		apiv1.ResourceCPU:    resource.MustParse("100m"),
		apiv1.ResourceMemory: resource.MustParse("50Mi"),
	},
}

var defaultIngressGatewayPorts = []apiv1.ServicePort{
	{Port: 15020, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15020), Name: "status-port"},
	{Port: 80, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(80), Name: "http2"},
	{Port: 443, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(443), Name: "https"},
	{Port: 15443, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15443), Name: "tls"},
	{Port: 31400, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(31400), Name: "tcp"},
}

var defaultEgressGatewayPorts = []apiv1.ServicePort{
	{Port: 80, Name: "http2", Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(80)},
	{Port: 443, Name: "https", Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(443)},
	{Port: 15443, Protocol: apiv1.ProtocolTCP, TargetPort: intstr.FromInt(15443), Name: "tls"},
}

func SetDefaults(config *Istio) {
	// MeshPolicy config
	if config.Spec.MeshPolicy.MTLSMode == "" {
		config.Spec.MeshPolicy.MTLSMode = defaultMeshPolicy
	}

	if config.Spec.ClusterName == "" {
		config.Spec.ClusterName = defaultClusterName
	}

	if config.Spec.NetworkName == "" {
		config.Spec.NetworkName = defaultNetworkName
	}

	if config.Spec.AutoMTLS == nil {
		config.Spec.AutoMTLS = utils.BoolPointer(true)
	}

	if config.Spec.IncludeIPRanges == "" {
		config.Spec.IncludeIPRanges = defaultIncludeIPRanges
	}
	if config.Spec.MountMtlsCerts == nil {
		config.Spec.MountMtlsCerts = utils.BoolPointer(false)
	}
	if config.Spec.Logging.Level == nil {
		config.Spec.Logging.Level = utils.StrPointer(defaultLogLevel)
	}
	if config.Spec.Proxy.Resources == nil {
		if config.Spec.DefaultResources == nil {
			config.Spec.Proxy.Resources = defaultProxyResources
		} else {
			config.Spec.Proxy.Resources = defaultResources
		}
	}
	if config.Spec.DefaultResources == nil {
		config.Spec.DefaultResources = defaultResources
	}

	// Istiod config
	if config.Spec.Istiod.Enabled == nil {
		config.Spec.Istiod.Enabled = utils.BoolPointer(true)
	}

	// Pilot config
	if config.Spec.Pilot.Enabled == nil {
		config.Spec.Pilot.Enabled = utils.BoolPointer(true)
	}
	if config.Spec.Pilot.Image == nil {
		config.Spec.Pilot.Image = utils.StrPointer(defaultPilotImage)
	}
	if config.Spec.Pilot.Sidecar == nil {
		config.Spec.Pilot.Sidecar = utils.BoolPointer(true)
	}
	if config.Spec.Pilot.ReplicaCount == nil {
		config.Spec.Pilot.ReplicaCount = utils.IntPointer(defaultReplicaCount)
	}
	if config.Spec.Pilot.MinReplicas == nil {
		config.Spec.Pilot.MinReplicas = utils.IntPointer(defaultMinReplicas)
	}
	if config.Spec.Pilot.MaxReplicas == nil {
		config.Spec.Pilot.MaxReplicas = utils.IntPointer(defaultMaxReplicas)
	}
	if config.Spec.Pilot.TraceSampling == 0 {
		config.Spec.Pilot.TraceSampling = defaultTraceSampling
	}
	if config.Spec.Pilot.EnableProtocolSniffingOutbound == nil {
		config.Spec.Pilot.EnableProtocolSniffingOutbound = utils.BoolPointer(true)
	}
	if config.Spec.Pilot.EnableProtocolSniffingInbound == nil {
		config.Spec.Pilot.EnableProtocolSniffingInbound = utils.BoolPointer(false)
	}
	if config.Spec.Pilot.CertProvider == "" {
		config.Spec.Pilot.CertProvider = PilotCertProviderTypeIstiod
	}

	// Gateways config
	if config.Spec.Gateways.Enabled == nil {
		config.Spec.Gateways.Enabled = utils.BoolPointer(true)
	}
	if config.Spec.Gateways.IngressConfig.Enabled == nil {
		config.Spec.Gateways.IngressConfig.Enabled = utils.BoolPointer(true)
	}
	if config.Spec.Gateways.IngressConfig.ReplicaCount == nil {
		config.Spec.Gateways.IngressConfig.ReplicaCount = utils.IntPointer(defaultReplicaCount)
	}
	if config.Spec.Gateways.IngressConfig.MinReplicas == nil {
		config.Spec.Gateways.IngressConfig.MinReplicas = utils.IntPointer(defaultMinReplicas)
	}
	if config.Spec.Gateways.IngressConfig.MaxReplicas == nil {
		config.Spec.Gateways.IngressConfig.MaxReplicas = utils.IntPointer(defaultMaxReplicas)
	}

	if len(config.Spec.Gateways.IngressConfig.Ports) == 0 {
		config.Spec.Gateways.IngressConfig.Ports = defaultIngressGatewayPorts
	}
	if config.Spec.Gateways.EgressConfig.Enabled == nil {
		config.Spec.Gateways.EgressConfig.Enabled = utils.BoolPointer(false)
	}
	if config.Spec.Gateways.EgressConfig.ReplicaCount == nil {
		config.Spec.Gateways.EgressConfig.ReplicaCount = utils.IntPointer(defaultReplicaCount)
	}
	if config.Spec.Gateways.EgressConfig.MinReplicas == nil {
		config.Spec.Gateways.EgressConfig.MinReplicas = utils.IntPointer(defaultMinReplicas)
	}
	if config.Spec.Gateways.EgressConfig.MaxReplicas == nil {
		config.Spec.Gateways.EgressConfig.MaxReplicas = utils.IntPointer(defaultMaxReplicas)
	}
	if config.Spec.Gateways.IngressConfig.ServiceType == "" {
		config.Spec.Gateways.IngressConfig.ServiceType = defaultIngressGatewayServiceType
	}
	if config.Spec.Gateways.EgressConfig.ServiceType == "" {
		config.Spec.Gateways.EgressConfig.ServiceType = defaultEgressGatewayServiceType
	}
	if len(config.Spec.Gateways.EgressConfig.Ports) == 0 {
		config.Spec.Gateways.EgressConfig.Ports = defaultEgressGatewayPorts
	}
	if config.Spec.Gateways.K8sIngress.Enabled == nil {
		config.Spec.Gateways.K8sIngress.Enabled = utils.BoolPointer(false)
	}
	if config.Spec.Gateways.K8sIngress.EnableHttps == nil {
		config.Spec.Gateways.K8sIngress.EnableHttps = utils.BoolPointer(false)
	}

	// SidecarInjector config
	if config.Spec.SidecarInjector.Enabled == nil {
		config.Spec.SidecarInjector.Enabled = utils.BoolPointer(false)
	}
	if config.Spec.SidecarInjector.AutoInjectionPolicyEnabled == nil {
		config.Spec.SidecarInjector.AutoInjectionPolicyEnabled = utils.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.Image == nil {
		config.Spec.SidecarInjector.Image = utils.StrPointer(defaultSidecarInjectorImage)
	}
	if config.Spec.SidecarInjector.ReplicaCount == nil {
		config.Spec.SidecarInjector.ReplicaCount = utils.IntPointer(defaultReplicaCount)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Enabled == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Enabled = utils.BoolPointer(false)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Image == "" {
		config.Spec.SidecarInjector.InitCNIConfiguration.Image = defaultInitCNIImage
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.BinDir == "" {
		config.Spec.SidecarInjector.InitCNIConfiguration.BinDir = defaultInitCNIBinDir
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.ConfDir == "" {
		config.Spec.SidecarInjector.InitCNIConfiguration.ConfDir = defaultInitCNIConfDir
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.ExcludeNamespaces == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.ExcludeNamespaces = []string{config.Namespace}
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.LogLevel == "" {
		config.Spec.SidecarInjector.InitCNIConfiguration.LogLevel = defaultInitCNILogLevel
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Chained == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Chained = utils.BoolPointer(true)
	}
	// Wasm Config
	if config.Spec.ProxyWasm.Enabled == nil {
		config.Spec.ProxyWasm.Enabled = utils.BoolPointer(false)
	}
	// CNI repair config
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Enabled == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Enabled = utils.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Hub == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Hub = utils.StrPointer("")
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Tag == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Tag = utils.StrPointer("")
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.LabelPods == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.LabelPods = utils.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.DeletePods == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.DeletePods = utils.BoolPointer(true)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.InitContainerName == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.InitContainerName = utils.StrPointer(defaultInitCNIContainerName)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.BrokenPodLabelKey == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.BrokenPodLabelKey = utils.StrPointer(defaultInitCNIBrokenPodLabelKey)
	}
	if config.Spec.SidecarInjector.InitCNIConfiguration.Repair.BrokenPodLabelValue == nil {
		config.Spec.SidecarInjector.InitCNIConfiguration.Repair.BrokenPodLabelValue = utils.StrPointer(defaultInitCNIBrokenPodLabelValue)
	}
	if config.Spec.SidecarInjector.Init.Resources == nil {
		config.Spec.SidecarInjector.Init.Resources = defaultInitResources
	}

	// Proxy config
	if config.Spec.Proxy.Image == "" {
		config.Spec.Proxy.Image = defaultProxyImage
	}
	// Proxy Init config
	if config.Spec.ProxyInit.Image == "" {
		config.Spec.ProxyInit.Image = defaultProxyInitImage
	}
	if config.Spec.Proxy.AccessLogFile == nil {
		config.Spec.Proxy.AccessLogFile = utils.StrPointer(defaultEnvoyAccessLogFile)
	}
	if config.Spec.Proxy.AccessLogFormat == nil {
		config.Spec.Proxy.AccessLogFormat = utils.StrPointer(defaultEnvoyAccessLogFormat)
	}
	if config.Spec.Proxy.AccessLogEncoding == nil {
		config.Spec.Proxy.AccessLogEncoding = utils.StrPointer(defaultEnvoyAccessLogEncoding)
	}
	if config.Spec.Proxy.ComponentLogLevel == "" {
		config.Spec.Proxy.ComponentLogLevel = "misc:error"
	}
	if config.Spec.Proxy.LogLevel == "" {
		config.Spec.Proxy.LogLevel = "warning"
	}
	if config.Spec.Proxy.DNSRefreshRate == "" {
		config.Spec.Proxy.DNSRefreshRate = "300s"
	}
	if config.Spec.Proxy.EnvoyStatsD.Enabled == nil {
		config.Spec.Proxy.EnvoyStatsD.Enabled = utils.BoolPointer(false)
	}
	if config.Spec.Proxy.EnvoyMetricsService.Enabled == nil {
		config.Spec.Proxy.EnvoyMetricsService.Enabled = utils.BoolPointer(false)
	}
	if config.Spec.Proxy.EnvoyMetricsService.TLSSettings == nil {
		config.Spec.Proxy.EnvoyMetricsService.TLSSettings = &TLSSettings{
			Mode: "DISABLE",
		}
	}
	if config.Spec.Proxy.EnvoyMetricsService.TCPKeepalive == nil {
		config.Spec.Proxy.EnvoyMetricsService.TCPKeepalive = &TCPKeepalive{
			Probes:   3,
			Time:     "10s",
			Interval: "10s",
		}
	}
	if config.Spec.Proxy.EnvoyAccessLogService.Enabled == nil {
		config.Spec.Proxy.EnvoyAccessLogService.Enabled = utils.BoolPointer(false)
	}
	if config.Spec.Proxy.EnvoyAccessLogService.TLSSettings == nil {
		config.Spec.Proxy.EnvoyAccessLogService.TLSSettings = &TLSSettings{
			Mode: "DISABLE",
		}
	}
	if config.Spec.Proxy.EnvoyAccessLogService.TCPKeepalive == nil {
		config.Spec.Proxy.EnvoyAccessLogService.TCPKeepalive = &TCPKeepalive{
			Probes:   3,
			Time:     "10s",
			Interval: "10s",
		}
	}
	if config.Spec.Proxy.ProtocolDetectionTimeout == nil {
		config.Spec.Proxy.ProtocolDetectionTimeout = utils.StrPointer("100ms")
	}
	if config.Spec.Proxy.ClusterDomain == "" {
		config.Spec.Proxy.ClusterDomain = "cluster.local"
	}
	if config.Spec.Proxy.EnableCoreDump == nil {
		config.Spec.Proxy.EnableCoreDump = utils.BoolPointer(false)
	}
	if config.Spec.Proxy.CoreDumpImage == "" {
		config.Spec.Proxy.CoreDumpImage = defaultProxyCoreDumpImage
	}

	// PDB config
	if config.Spec.DefaultPodDisruptionBudget.Enabled == nil {
		config.Spec.DefaultPodDisruptionBudget.Enabled = utils.BoolPointer(false)
	}
	// Outbound traffic policy config
	if config.Spec.OutboundTrafficPolicy.Mode == "" {
		config.Spec.OutboundTrafficPolicy.Mode = outboundTrafficPolicyAllowAny
	}
	// Tracing config
	if config.Spec.Tracing.Enabled == nil {
		config.Spec.Tracing.Enabled = utils.BoolPointer(true)
	}
	if config.Spec.Tracing.Tracer == "" {
		config.Spec.Tracing.Tracer = TracerTypeZipkin
	}
	if config.Spec.Tracing.Zipkin.Address == "" {
		config.Spec.Tracing.Zipkin.Address = fmt.Sprintf(defaultZipkinAddress, config.Namespace)
	}
	if config.Spec.Tracing.Tracer == TracerTypeDatadog {
		if config.Spec.Tracing.Datadog.Address == "" {
			config.Spec.Tracing.Datadog.Address = "$(HOST_IP):8126"
		}
	}
	if config.Spec.Tracing.Tracer == TracerTypeStackdriver {
		if config.Spec.Tracing.Strackdriver.Debug == nil {
			config.Spec.Tracing.Strackdriver.Debug = utils.BoolPointer(false)
		}
		if config.Spec.Tracing.Strackdriver.MaxNumberOfAttributes == nil {
			config.Spec.Tracing.Strackdriver.MaxNumberOfAttributes = utils.IntPointer(200)
		}
		if config.Spec.Tracing.Strackdriver.MaxNumberOfAnnotations == nil {
			config.Spec.Tracing.Strackdriver.MaxNumberOfAnnotations = utils.IntPointer(200)
		}
		if config.Spec.Tracing.Strackdriver.MaxNumberOfMessageEvents == nil {
			config.Spec.Tracing.Strackdriver.MaxNumberOfMessageEvents = utils.IntPointer(200)
		}
	}

	// Multi mesh support
	if config.Spec.MultiMesh == nil {
		config.Spec.MultiMesh = utils.BoolPointer(false)
	}

	// Istio CoreDNS for multi mesh support
	if config.Spec.IstioCoreDNS.Enabled == nil {
		config.Spec.IstioCoreDNS.Enabled = utils.BoolPointer(false)
	}
	if config.Spec.IstioCoreDNS.Image == nil {
		config.Spec.IstioCoreDNS.Image = utils.StrPointer(defaultCoreDNSImage)
	}
	if config.Spec.IstioCoreDNS.PluginImage == "" {
		config.Spec.IstioCoreDNS.PluginImage = defaultCoreDNSPluginImage
	}
	if config.Spec.IstioCoreDNS.ReplicaCount == nil {
		config.Spec.IstioCoreDNS.ReplicaCount = utils.IntPointer(defaultReplicaCount)
	}

	if config.Spec.ImagePullPolicy == "" {
		config.Spec.ImagePullPolicy = defaultImagePullPolicy
	}

	if config.Spec.MeshExpansion == nil {
		config.Spec.MeshExpansion = utils.BoolPointer(false)
	}

	if config.Spec.UseMCP == nil {
		config.Spec.UseMCP = utils.BoolPointer(false)
	}

	if config.Spec.TrustDomain == "" {
		config.Spec.TrustDomain = "cluster.local"
	}

	if config.Spec.Proxy.UseMetadataExchangeFilter == nil {
		config.Spec.Proxy.UseMetadataExchangeFilter = utils.BoolPointer(false)
	}

	if config.Spec.JWTPolicy == "" {
		config.Spec.JWTPolicy = JWTPolicyFirstPartyJWT
	}
}

func SetRemoteIstioDefaults(remoteconfig *RemoteIstio) {
	if remoteconfig.Spec.IncludeIPRanges == "" {
		remoteconfig.Spec.IncludeIPRanges = defaultIncludeIPRanges
	}
	// SidecarInjector config
	if remoteconfig.Spec.SidecarInjector.ReplicaCount == nil {
		remoteconfig.Spec.SidecarInjector.ReplicaCount = utils.IntPointer(defaultReplicaCount)
	}
	if remoteconfig.Spec.Proxy.UseMetadataExchangeFilter == nil {
		remoteconfig.Spec.Proxy.UseMetadataExchangeFilter = utils.BoolPointer(false)
	}
}

func (in *MeshGateway) SetDefaults() {
	if in.Spec.ReplicaCount == nil {
		in.Spec.ReplicaCount = utils.IntPointer(defaultReplicaCount)
	}
	if in.Spec.MinReplicas == nil {
		in.Spec.MinReplicas = utils.IntPointer(defaultReplicaCount)
	}
	if in.Spec.MaxReplicas == nil {
		in.Spec.MaxReplicas = utils.IntPointer(defaultReplicaCount)
	}
	if in.Spec.Resources == nil {
		in.Spec.Resources = defaultProxyResources
	}

	if in.Spec.Type == GatewayTypeIngress && in.Spec.ServiceType == "" {
		in.Spec.ServiceType = defaultIngressGatewayServiceType
	}
	if in.Spec.Type == GatewayTypeEgress && in.Spec.ServiceType == "" {
		in.Spec.ServiceType = defaultEgressGatewayServiceType
	}
}
