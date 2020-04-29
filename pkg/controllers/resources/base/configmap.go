package base

import (
	"fmt"

	"github.com/ghodss/yaml"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"encoding/json"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/controllers/resources"
	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
)

var cmLabels = map[string]string{
	"app": "istio",
}

func GetData(in *devopsv1beta1.EnvoyServiceCommonConfiguration) map[string]interface{} {
	data := map[string]interface{}{
		"address": fmt.Sprintf("%s:%d", in.Host, in.Port),
	}
	if in.TLSSettings != nil {
		data["tlsSettings"] = in.TLSSettings
	}
	if in.TCPKeepalive != nil {
		data["tcpKeepalive"] = in.TCPKeepalive
	}

	return data
}

func GetDataJSON(in *devopsv1beta1.EnvoyServiceCommonConfiguration) string {
	j, err := json.Marshal(GetData(in))
	if err != nil {
		return ""
	}

	return string(j)
}

func (r *Reconciler) configMap() runtime.Object {
	return &apiv1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(IstioConfigMapName, cmLabels, r.Config),
		Data: map[string]string{
			"mesh":         r.meshConfig(),
			"meshNetworks": r.meshNetworks(),
		},
	}
}

func (r *Reconciler) meshConfig() string {
	defaultConfig := map[string]interface{}{
		"connectTimeout":         "10s",
		"configPath":             "/etc/istio/proxy",
		"binaryPath":             "/usr/local/bin/envoy",
		"serviceCluster":         "istio-proxy",
		"drainDuration":          "45s",
		"parentShutdownDuration": "1m0s",
		"proxyAdminPort":         15000,
		"concurrency":            0,
		"controlPlaneAuthPolicy": templates.ControlPlaneAuthPolicy(utils.PointerToBool(r.Config.Spec.Istiod.Enabled), r.Config.Spec.ControlPlaneSecurityEnabled),
		"discoveryAddress":       resources.GetDiscoveryAddress(r.Config),
	}

	if utils.PointerToBool(r.Config.Spec.Proxy.EnvoyStatsD.Enabled) {
		defaultConfig["statsdUdpAddress"] = fmt.Sprintf("%s:%d", r.Config.Spec.Proxy.EnvoyStatsD.Host, r.Config.Spec.Proxy.EnvoyStatsD.Port)
	}

	if utils.PointerToBool(r.Config.Spec.Proxy.EnvoyMetricsService.Enabled) {
		defaultConfig["envoyAccessLogService"] = GetData(&r.Config.Spec.Proxy.EnvoyMetricsService)
	}

	if utils.PointerToBool(r.Config.Spec.Proxy.EnvoyAccessLogService.Enabled) {
		defaultConfig["envoyAccessLogService"] = GetData(&r.Config.Spec.Proxy.EnvoyMetricsService)
	}

	if utils.PointerToBool(r.Config.Spec.Tracing.Enabled) {
		switch r.Config.Spec.Tracing.Tracer {
		case devopsv1beta1.TracerTypeZipkin:
			defaultConfig["tracing"] = map[string]interface{}{
				"zipkin": map[string]interface{}{
					"address": r.Config.Spec.Tracing.Zipkin.Address,
				},
			}
		case devopsv1beta1.TracerTypeLightstep:
			lightStep := map[string]interface{}{
				"address":     r.Config.Spec.Tracing.Lightstep.Address,
				"accessToken": r.Config.Spec.Tracing.Lightstep.AccessToken,
				"secure":      r.Config.Spec.Tracing.Lightstep.Secure,
			}
			if r.Config.Spec.Tracing.Lightstep.Secure {
				lightStep["cacertPath"] = r.Config.Spec.Tracing.Lightstep.CacertPath
			}
			defaultConfig["tracing"] = map[string]interface{}{
				"lightstep": lightStep,
			}
		case devopsv1beta1.TracerTypeDatadog:
			defaultConfig["tracing"] = map[string]interface{}{
				"datadog": map[string]interface{}{
					"address": r.Config.Spec.Tracing.Datadog.Address,
				},
			}
		case devopsv1beta1.TracerTypeStackdriver:
			defaultConfig["tracing"] = map[string]interface{}{
				"stackdriver": r.Config.Spec.Tracing.Strackdriver,
			}
		}
	}

	meshConfig := map[string]interface{}{
		"disablePolicyChecks":     false,
		"disableMixerHttpReports": false,
		"enableTracing":           r.Config.Spec.Tracing.Enabled,
		"accessLogFile":           r.Config.Spec.Proxy.AccessLogFile,
		"accessLogFormat":         r.Config.Spec.Proxy.AccessLogFormat,
		"accessLogEncoding":       r.Config.Spec.Proxy.AccessLogEncoding,
		"policyCheckFailOpen":     false,
		"ingressService":          "istio-ingressgateway",
		"ingressClass":            "istio",
		"ingressControllerMode":   2,
		"trustDomain":             r.Config.Spec.TrustDomain,
		"trustDomainAliases":      r.Config.Spec.TrustDomainAliases,
		"enableAutoMtls":          utils.PointerToBool(r.Config.Spec.AutoMTLS),
		"outboundTrafficPolicy": map[string]interface{}{
			"mode": r.Config.Spec.OutboundTrafficPolicy.Mode,
		},
		"defaultConfig":               defaultConfig,
		"rootNamespace":               r.Config.Namespace,
		"connectTimeout":              "10s",
		"localityLbSetting":           r.getLocalityLBConfiguration(),
		"enableEnvoyAccessLogService": utils.PointerToBool(r.Config.Spec.Proxy.EnvoyAccessLogService.Enabled),
		"protocolDetectionTimeout":    r.Config.Spec.Proxy.ProtocolDetectionTimeout,
		"dnsRefreshRate":              r.Config.Spec.Proxy.DNSRefreshRate,
		"certificates":                r.Config.Spec.Certificates,
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		meshConfig["sdsUdsPath"] = "unix:/etc/istio/proxy/SDS"
	}

	if utils.PointerToBool(r.Config.Spec.UseMCP) {
		meshConfig["configSources"] = []map[string]interface{}{
			r.defaultConfigSource(),
		}
	}

	marshaledConfig, _ := yaml.Marshal(meshConfig)
	return string(marshaledConfig)
}

func (r *Reconciler) getLocalityLBConfiguration() *devopsv1beta1.LocalityLBConfiguration {
	var localityLbConfiguration *devopsv1beta1.LocalityLBConfiguration

	if r.Config.Spec.LocalityLB == nil || !utils.PointerToBool(r.Config.Spec.LocalityLB.Enabled) {
		return localityLbConfiguration
	}

	if r.Config.Spec.LocalityLB != nil {
		localityLbConfiguration = r.Config.Spec.LocalityLB.DeepCopy()
		localityLbConfiguration.Enabled = nil
		if localityLbConfiguration.Distribute != nil && localityLbConfiguration.Failover != nil {
			localityLbConfiguration.Failover = nil
		}
	}

	return localityLbConfiguration
}

func (r *Reconciler) meshNetworks() string {
	marshaledConfig, _ := yaml.Marshal(r.Config.Spec.GetMeshNetworks())
	return string(marshaledConfig)
}

func (r *Reconciler) mixerServer(mixerType string) string {
	if r.remote {
		return fmt.Sprintf("istio-%s.%s:%s", mixerType, r.Config.Namespace, "15004")
	}
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		return fmt.Sprintf("istio-%s.%s.svc.%s:%s", mixerType, r.Config.Namespace, r.Config.Spec.Proxy.ClusterDomain, "15004")
	}
	return fmt.Sprintf("istio-%s.%s.svc.%s:%s", mixerType, r.Config.Namespace, r.Config.Spec.Proxy.ClusterDomain, "9091")
}

func (r *Reconciler) defaultConfigSource() map[string]interface{} {
	cs := map[string]interface{}{
		"address": fmt.Sprintf("istio-galley.%s.svc:9901", r.Config.Namespace),
	}
	if r.Config.Spec.ControlPlaneSecurityEnabled {
		cs["tlsSettings"] = map[string]interface{}{
			"mode": "ISTIO_MUTUAL",
		}
	}
	return cs
}
