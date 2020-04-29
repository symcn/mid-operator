package ingressgateway

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) meshExpansionGateway() *k8sutils.DynamicObject {
	servers := make([]map[string]interface{}, 0)

	if utils.PointerToBool(r.Config.Spec.Pilot.Enabled) {
		servers = append(servers, map[string]interface{}{
			"port": map[string]interface{}{
				"name":     "tcp-pilot",
				"protocol": "TCP",
				"number":   15011,
			},
			"hosts": utils.EmptyTypedStrSlice("*"),
		})
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		servers = append(servers, map[string]interface{}{
			"port": map[string]interface{}{
				"name":     "tcp-istiod",
				"protocol": "TCP",
				"number":   15012,
			},
			"hosts": utils.EmptyTypedStrSlice("*"),
		})
	}

	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		},
		Kind:      "Gateway",
		Name:      "meshexpansion-gateway",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"servers":  servers,
			"selector": r.labels(),
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) clusterAwareGateway() *k8sutils.DynamicObject {
	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		},
		Kind:      "Gateway",
		Name:      "cluster-aware-gateway",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"port": map[string]interface{}{
						"name":     "tls",
						"protocol": "TLS",
						"number":   443,
					},
					"tls": map[string]interface{}{
						"mode": "AUTO_PASSTHROUGH",
					},
					"hosts": utils.EmptyTypedStrSlice("*.local"),
				},
			},
			"selector": r.labels(),
		},
		Owner: r.Config,
	}
}
