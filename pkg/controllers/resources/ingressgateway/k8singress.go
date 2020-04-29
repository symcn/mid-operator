package ingressgateway

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) k8sIngressGateway() *k8sutils.DynamicObject {
	spec := map[string]interface{}{
		"servers": []map[string]interface{}{
			{
				"port": map[string]interface{}{
					"name":     "http",
					"protocol": "HTTP2",
					"number":   80,
				},
				"hosts": utils.EmptyTypedStrSlice("*"),
			},
		},
		"selector": r.labels(),
	}

	if utils.PointerToBool(r.Config.Spec.Gateways.K8sIngress.EnableHttps) {
		spec["servers"] = append([]interface{}{spec["servers"]}, map[string]interface{}{
			"port": map[string]interface{}{
				"name":     "https-default",
				"protocol": "HTTPS",
				"number":   443,
			},
			"hosts": utils.EmptyTypedStrSlice("*"),
			"tls": map[string]interface{}{
				"mode":              "SIMPLE",
				"serverCertificate": "/etc/istio/ingressgateway-certs/tls.crt",
				"privateKey":        "/etc/istio/ingressgateway-certs/tls.key",
			},
		})
	}

	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		},
		Kind:      "Gateway",
		Name:      k8sIngressGatewayName,
		Namespace: r.Config.Namespace,
		Spec:      spec,
		Owner:     r.Config,
	}
}
