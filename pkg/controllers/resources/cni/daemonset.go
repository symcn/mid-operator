package cni

import (
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) daemonSet() runtime.Object {
	labels := utils.MergeStringMaps(cniLabels, labelSelector)
	hostPathType := corev1.HostPathUnset
	return &appsv1.DaemonSet{
		ObjectMeta: templates.ObjectMeta(daemonSetName, labels, r.Config),
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: utils.IntstrPointer(1),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: templates.DefaultDeployAnnotations(),
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"beta.kubernetes.io/os": "linux",
					},
					HostNetwork: true,
					Tolerations: []corev1.Toleration{
						{
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoExecute,
						},
						{
							Key:      "CriticalAddonsOnly",
							Operator: corev1.TolerationOpExists,
						},
					},
					TerminationGracePeriodSeconds: utils.Int64Pointer(5),
					ServiceAccountName:            serviceAccountName,
					Containers:                    r.container(),
					Volumes: []corev1.Volume{
						{
							Name: "cni-bin-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: r.Config.Spec.SidecarInjector.InitCNIConfiguration.BinDir,
									Type: &hostPathType,
								},
							},
						},
						{
							Name: "cni-net-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: r.Config.Spec.SidecarInjector.InitCNIConfiguration.ConfDir,
									Type: &hostPathType,
								},
							},
						},
					},
					Affinity:          r.Config.Spec.SidecarInjector.InitCNIConfiguration.Affinity,
					PriorityClassName: r.Config.Spec.PriorityClassName,
				},
			},
		},
	}
}

func (r *Reconciler) container() []corev1.Container {
	cniConfig := r.Config.Spec.SidecarInjector.InitCNIConfiguration
	containers := []corev1.Container{
		{
			Name:            "install-cni",
			Image:           cniConfig.Image,
			ImagePullPolicy: r.Config.Spec.ImagePullPolicy,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "cni-bin-dir",
					MountPath: "/host/opt/cni/bin",
				},
				{
					Name:      "cni-net-dir",
					MountPath: "/host/etc/cni/net.d",
				},
			},
			Command: []string{"/install-cni.sh"},
			Env: []corev1.EnvVar{
				{
					Name: "CNI_NETWORK_CONFIG",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "istio-cni-config",
							},
							Key: "cni_network_config",
						},
					},
				},
				{
					Name:  "CNI_NET_DIR",
					Value: "/etc/cni/net.d",
				},
				{
					Name:  "CHAINED_CNI_PLUGIN",
					Value: strconv.FormatBool(utils.PointerToBool(cniConfig.Chained)),
				},
			},
			TerminationMessagePath:   corev1.TerminationMessagePathDefault,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		},
	}

	if utils.PointerToBool(cniConfig.Repair.Enabled) {
		image := cniConfig.Image
		if !strings.Contains(cniConfig.Image, "/") {
			image = fmt.Sprintf("%s/%s:%s", r.repairHub(), r.repairImage(), r.repairTag())
		}
		containers = append(containers, corev1.Container{
			Name:  "repair-cni",
			Image: image,
			Command: []string{
				"/opt/cni/bin/istio-cni-repair",
			},
			Env: []corev1.EnvVar{
				{
					Name: "REPAIR_NODE-NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "spec.nodeName",
						},
					},
				},
				{
					Name:  "REPAIR_LABEL-PODS",
					Value: strconv.FormatBool(utils.PointerToBool(cniConfig.Repair.LabelPods)),
				},
				{
					Name:  "REPAIR_DELETE-PODS",
					Value: strconv.FormatBool(utils.PointerToBool(cniConfig.Repair.DeletePods)),
				},
				{
					Name:  "REPAIR_RUN-AS-DAEMON",
					Value: "true",
				},
				{
					Name:  "REPAIR_SIDECAR-ANNOTATION",
					Value: "sidecar.istio.io/status",
				},
				{
					Name:  "REPAIR_INIT-CONTAINER-NAME",
					Value: utils.PointerToString(cniConfig.Repair.InitContainerName),
				},
				{
					Name:  "REPAIR_BROKEN-POD-LABEL-KEY",
					Value: utils.PointerToString(cniConfig.Repair.BrokenPodLabelKey),
				},
				{
					Name:  "REPAIR_BROKEN-POD-LABEL-VALUE",
					Value: utils.PointerToString(cniConfig.Repair.BrokenPodLabelValue),
				},
			},
		})
	}

	return containers
}

func (r *Reconciler) repairHub() string {
	repairConfig := r.Config.Spec.SidecarInjector.InitCNIConfiguration.Repair
	if utils.PointerToString(repairConfig.Hub) == "" {
		return "docker.io/istio"
	}

	return utils.PointerToString(repairConfig.Hub)
}

func (r *Reconciler) repairImage() string {
	cniConfig := r.Config.Spec.SidecarInjector.InitCNIConfiguration
	if cniConfig.Image == "" {
		return "install-cni"
	}

	return cniConfig.Image
}

func (r *Reconciler) repairTag() string {
	repairConfig := r.Config.Spec.SidecarInjector.InitCNIConfiguration.Repair
	if utils.PointerToString(repairConfig.Tag) == "" {
		return "1.5.1"
	}

	return utils.PointerToString(repairConfig.Tag)
}
