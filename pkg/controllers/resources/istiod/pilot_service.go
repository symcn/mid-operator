package istiod

import (
	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *Reconciler) pilotServicePorts() []apiv1.ServicePort {
	ports := []apiv1.ServicePort{
		{
			Name:       "grpc-xds",
			Port:       15010,
			TargetPort: intstr.FromInt(15010),
			Protocol:   apiv1.ProtocolTCP,
		},
		{
			Name:       "https-xds",
			Port:       15011,
			TargetPort: intstr.FromInt(15011),
			Protocol:   apiv1.ProtocolTCP,
		},
		{
			Name:       "https-dns",
			Port:       15012,
			TargetPort: intstr.FromInt(15012),
			Protocol:   apiv1.ProtocolTCP,
		},
		{
			Name:       "http-legacy-discovery",
			Port:       8080,
			TargetPort: intstr.FromInt(8080),
			Protocol:   apiv1.ProtocolTCP,
		},
		{
			Name:       "http-monitoring",
			Port:       15014,
			TargetPort: intstr.FromInt(15014),
			Protocol:   apiv1.ProtocolTCP,
		},
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		ports = append(ports, apiv1.ServicePort{
			Name:       "https-webhook",
			Port:       443,
			TargetPort: intstr.FromInt(15017),
			Protocol:   apiv1.ProtocolTCP,
		})
	}

	return ports
}

func (r *Reconciler) pilotServicervice() runtime.Object {
	return &apiv1.Service{
		ObjectMeta: templates.ObjectMeta(ServiceNamePilot, utils.MergeStringMaps(pilotLabels, pilotLabelSelector), r.Config),
		Spec: apiv1.ServiceSpec{
			Ports:    r.pilotServicePorts(),
			Selector: pilotLabelSelector,
		},
	}
}
