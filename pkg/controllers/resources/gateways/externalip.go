package gateways

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/symcn/mid-operator/pkg/k8sutils"
)

func (r *Reconciler) GetGatewayAddress() ([]string, error) {
	var service corev1.Service
	var ips []string

	err := r.Get(context.Background(), types.NamespacedName{
		Name:      r.gatewayName(),
		Namespace: r.gw.Namespace,
	}, &service)
	if err != nil {
		return nil, err
	}

	ips, err = k8sutils.GetServiceEndpointIPs(service)
	if err != nil {
		return nil, err
	}

	return ips, nil
}
