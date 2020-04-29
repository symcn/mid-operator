package egressgateway

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

const (
	multimeshResourceNamePrefix = "istio-multicluster"
)

func (r *Reconciler) multimeshEgressGateway() *k8sutils.DynamicObject {
	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		},
		Kind:      "Gateway",
		Name:      multimeshResourceNamePrefix + "-egressgateway",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"hosts": utils.EmptyTypedStrSlice("*.global"),
					"port": map[string]interface{}{
						"name":     "tls",
						"protocol": "TLS",
						"number":   15443,
					},
					"tls": map[string]interface{}{
						"mode": "AUTO_PASSTHROUGH",
					},
				},
			},
			"selector": r.labels(),
		},
		Owner: r.Config,
	}
}
