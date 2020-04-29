package proxywasm

import (
	"fmt"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

const (
	httpStatsFilterYAML = `
- applyTo: HTTP_FILTER
  match:
    context: SIDECAR_OUTBOUND
    proxy:
      proxyVersion: '^1\.5.*'
    listener:
      filterChain:
        filter:
          name: envoy.http_connection_manager
          subFilter:
            name: envoy.router
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.http.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.http.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio"
              }
            root_id: stats_outbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: stats_outbound
- applyTo: HTTP_FILTER
  match:
    context: SIDECAR_INBOUND
    proxy:
      proxyVersion: '^1\.5.*'
    listener:
      filterChain:
        filter:
          name: envoy.http_connection_manager
          subFilter:
            name: envoy.router
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.http.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.http.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio"
              }
            root_id: stats_inbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: stats_inbound
- applyTo: HTTP_FILTER
  match:
    context: GATEWAY
    proxy:
      proxyVersion: '^1\.5.*'
    listener:
      filterChain:
        filter:
          name: envoy.http_connection_manager
          subFilter:
            name: envoy.router
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.http.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.http.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio"
              }
            root_id: stats_outbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: stats_outbound
`
	tcpStatsFilterYAML = `
- applyTo: NETWORK_FILTER
  match:
    context: SIDECAR_INBOUND
    listener:
      filterChain:
        filter:
          name: envoy.tcp_proxy
    proxy:
      proxyVersion: 1\.5.*
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.network.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.network.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio",
              }
            root_id: stats_inbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: stats_inbound
- applyTo: NETWORK_FILTER
  match:
    context: SIDECAR_OUTBOUND
    listener:
      filterChain:
        filter:
          name: envoy.tcp_proxy
    proxy:
      proxyVersion: 1\.5.*
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.network.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.network.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio",
              }
            root_id: stats_outbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: stats_outbound
- applyTo: NETWORK_FILTER
  match:
    context: GATEWAY
    listener:
      filterChain:
        filter:
          name: envoy.tcp_proxy
    proxy:
      proxyVersion: 1\.5.*
  patch:
    operation: INSERT_BEFORE
    value:
      name: envoy.filters.network.wasm
      typed_config:
        '@type': type.googleapis.com/udpa.type.v1.TypedStruct
        type_url: type.googleapis.com/envoy.config.filter.network.wasm.v2.Wasm
        value:
          config:
            configuration: |
              {
                "debug": "false",
                "stat_prefix": "istio",
              }
            root_id: stats_outbound
            vm_config:
              code:
                local:
                  %[1]s
              runtime: %[2]s
              vm_id: stats_outbound
`
	statsWasmLocal   = "filename: /etc/istio/extensions/stats-filter.wasm"
	statsNoWasmLocal = "inline_string: envoy.wasm.stats"
)

func (r *Reconciler) httpStatsFsilter() *k8sutils.DynamicObject {

	wasmEnabled := utils.PointerToBool(r.Config.Spec.ProxyWasm.Enabled)

	vmConfigLocal := statsNoWasmLocal
	vmConfigRuntime := noWasmRuntime
	if wasmEnabled {
		vmConfigLocal = statsWasmLocal
		vmConfigRuntime = wasmRuntime
	}

	var y []map[string]interface{}
	yaml.Unmarshal([]byte(fmt.Sprintf(httpStatsFilterYAML, vmConfigLocal, vmConfigRuntime)), &y)

	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      componentName + "-stats",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"configPatches": y,
		},
		Owner: r.Config,
	}
}

func (r *Reconciler) tcpStatsFilter() *k8sutils.DynamicObject {

	wasmEnabled := utils.PointerToBool(r.Config.Spec.ProxyWasm.Enabled)

	vmConfigLocal := statsNoWasmLocal
	vmConfigRuntime := noWasmRuntime
	if wasmEnabled {
		vmConfigLocal = statsWasmLocal
		vmConfigRuntime = wasmRuntime
	}

	var y []map[string]interface{}
	yaml.Unmarshal([]byte(fmt.Sprintf(tcpStatsFilterYAML, vmConfigLocal, vmConfigRuntime)), &y)

	return &k8sutils.DynamicObject{
		Gvr: schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "envoyfilters",
		},
		Kind:      "EnvoyFilter",
		Name:      componentName + "-tcp-stats",
		Namespace: r.Config.Namespace,
		Spec: map[string]interface{}{
			"configPatches": y,
		},
		Owner: r.Config,
	}
}
