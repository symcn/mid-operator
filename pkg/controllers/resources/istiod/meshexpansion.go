package istiod

import (
	"github.com/symcn/mid-operator/pkg/k8sutils"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r *Reconciler) meshExpansionVirtualService() *k8sutils.DynamicObject {
	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "virtualservices",
		},
		Kind:      "VirtualService",
		Name:      "meshexpansion-vs-istiod",
		Namespace: r.Config.Namespace,
		Labels:    istiodLabels,
		Spec: map[string]interface{}{
			"hosts": []string{
				"istiod." + r.Config.Namespace + ".svc." + r.Config.Spec.Proxy.ClusterDomain,
			},
			"gateways": []string{
				"meshexpansion-gateway",
			},
			"tcp": []map[string]interface{}{
				{
					"match": []map[string]interface{}{
						{
							"port": 15012,
						},
					},
					"route": []map[string]interface{}{
						{
							"destination": map[string]interface{}{
								"host": "istiod." + r.Config.Namespace + ".svc." + r.Config.Spec.Proxy.ClusterDomain,
								"port": map[string]interface{}{
									"number": 15012,
								},
							},
						},
					},
				},
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) meshExpansionDestinationRule() *k8sutils.DynamicObject {
	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      "meshexpansion-dr-istiod",
		Namespace: r.Config.Namespace,
		Labels:    istiodLabels,
		Spec: map[string]interface{}{
			"host": "istiod." + r.Config.Namespace + ".svc." + r.Config.Spec.Proxy.ClusterDomain,
			"trafficPolicy": map[string]interface{}{
				"portLevelSettings": []map[string]interface{}{
					{
						"port": map[string]interface{}{
							"number": 15012,
						},
						"tls": map[string]interface{}{
							"mode": "DISABLE",
						},
					},
				},
			},
		},
		Owner: r.Config,
	}
}
