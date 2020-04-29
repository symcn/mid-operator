package istiod

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) service() runtime.Object {
	return &corev1.Service{
		ObjectMeta: templates.ObjectMeta(ServiceNameIstiod, istiodLabels, r.Config),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "https-dns",
					Port:       15012,
					TargetPort: intstr.FromInt(15012),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "https-webhook",
					Port:       443,
					TargetPort: intstr.FromInt(15017),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: utils.MergeStringMaps(istiodLabels, pilotLabelSelector),
		},
	}
}
