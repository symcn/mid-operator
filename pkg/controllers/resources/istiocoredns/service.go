package istiocoredns

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) service() runtime.Object {
	return &corev1.Service{
		ObjectMeta: templates.ObjectMeta(serviceName, utils.MergeStringMaps(labels, labelSelector), r.Config),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "dns",
					Port:       53,
					Protocol:   corev1.ProtocolUDP,
					TargetPort: intstr.FromInt(53),
				},
				{
					Name:       "dns-tcp",
					Port:       53,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(53),
				},
			},
			Selector: labelSelector,
		},
	}
}
