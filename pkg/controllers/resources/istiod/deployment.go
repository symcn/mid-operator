package istiod

import (
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/controllers/resources"
	"github.com/symcn/mid-operator/pkg/controllers/resources/base"
	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) containerArgs() []string {
	containerArgs := []string{
		"discovery",
		"--monitoringAddr=:15014",
		"--domain",
		r.Config.Spec.Proxy.ClusterDomain,
		"--keepaliveMaxServerConnectionAge",
		"30m",
		"--trust-domain",
		r.Config.Spec.TrustDomain,
		"--disable-install-crds=true",
	}

	if r.Config.Spec.Logging.Level != nil {
		containerArgs = append(containerArgs, fmt.Sprintf("--log_output_level=%s", utils.PointerToString(r.Config.Spec.Logging.Level)))
	}

	if r.Config.Spec.ControlPlaneSecurityEnabled && !utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		containerArgs = append(containerArgs, "--secureGrpcAddr", ":15011")
	} else {
		containerArgs = append(containerArgs, "--secureGrpcAddr", "")
	}

	if r.Config.Spec.WatchOneNamespace {
		containerArgs = append(containerArgs, "-a", r.Config.Namespace)
	}

	if len(r.Config.Spec.Pilot.AdditionalContainerArgs) != 0 {
		containerArgs = append(containerArgs, r.Config.Spec.Pilot.AdditionalContainerArgs...)
	}

	return containerArgs
}

func (r *Reconciler) containerEnvs() []corev1.EnvVar {
	envs := []corev1.EnvVar{
		{
			Name:  "PILOT_PUSH_THROTTLE",
			Value: "100",
		},
		{
			Name:  "PILOT_TRACE_SAMPLING",
			Value: fmt.Sprintf("%d", r.Config.Spec.Pilot.TraceSampling),
		},
		{
			Name:  "MESHNETWORKS_HASH",
			Value: r.Config.Spec.GetMeshNetworksHash(),
		},
		{
			Name:  "PILOT_ENABLE_PROTOCOL_SNIFFING_FOR_OUTBOUND",
			Value: strconv.FormatBool(utils.PointerToBool(r.Config.Spec.Pilot.EnableProtocolSniffingOutbound)),
		},
		{
			Name:  "PILOT_ENABLE_PROTOCOL_SNIFFING_FOR_INBOUND",
			Value: strconv.FormatBool(utils.PointerToBool(r.Config.Spec.Pilot.EnableProtocolSniffingInbound)),
		},
		{
			Name:  "INJECTION_WEBHOOK_CONFIG_NAME",
			Value: "istio-sidecar-injector",
		},
	}

	envs = append(envs, templates.IstioProxyEnv(r.Config)...)

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		envs = append(envs, []corev1.EnvVar{
			{
				Name:  "ISTIOD_ADDR",
				Value: resources.GetDiscoveryAddress(r.Config, "istiod"),
			},
			{
				Name:  "PILOT_EXTERNAL_GALLEY",
				Value: "false",
			},
		}...)
	}

	if r.Config.Spec.LocalityLB != nil && utils.PointerToBool(r.Config.Spec.LocalityLB.Enabled) {
		envs = append(envs, corev1.EnvVar{
			Name:  "PILOT_ENABLE_LOCALITY_LOAD_BALANCING",
			Value: "1",
		})
	}

	envs = k8sutils.MergeEnvVars(envs, r.Config.Spec.Pilot.AdditionalEnvVars)

	return envs
}

func (r *Reconciler) containerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
		{ContainerPort: 15010, Protocol: corev1.ProtocolTCP},
		{ContainerPort: 15017, Protocol: corev1.ProtocolTCP},
	}
}

func (r *Reconciler) proxyVolumeMounts() []corev1.VolumeMount {
	vms := []corev1.VolumeMount{
		{
			Name:      "pilot-envoy-config",
			MountPath: "/var/lib/envoy",
		},
	}

	if r.Config.Spec.ControlPlaneSecurityEnabled && utils.PointerToBool(r.Config.Spec.MountMtlsCerts) {
		vms = append(vms, corev1.VolumeMount{
			Name:      "istio-certs",
			MountPath: "/etc/certs",
			ReadOnly:  true,
		})
	}

	return vms
}

func (r *Reconciler) containers() []corev1.Container {
	discoveryContainer := corev1.Container{
		Name:            "discovery",
		Image:           utils.PointerToString(r.Config.Spec.Pilot.Image),
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Args:            r.containerArgs(),
		Ports:           r.containerPorts(),
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/ready",
					Port:   intstr.FromInt(8080),
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       5,
			TimeoutSeconds:      5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
		},
		Env: r.containerEnvs(),
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.Pilot.Resources,
			r.Config.Spec.DefaultResources,
		),
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:    utils.Int64Pointer(1337),
			RunAsGroup:   utils.Int64Pointer(1337),
			RunAsNonRoot: utils.BoolPointer(true),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{
					"ALL",
				},
			},
		},
		VolumeMounts:             r.volumeMounts(),
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}

	containers := []corev1.Container{
		discoveryContainer,
	}

	args := []string{
		"proxy",
		"--serviceCluster",
		"istio-pilot",
		"--templateFile",
		"/var/lib/envoy/envoy.yaml.tmpl",
		"--controlPlaneAuthPolicy",
		templates.ControlPlaneAuthPolicy(utils.PointerToBool(r.Config.Spec.Istiod.Enabled), r.Config.Spec.ControlPlaneSecurityEnabled),
		"--domain",
		r.Config.Namespace + ".svc." + r.Config.Spec.Proxy.ClusterDomain,
		"--trust-domain",
		r.Config.Spec.TrustDomain,
	}
	if r.Config.Spec.Proxy.LogLevel != "" {
		args = append(args, fmt.Sprintf("--proxyLogLevel=%s", r.Config.Spec.Proxy.LogLevel))
	}
	if r.Config.Spec.Proxy.ComponentLogLevel != "" {
		args = append(args, fmt.Sprintf("--proxyComponentLogLevel=%s", r.Config.Spec.Proxy.ComponentLogLevel))
	}
	if r.Config.Spec.Logging.Level != nil {
		args = append(args, fmt.Sprintf("--log_output_level=%s", utils.PointerToString(r.Config.Spec.Logging.Level)))
	}

	if r.Config.Spec.ControlPlaneSecurityEnabled && !utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		proxyContainer := corev1.Container{
			Name:            "istio-proxy",
			Image:           r.Config.Spec.Proxy.Image,
			ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
			Ports: []corev1.ContainerPort{
				{ContainerPort: 15011, Protocol: corev1.ProtocolTCP},
			},
			Args: args,
			Env:  templates.IstioProxyEnv(r.Config),
			Resources: templates.GetResourcesRequirementsOrDefault(
				r.Config.Spec.Proxy.Resources,
				r.Config.Spec.DefaultResources,
			),
			VolumeMounts:             r.proxyVolumeMounts(),
			TerminationMessagePath:   corev1.TerminationMessagePathDefault,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		}

		if r.Config.Spec.Proxy.LogLevel != "" {
			proxyContainer.Args = append(proxyContainer.Args, fmt.Sprintf("--proxyLogLevel=%s", r.Config.Spec.Proxy.LogLevel))
		}
		if r.Config.Spec.Proxy.ComponentLogLevel != "" {
			proxyContainer.Args = append(proxyContainer.Args, fmt.Sprintf("--proxyComponentLogLevel=%s", r.Config.Spec.Proxy.ComponentLogLevel))
		}

		containers = append(containers, proxyContainer)
	}

	return containers
}

func (r *Reconciler) volumeMounts() []corev1.VolumeMount {
	vms := []corev1.VolumeMount{
		{
			Name:      "config-volume",
			MountPath: "/etc/istio/config",
		},
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		if r.Config.Spec.JWTPolicy == v1beta1.JWTPolicyThirdPartyJWT {
			vms = append(vms, corev1.VolumeMount{
				Name:      "istio-token",
				MountPath: "/var/run/secrets/tokens",
				ReadOnly:  true,
			})
		}

		vms = append(vms, []corev1.VolumeMount{
			{
				Name:      "local-certs",
				MountPath: "/var/run/secrets/istio-dns",
			},
			{
				Name:      "cacerts",
				MountPath: "/etc/cacerts",
				ReadOnly:  true,
			},
			{
				Name:      "inject",
				MountPath: "/var/lib/istio/inject",
				ReadOnly:  true,
			},
			{
				Name:      "istiod",
				MountPath: "/var/lib/istio/local",
				ReadOnly:  true,
			},
		}...)
	}

	return vms
}

func (r *Reconciler) volumes() []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "config-volume",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: base.IstioConfigMapName,
					},
					DefaultMode: utils.IntPointer(420),
				},
			},
		},
		{
			Name: "pilot-envoy-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapNameEnvoy,
					},
					DefaultMode: utils.IntPointer(420),
				},
			},
		},
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		volumes = append(volumes, corev1.Volume{
			Name: "local-certs",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		})

		if r.Config.Spec.JWTPolicy == v1beta1.JWTPolicyThirdPartyJWT {
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

		volumes = append(volumes, []corev1.Volume{
			{
				Name: "istiod",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "istiod",
						},
						Optional:    utils.BoolPointer(true),
						DefaultMode: utils.IntPointer(420),
					},
				},
			},
			{
				Name: "cacerts",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "cacerts",
						Optional:    utils.BoolPointer(true),
						DefaultMode: utils.IntPointer(420),
					},
				},
			},
			{
				Name: "inject",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "istio-sidecar-injector",
						},
						Optional:    utils.BoolPointer(true),
						DefaultMode: utils.IntPointer(420),
					},
				},
			},
		}...)
	}

	if r.Config.Spec.ControlPlaneSecurityEnabled && utils.PointerToBool(r.Config.Spec.MountMtlsCerts) {
		volumes = append(volumes, corev1.Volume{
			Name: "istio-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("istio.%s", serviceAccountName),
					Optional:    utils.BoolPointer(true),
					DefaultMode: utils.IntPointer(420),
				},
			},
		})
	}

	return volumes
}

func (r *Reconciler) deployment() runtime.Object {
	deployment := &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, utils.MergeStringMaps(istiodLabels, pilotLabelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: utils.IntPointer(k8sutils.GetHPAReplicaCountOrDefault(r.Client, types.NamespacedName{
				Name:      hpaName,
				Namespace: r.Config.Namespace,
			}, utils.PointerToInt32(r.Config.Spec.Pilot.ReplicaCount))),
			Strategy: templates.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: pilotLabelSelector,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      utils.MergeStringMaps(istiodLabels, pilotLabelSelector),
					Annotations: utils.MergeStringMaps(templates.DefaultDeployAnnotations(), r.Config.Spec.Pilot.PodAnnotations),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: utils.Int64Pointer(1337),
					},
					Containers:   r.containers(),
					Volumes:      r.volumes(),
					Affinity:     r.Config.Spec.Pilot.Affinity,
					NodeSelector: r.Config.Spec.Pilot.NodeSelector,
					Tolerations:  r.Config.Spec.Pilot.Tolerations,
				},
			},
		},
	}

	return deployment
}
