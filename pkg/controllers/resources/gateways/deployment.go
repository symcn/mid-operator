package gateways

import (
	"encoding/json"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/controllers/resources"
	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) deployment() runtime.Object {
	var initContainers []corev1.Container
	if utils.PointerToBool(r.Config.Spec.Proxy.EnableCoreDump) && r.Config.Spec.Proxy.CoreDumpImage != "" {
		initContainers = []corev1.Container{GetCoreDumpContainer(r.Config)}
	}

	var containers = make([]corev1.Container, 0)

	args := []string{
		"proxy",
		"router",
		"--domain", fmt.Sprintf("$(POD_NAMESPACE).svc.%s", r.Config.Spec.Proxy.ClusterDomain),
		"--log_output_level", "info",
		"--drainDuration", "45s",
		"--parentShutdownDuration", "1m0s",
		"--connectTimeout", "10s",
		"--serviceCluster", r.gw.Name,
		"--proxyAdminPort", "15000",
		"--statusPort", "15020",
		"--controlPlaneAuthPolicy", templates.ControlPlaneAuthPolicy(utils.PointerToBool(r.Config.Spec.Istiod.Enabled), r.Config.Spec.ControlPlaneSecurityEnabled),
		"--discoveryAddress", resources.GetDiscoveryAddress(r.Config),
		"--trust-domain", r.Config.Spec.TrustDomain,
	}

	if utils.PointerToBool(r.Config.Spec.Tracing.Enabled) {
		if r.Config.Spec.Tracing.Tracer == devopsv1beta1.TracerTypeLightstep {
			args = append(args, "--lightstepAddress", r.Config.Spec.Tracing.Lightstep.Address)
			args = append(args, "--lightstepAccessToken", r.Config.Spec.Tracing.Lightstep.AccessToken)
			args = append(args, fmt.Sprintf("--lightstepSecure=%t", r.Config.Spec.Tracing.Lightstep.Secure))
			args = append(args, "--lightstepCacertPath", r.Config.Spec.Tracing.Lightstep.CacertPath)
		} else if r.Config.Spec.Tracing.Tracer == devopsv1beta1.TracerTypeZipkin {
			args = append(args, "--zipkinAddress", r.Config.Spec.Tracing.Zipkin.Address)
		} else if r.Config.Spec.Tracing.Tracer == devopsv1beta1.TracerTypeDatadog {
			args = append(args, "--datadogAgentAddress", r.Config.Spec.Tracing.Datadog.Address)
		}
	}

	if r.Config.Spec.Proxy.LogLevel != "" {
		args = append(args, "--proxyLogLevel", r.Config.Spec.Proxy.LogLevel)
	}

	if r.Config.Spec.Proxy.ComponentLogLevel != "" {
		args = append(args, "--proxyComponentLogLevel", r.Config.Spec.Proxy.ComponentLogLevel)
	}

	if r.Config.Spec.Logging.Level != nil {
		args = append(args, fmt.Sprintf("--log_output_level=%s", utils.PointerToString(r.Config.Spec.Logging.Level)))
	}

	if utils.PointerToBool(r.Config.Spec.Proxy.EnvoyMetricsService.Enabled) {
		envoyMetricsServiceJSON, err := r.getEnvoyServiceConfigurationJSON(r.Config.Spec.Proxy.EnvoyMetricsService)
		if err == nil {
			args = append(args, "--envoyMetricsService", fmt.Sprintf("%s", string(envoyMetricsServiceJSON)))
		}
	}

	if utils.PointerToBool(r.Config.Spec.Proxy.EnvoyAccessLogService.Enabled) {
		envoyAccessLogServiceJSON, err := r.getEnvoyServiceConfigurationJSON(r.Config.Spec.Proxy.EnvoyAccessLogService)
		if err == nil {
			args = append(args, "--envoyAccessLogService", fmt.Sprintf("%s", string(envoyAccessLogServiceJSON)))
		}
	}

	containers = append(containers, corev1.Container{
		Name:            "istio-proxy",
		Image:           r.Config.Spec.Proxy.Image,
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Args:            args,
		Ports:           r.ports(),
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/healthz/ready",
					Port:   intstr.FromInt(15020),
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 1,
			PeriodSeconds:       2,
			FailureThreshold:    30,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
		},
		Env: append(templates.IstioProxyEnv(r.Config), r.envVars()...),
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.gw.Spec.Resources,
			r.Config.Spec.Proxy.Resources,
		),
		VolumeMounts:             r.volumeMounts(),
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	})

	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(r.gatewayName(), r.labels(), r.gw),
		Spec: appsv1.DeploymentSpec{
			Replicas: utils.IntPointer(k8sutils.GetHPAReplicaCountOrDefault(r.Client, types.NamespacedName{
				Name:      r.hpaName(),
				Namespace: r.Config.Namespace,
			}, *r.gw.Spec.ReplicaCount)),
			Selector: &metav1.LabelSelector{
				MatchLabels: r.labels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.labels(),
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: r.serviceAccountName(),
					InitContainers:     initContainers,
					Containers:         containers,
					Volumes:            r.volumes(),
					Affinity:           r.gw.Spec.Affinity,
					NodeSelector:       r.gw.Spec.NodeSelector,
					Tolerations:        r.gw.Spec.Tolerations,
					PriorityClassName:  r.Config.Spec.PriorityClassName,
				},
			},
		},
	}
}

func (r *Reconciler) getEnvoyServiceConfigurationJSON(config devopsv1beta1.EnvoyServiceCommonConfiguration) (string, error) {
	type Properties struct {
		Address      string                      `json:"address,omitempty"`
		TLSSettings  *devopsv1beta1.TLSSettings  `json:"tlsSettings,omitempty"`
		TCPKeepalive *devopsv1beta1.TCPKeepalive `json:"tcpKeepalive,omitempty"`
	}

	properties := Properties{
		Address:      fmt.Sprintf("%s:%d", config.Host, config.Port),
		TLSSettings:  config.TLSSettings,
		TCPKeepalive: config.TCPKeepalive,
	}

	data, err := json.Marshal(properties)
	if err != nil {
		return "", err
	}

	return string(data), err
}

func (r *Reconciler) ports() []corev1.ContainerPort {
	var ports []corev1.ContainerPort
	for _, port := range r.gw.Spec.Ports {
		ports = append(ports, corev1.ContainerPort{
			ContainerPort: port.Port, Protocol: port.Protocol, Name: port.Name,
		})
	}
	ports = append(ports, corev1.ContainerPort{
		ContainerPort: 15090, Protocol: corev1.ProtocolTCP, Name: "http-envoy-prom",
	})
	return ports
}

func (r *Reconciler) envVars() []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			},
		},
		{
			Name: "SERVICE_ACCOUNT",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath:  "spec.serviceAccountName",
					APIVersion: "v1",
				},
			},
		},
		{
			Name: "ISTIO_META_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath:  "metadata.name",
					APIVersion: "v1",
				},
			},
		},
		{
			Name: "ISTIO_META_CONFIG_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath:  "metadata.namespace",
					APIVersion: "v1",
				},
			},
		},
		{
			Name:  "ISTIO_META_ROUTER_MODE",
			Value: "sni-dnat",
		},
		{
			Name: "NODE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath:  "spec.nodeName",
					APIVersion: "v1",
				},
			},
		},
		{
			Name:  "ISTIO_META_WORKLOAD_NAME",
			Value: r.gatewayName(),
		},
		{
			Name:  "ISTIO_META_OWNER",
			Value: fmt.Sprintf("kubernetes://apis/apps/v1/namespaces/%s/deployments/%s", r.Config.Namespace, r.gatewayName()),
		},
	}

	if r.gw.Spec.Type == devopsv1beta1.GatewayTypeIngress && (utils.PointerToBool(r.Config.Spec.Istiod.Enabled)) {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "ISTIO_META_USER_SDS",
			Value: "true",
		})
	}

	if r.gw.Spec.Type == devopsv1beta1.GatewayTypeIngress && utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "CA_ADDR",
			Value: r.Config.Spec.CAAddress,
		})
	}

	if r.Config.Spec.ClusterName != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "ISTIO_META_CLUSTER_ID",
			Value: r.Config.Spec.ClusterName,
		})
	} else {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "ISTIO_META_CLUSTER_ID",
			Value: "Kubernetes",
		})
	}

	if r.Config.Spec.NetworkName != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "ISTIO_META_NETWORK",
			Value: r.Config.Spec.NetworkName,
		})
	}
	if r.gw.Spec.RequestedNetworkView != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "ISTIO_META_REQUESTED_NETWORK_VIEW",
			Value: r.gw.Spec.RequestedNetworkView,
		})
	}
	if utils.PointerToBool(r.Config.Spec.AutoMTLS) {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "ISTIO_AUTO_MTLS_ENABLED",
			Value: "true",
		})
	}

	if r.Config.Spec.MeshID != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "ISTIO_META_MESH_ID",
			Value: r.Config.Spec.MeshID,
		})
	} else if r.Config.Spec.TrustDomain != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "ISTIO_META_MESH_ID",
			Value: r.Config.Spec.TrustDomain,
		})
	}

	if utils.PointerToBool(r.Config.Spec.Tracing.Enabled) {
		if r.Config.Spec.Tracing.Tracer == devopsv1beta1.TracerTypeDatadog {
			envVars = append(envVars, corev1.EnvVar{
				Name: "HOST_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			})
		} else if r.Config.Spec.Tracing.Tracer == devopsv1beta1.TracerTypeStackdriver {
			envVars = append(envVars, corev1.EnvVar{
				Name:  "STACKDRIVER_TRACING_ENABLED",
				Value: "true",
			})
			envVars = append(envVars, corev1.EnvVar{
				Name:  "STACKDRIVER_TRACING_DEBUG",
				Value: strconv.FormatBool(utils.PointerToBool(r.Config.Spec.Tracing.Strackdriver.Debug)),
			})
			if r.Config.Spec.Tracing.Strackdriver.MaxNumberOfAnnotations != nil {
				envVars = append(envVars, corev1.EnvVar{
					Name:  "STACKDRIVER_TRACING_MAX_NUMBER_OF_ANNOTATIONS",
					Value: string(utils.PointerToInt32(r.Config.Spec.Tracing.Strackdriver.MaxNumberOfAnnotations)),
				})
			}
			if r.Config.Spec.Tracing.Strackdriver.MaxNumberOfAttributes != nil {
				envVars = append(envVars, corev1.EnvVar{
					Name:  "STACKDRIVER_TRACING_MAX_NUMBER_OF_ATTRIBUTES",
					Value: string(utils.PointerToInt32(r.Config.Spec.Tracing.Strackdriver.MaxNumberOfAttributes)),
				})
			}
			if r.Config.Spec.Tracing.Strackdriver.MaxNumberOfMessageEvents != nil {
				envVars = append(envVars, corev1.EnvVar{
					Name:  "STACKDRIVER_TRACING_MAX_NUMBER_OF_MESSAGE_EVENTS",
					Value: string(utils.PointerToInt32(r.Config.Spec.Tracing.Strackdriver.MaxNumberOfMessageEvents)),
				})
			}
		}
	}

	envVars = k8sutils.MergeEnvVars(envVars, r.gw.Spec.AdditionalEnvVars)

	return envVars
}

func (r *Reconciler) volumeMounts() []corev1.VolumeMount {
	vms := []corev1.VolumeMount{
		{
			Name:      fmt.Sprintf("%s-certs", r.gw.Name),
			MountPath: fmt.Sprintf("/etc/istio/%s-certs", r.gw.Spec.Type+"gateway"),
			ReadOnly:  true,
		},
		{
			Name:      fmt.Sprintf("%s-ca-certs", r.gw.Name),
			MountPath: fmt.Sprintf("/etc/istio/%s-ca-certs", r.gw.Spec.Type+"gateway"),
			ReadOnly:  true,
		},
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) && r.Config.Spec.Pilot.CertProvider == devopsv1beta1.PilotCertProviderTypeIstiod {
		vms = append(vms, corev1.VolumeMount{
			Name:      "istiod-ca-cert",
			MountPath: "/var/run/secrets/istio",
		})
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) && r.Config.Spec.JWTPolicy == devopsv1beta1.JWTPolicyThirdPartyJWT {
		vms = append(vms, corev1.VolumeMount{
			Name:      "istio-token",
			MountPath: "/var/run/secrets/tokens",
			ReadOnly:  true,
		})
	}

	if r.gw.Spec.Type == devopsv1beta1.GatewayTypeIngress && utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		vms = append(vms, corev1.VolumeMount{
			Name:      "ingressgatewaysdsudspath",
			MountPath: "/var/run/ingress_gateway",
		})
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) && utils.PointerToBool(r.Config.Spec.MountMtlsCerts) {
		vms = append(vms, corev1.VolumeMount{
			Name:      "istio-certs",
			MountPath: "/etc/certs",
			ReadOnly:  true,
		})
	}

	vms = append(vms, corev1.VolumeMount{
		Name:      "podinfo",
		MountPath: "/etc/istio/pod",
	})

	return vms
}

func (r *Reconciler) volumes() []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: fmt.Sprintf("%s-certs", r.gw.Name),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%s-certs", r.gw.Name),
					Optional:    utils.BoolPointer(true),
					DefaultMode: utils.IntPointer(420),
				},
			},
		},
		{
			Name: fmt.Sprintf("%s-ca-certs", r.gw.Name),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%s-ca-certs", r.gw.Name),
					Optional:    utils.BoolPointer(true),
					DefaultMode: utils.IntPointer(420),
				},
			},
		},
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) && r.Config.Spec.Pilot.CertProvider == devopsv1beta1.PilotCertProviderTypeIstiod {
		volumes = append(volumes, corev1.Volume{
			Name: "istiod-ca-cert",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "istio-ca-root-cert",
					},
				},
			},
		})
	}

	volumes = append(volumes, corev1.Volume{
		Name: "podinfo",
		VolumeSource: corev1.VolumeSource{
			DownwardAPI: &corev1.DownwardAPIVolumeSource{
				DefaultMode: utils.IntPointer(420),
				Items: []corev1.DownwardAPIVolumeFile{
					{
						Path: "labels",
						FieldRef: &corev1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.labels",
						},
					},
					{
						Path: "annotations",
						FieldRef: &corev1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.annotations",
						},
					},
				},
			},
		},
	})

	if r.gw.Spec.Type == devopsv1beta1.GatewayTypeIngress && utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		volumes = append(volumes, corev1.Volume{
			Name: "ingressgatewaysdsudspath",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) && r.Config.Spec.JWTPolicy == devopsv1beta1.JWTPolicyThirdPartyJWT {
		volumes = append(volumes, corev1.Volume{
			Name: "istio-token",
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: []corev1.VolumeProjection{
						{
							ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
								// Audience:          r.Config.Spec.SDS.TokenAudience,
								ExpirationSeconds: utils.Int64Pointer(43200),
								Path:              "istio-token",
							},
						},
					},
				},
			},
		})
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) && utils.PointerToBool(r.Config.Spec.MountMtlsCerts) {
		volumes = append(volumes, corev1.Volume{
			Name: "istio-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: fmt.Sprintf("istio.%s", r.serviceAccountName()),
					Optional:   utils.BoolPointer(true),
				},
			},
		})
	}

	return volumes
}

// GetCoreDumpContainer get core dump init container for Envoy proxies
func GetCoreDumpContainer(config *devopsv1beta1.Istio) corev1.Container {
	return corev1.Container{
		Name:            "enable-core-dump",
		Image:           config.Spec.Proxy.CoreDumpImage,
		ImagePullPolicy: config.Spec.ImagePullPolicy,
		Command: []string{
			"/bin/sh",
		},
		Args: []string{
			"-c",
			"sysctl -w kernel.core_pattern=/var/lib/istio/core.proxy && ulimit -c unlimited",
		},
		Resources: templates.GetResourcesRequirementsOrDefault(config.Spec.SidecarInjector.Init.Resources, config.Spec.DefaultResources),
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: utils.BoolPointer(true),
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{
					"SYS_ADMIN",
				},
				Drop: []corev1.Capability{
					"ALL",
				},
			},
			Privileged:             utils.BoolPointer(true),
			ReadOnlyRootFilesystem: utils.BoolPointer(false),
			RunAsGroup:             utils.Int64Pointer(0),
			RunAsNonRoot:           utils.BoolPointer(false),
			RunAsUser:              utils.Int64Pointer(0),
		},
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}
}
