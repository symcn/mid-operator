package k8sutils

import (
	"bytes"
	"net"
	"sort"

	corev1 "k8s.io/api/core/v1"
)

type IngressSetupPendingError struct{}

func (e IngressSetupPendingError) Error() string {
	return "ingress gateway endpoint address is pending"
}

func GetServiceEndpointIPs(service corev1.Service) ([]string, error) {
	ips := make([]string, 0)

	switch service.Spec.Type {
	case corev1.ServiceTypeClusterIP:
		if service.Spec.ClusterIP != corev1.ClusterIPNone {
			ips = []string{
				service.Spec.ClusterIP,
			}
		}
	case corev1.ServiceTypeLoadBalancer:
		if len(service.Status.LoadBalancer.Ingress) < 1 {
			return ips, IngressSetupPendingError{}
		}

		if service.Status.LoadBalancer.Ingress[0].IP != "" {
			ips = []string{
				service.Status.LoadBalancer.Ingress[0].IP,
			}
		} else if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
			hostIPs, err := net.LookupIP(service.Status.LoadBalancer.Ingress[0].Hostname)
			if err != nil {
				return ips, err
			}
			sort.Slice(hostIPs, func(i, j int) bool {
				return bytes.Compare(hostIPs[i], hostIPs[j]) < 0
			})
			for _, ip := range hostIPs {
				if ip.To4() != nil {
					ips = append(ips, ip.String())
				}
			}
		}
	}

	return ips, nil
}
