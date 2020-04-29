package istiod

import (
	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) horizontalPodAutoscaler() runtime.Object {
	return &autoscalev2beta1.HorizontalPodAutoscaler{
		ObjectMeta: templates.ObjectMeta(hpaName, istiodLabels, r.Config),
		Spec: autoscalev2beta1.HorizontalPodAutoscalerSpec{
			MaxReplicas: utils.PointerToInt32(r.Config.Spec.Pilot.MaxReplicas),
			MinReplicas: r.Config.Spec.Pilot.MinReplicas,
			ScaleTargetRef: autoscalev2beta1.CrossVersionObjectReference{
				Name:       deploymentName,
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			Metrics: templates.TargetAvgCpuUtil80(),
		},
	}
}
