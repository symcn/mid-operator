package gateways

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) service() runtime.Object {
	return &corev1.Service{
		ObjectMeta: templates.ObjectMetaWithAnnotations(r.gatewayName(), utils.MergeStringMaps(r.gw.Spec.ServiceLabels, r.labels()), r.gw.Spec.ServiceAnnotations, r.gw),
		Spec: corev1.ServiceSpec{
			LoadBalancerIP: r.gw.Spec.LoadBalancerIP,
			Type:           r.gw.Spec.ServiceType,
			Ports:          r.servicePorts(r.gw.Name),
			Selector:       r.labelSelector(),
		},
	}
}

func (r *Reconciler) servicePorts(name string) []corev1.ServicePort {
	if name == defaultIngressgatewayName {
		ports := r.gw.Spec.Ports
		if utils.PointerToBool(r.Config.Spec.MeshExpansion) {
			ports = append(ports, corev1.ServicePort{
				Port: 853, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(853), Name: "tcp-dns-tls",
			})

			if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
				ports = append(ports, corev1.ServicePort{
					Port: 15012, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(15012), Name: "tcp-istiod-grpc-tls",
				})
			}
			if utils.PointerToBool(r.Config.Spec.Pilot.Enabled) {
				ports = append(ports, corev1.ServicePort{
					Port: 15011, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(15011), Name: "tcp-pilot-grpc-tls",
				})
			}
		}
		return ports
	}
	return r.gw.Spec.Ports
}
