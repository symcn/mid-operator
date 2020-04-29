package templates

import (
	appsv1 "k8s.io/api/apps/v1"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/utils"
)

func DefaultDeployAnnotations() map[string]string {
	return map[string]string{
		"sidecar.istio.io/inject":                    "false",
		"scheduler.alpha.kubernetes.io/critical-pod": "",
	}
}

func GetResourcesRequirementsOrDefault(requirements *corev1.ResourceRequirements, defaults *corev1.ResourceRequirements) corev1.ResourceRequirements {
	if requirements != nil {
		return *requirements
	}

	return *defaults
}

func DefaultRollingUpdateStrategy() appsv1.DeploymentStrategy {
	return appsv1.DeploymentStrategy{
		RollingUpdate: &appsv1.RollingUpdateDeployment{
			MaxSurge:       utils.IntstrPointer(1),
			MaxUnavailable: utils.IntstrPointer(0),
		},
	}
}

func TargetAvgCpuUtil80() []autoscalev2beta1.MetricSpec {
	return []autoscalev2beta1.MetricSpec{
		{
			Type: autoscalev2beta1.ResourceMetricSourceType,
			Resource: &autoscalev2beta1.ResourceMetricSource{
				Name:                     corev1.ResourceCPU,
				TargetAverageUtilization: utils.IntPointer(80),
			},
		},
	}
}

func IstioProxyEnv(config *devopsv1beta1.Istio) []corev1.EnvVar {
	envs := []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: "INSTANCE_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
	}

	// envs = append(envs, corev1.EnvVar{
	// 	Name:  "SDS_ENABLED",
	// 	Value: strconv.FormatBool(utils.PointerToBool(config.Spec.SDS.Enabled)),
	// })

	if utils.PointerToBool(config.Spec.Istiod.Enabled) {
		envs = append(envs, []corev1.EnvVar{
			{
				Name:  "JWT_POLICY",
				Value: string(config.Spec.JWTPolicy),
			},
			{
				Name:  "PILOT_CERT_PROVIDER",
				Value: string(config.Spec.Pilot.CertProvider),
			},
		}...)
	}

	return envs
}
