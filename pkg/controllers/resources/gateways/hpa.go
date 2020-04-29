package gateways

import (
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
)

func (r *Reconciler) horizontalPodAutoscaler() runtime.Object {
	return &autoscalev2beta1.HorizontalPodAutoscaler{
		ObjectMeta: templates.ObjectMeta(r.hpaName(), r.labelSelector(), r.gw),
		Spec: autoscalev2beta1.HorizontalPodAutoscalerSpec{
			MaxReplicas: *r.gw.Spec.MaxReplicas,
			MinReplicas: r.gw.Spec.MinReplicas,
			ScaleTargetRef: autoscalev2beta1.CrossVersionObjectReference{
				Name:       r.gatewayName(),
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			Metrics: templates.TargetAvgCpuUtil80(),
		},
	}
}
