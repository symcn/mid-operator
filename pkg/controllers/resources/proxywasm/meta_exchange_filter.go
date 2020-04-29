package proxywasm

import (
	"fmt"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

const (
	metadataExchangeFilterYAML = `
- applyTo: HTTP_FILTER
  match:
    context: ANY # inbound, outbound, and gateway
    proxy:
      proxyVersion: '^1\.5.*'
    listener:
      filterChain:
        filter:
          name: "envoy.http_connection_manager"
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.http.wasm
      typed_config:
        "@type": type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.http.wasm.v2.Wasm
        value:
          config:
            configuration: envoy.wasm.metadata_exchange
            vm_config:
              code:
                local:
                  %s
              runtime: %s
`
	TCPMetadataExchangeFilterYAML = `
    - applyTo: NETWORK_FILTER
      match:
        context: SIDECAR_INBOUND
        proxy:
          proxyVersion: '^1\.5.*'
        listener: {}
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.network.metadata_exchange
          config:
            protocol: istio-peer-exchange
    - applyTo: CLUSTER
      match:
        context: SIDECAR_OUTBOUND
        proxy:
          proxyVersion: '^1\.5.*'
        cluster: {}
      patch:
        operation: MERGE
        value:
          filters:
          - name: envoy.filters.network.upstream.metadata_exchange
            typed_config:
              "@type": type.googleapis.com/udpa.type.v1.TypedStruct
              type_url: type.googleapis.com/envoy.tcp.metadataexchange.config.MetadataExchange
              value:
                protocol: istio-peer-exchange
    - applyTo: CLUSTER
      match:
        context: GATEWAY
        proxy:
          proxyVersion: '^1\.5.*'
        cluster: {}
      patch:
        operation: MERGE
        value:
          filters:
          - name: envoy.filters.network.upstream.metadata_exchange
            typed_config:
              "@type": type.googleapis.com/udpa.type.v1.TypedStruct
              type_url: type.googleapis.com/envoy.tcp.metadataexchange.config.MetadataExchange
              value:
                protocol: istio-peer-exchange
`
	metaExchangeWasmLocal   = "filename: /etc/istio/extensions/metadata-exchange-filter.wasm"
	metaExchangeNoWasmLocal = "inline_string: envoy.wasm.metadata_exchange"
)

func (r *Reconciler) metaexchangeEnvoyFilter() *k8sutils.DynamicObject {

	wasmEnabled := utils.PointerToBool(r.Config.Spec.ProxyWasm.Enabled)

	vmConfigLocal := metaExchangeNoWasmLocal
	vmConfigRuntime := noWasmRuntime
	if wasmEnabled {
		vmConfigLocal = metaExchangeWasmLocal
		vmConfigRuntime = wasmRuntime
	}

	var y []map[string]interface{}
	yaml.Unmarshal([]byte(fmt.Sprintf(metadataExchangeFilterYAML, vmConfigLocal, vmConfigRuntime)), &y)

	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      componentName + "-metadata-exchange",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"configPatches": y,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) TCPMetaexchangeEnvoyFilter() *k8sutils.DynamicObject {
	var y []map[string]interface{}
	yaml.Unmarshal([]byte(TCPMetadataExchangeFilterYAML), &y)

	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      componentName + "-tcp-metadata-exchange",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"configPatches": y,
		},
		Owner: r.Config,
	}
}
