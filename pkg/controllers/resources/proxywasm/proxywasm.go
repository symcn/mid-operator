package proxywasm

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/controllers/resources"
	"github.com/symcn/mid-operator/pkg/k8sutils"

	"github.com/symcn/mid-operator/pkg/utils"
)

const (
	componentName = "proxywasm"
	wasmRuntime   = "envoy.wasm.runtime.v8"
	noWasmRuntime = "envoy.wasm.runtime.null"
)

type Reconciler struct {
	resources.Reconciler
	dynamic dynamic.Interface
}

func New(client client.Client, dc dynamic.Interface, config *devopsv1beta1.Istio) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		dynamic: dc,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	exchangeFilterDesiredState := k8sutils.DesiredStateAbsent
	statsFilterDesiredState := k8sutils.DesiredStateAbsent
	if utils.PointerToBool(r.Config.Spec.MixerlessTelemetry.Enabled) {
		exchangeFilterDesiredState = k8sutils.DesiredStatePresent
		statsFilterDesiredState = k8sutils.DesiredStatePresent
	}

	if utils.PointerToBool(r.Config.Spec.Proxy.UseMetadataExchangeFilter) {
		exchangeFilterDesiredState = k8sutils.DesiredStatePresent
	}

	drs := []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.metaexchangeEnvoyFilter, DesiredState: exchangeFilterDesiredState},
		{DynamicResource: r.TCPMetaexchangeEnvoyFilter, DesiredState: exchangeFilterDesiredState},
		{DynamicResource: r.httpStatsFsilter, DesiredState: statsFilterDesiredState},
		{DynamicResource: r.tcpStatsFilter, DesiredState: statsFilterDesiredState},
	}

	for _, dr := range drs {
		o := dr.DynamicResource()
		err := o.Reconcile(log, r.dynamic, dr.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}

	log.Info("Reconciled")

	return nil
}
