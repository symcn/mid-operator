package istiocoredns

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) coreDNSContainer() corev1.Container {
	return corev1.Container{
		Name:            "coredns",
		Image:           utils.PointerToString(r.Config.Spec.IstioCoreDNS.Image),
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Args: []string{
			"-conf",
			"/etc/coredns/Corefile",
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config-volume",
				MountPath: "/etc/coredns",
				ReadOnly:  true,
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "dns",
				ContainerPort: 53,
				Protocol:      corev1.ProtocolUDP,
			},
			{
				Name:          "dns-tcp",
				ContainerPort: 53,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "metrics",
				ContainerPort: 9153,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/health",
					Port:   intstr.FromInt(8080),
					Scheme: corev1.URISchemeHTTP,
				},
			},
			InitialDelaySeconds: 60,
			PeriodSeconds:       5,
			FailureThreshold:    5,
			SuccessThreshold:    1,
			TimeoutSeconds:      5,
		},
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.IstioCoreDNS.Resources,
			r.Config.Spec.DefaultResources,
		),
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}
}

func (r *Reconciler) coreDNSPluginContainer() corev1.Container {
	return corev1.Container{
		Name:            "istio-coredns-plugin",
		Image:           r.Config.Spec.IstioCoreDNS.PluginImage,
		ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
		Command: []string{
			"/usr/local/bin/plugin",
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "dns-grpc",
				ContainerPort: 8053,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Resources: templates.GetResourcesRequirementsOrDefault(
			r.Config.Spec.IstioCoreDNS.Resources,
			r.Config.Spec.DefaultResources,
		),
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}
}

func (r *Reconciler) deployment() runtime.Object {
	return &appsv1.Deployment{
		ObjectMeta: templates.ObjectMeta(deploymentName, utils.MergeStringMaps(labels, labelSelector), r.Config),
		Spec: appsv1.DeploymentSpec{
			Replicas: r.Config.Spec.IstioCoreDNS.ReplicaCount,
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       utils.IntstrPointer(1),
					MaxUnavailable: utils.IntstrPointer(0),
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: utils.MergeStringMaps(labels, labelSelector),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      utils.MergeStringMaps(labels, labelSelector),
					Annotations: utils.MergeStringMaps(templates.DefaultDeployAnnotations(), r.Config.Spec.IstioCoreDNS.PodAnnotations),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []corev1.Container{
						r.coreDNSContainer(),
						r.coreDNSPluginContainer(),
					},
					DNSPolicy: corev1.DNSDefault,
					Volumes: []corev1.Volume{
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "Corefile",
											Path: "Corefile",
										},
									},
									DefaultMode: utils.IntPointer(420),
								},
							},
						},
					},
					Affinity:          r.Config.Spec.IstioCoreDNS.Affinity,
					NodeSelector:      r.Config.Spec.IstioCoreDNS.NodeSelector,
					Tolerations:       r.Config.Spec.IstioCoreDNS.Tolerations,
					PriorityClassName: r.Config.Spec.PriorityClassName,
				},
			},
		},
	}
}
