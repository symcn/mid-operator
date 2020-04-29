package ingressgateway

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

const (
	multimeshResourceNamePrefix = "istio-multicluster"
)

func (r *Reconciler) multimeshIngressGateway() *k8sutils.DynamicObject {
	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		},
		Kind:      "Gateway",
		Name:      multimeshResourceNamePrefix + "-ingressgateway",
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
			"selector": resourceLabels,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) multimeshEnvoyFilter() *k8sutils.DynamicObject {
	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      multimeshResourceNamePrefix + "-ingressgateway",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"workloadSelector": map[string]interface{}{
				"labels": resourceLabels,
			},
			"configPatches": []map[string]interface{}{
				{
					"applyTo": "NETWORK_FILTER",
					"match": map[string]interface{}{
						"context": "GATEWAY",
						"listener": map[string]interface{}{
							"portNumber": 15443,
							"filterChain": map[string]interface{}{
								"filter": map[string]interface{}{
									"name": "envoy.filters.network.sni_cluster",
								},
							},
						},
					},
					"patch": map[string]interface{}{
						"operation": "INSERT_AFTER",
						"value": map[string]interface{}{
							"name": "envoy.filters.network.tcp_cluster_rewrite",
							"config": map[string]interface{}{
								"cluster_pattern":     "\\.global$",
								"cluster_replacement": ".svc." + r.Config.Spec.Proxy.ClusterDomain,
							},
						},
					},
				},
			},
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) multimeshDestinationRule() *k8sutils.DynamicObject {
	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules",
		},
		Kind:      "DestinationRule",
		Name:      multimeshResourceNamePrefix + "-destinationrule",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"host": "*.global",
			"trafficPolicy": map[string]interface{}{
				"tls": map[string]interface{}{
					"mode": "ISTIO_MUTUAL",
				},
			},
		},
		Owner: r.Config,
	}
}
