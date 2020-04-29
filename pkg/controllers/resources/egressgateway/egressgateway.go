package egressgateway

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/controllers/resources"
	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

const (
	componentName = "egressgateway"
	resourceName  = "istio-egressgateway"
)

var (
	resourceLabels = map[string]string{
		"app":   "istio-egressgateway",
		"istio": "egressgateway",
	}
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

	var desiredState k8sutils.DesiredState

	if utils.PointerToBool(r.Config.Spec.Gateways.Enabled) && utils.PointerToBool(r.Config.Spec.Gateways.EgressConfig.Enabled) {
		desiredState = k8sutils.DesiredStatePresent
	} else {
		desiredState = k8sutils.DesiredStateAbsent
	}

	spec := devopsv1beta1.MeshGatewaySpec{
		MeshGatewayConfiguration: r.Config.Spec.Gateways.EgressConfig.MeshGatewayConfiguration,
		Ports:                    r.Config.Spec.Gateways.EgressConfig.Ports,
		Type:                     devopsv1beta1.GatewayTypeEgress,
	}
	spec.Labels = r.labels()
	object := &devopsv1beta1.MeshGateway{
		ObjectMeta: templates.ObjectMeta(resourceName, spec.Labels, r.Config),
		Spec:       spec,
	}

	err := k8sutils.Reconcile(log, r.Client, object, desiredState)
	if err != nil {
		return emperror.WrapWith(err, "failed to reconcile resource", "resource", object.GetObjectKind().GroupVersionKind())
	}

	var multimeshEgressGatewayDesiredState k8sutils.DesiredState
	if utils.PointerToBool(r.Config.Spec.MultiMesh) && utils.PointerToBool(r.Config.Spec.Gateways.EgressConfig.Enabled) {
		multimeshEgressGatewayDesiredState = k8sutils.DesiredStatePresent
	} else {
		multimeshEgressGatewayDesiredState = k8sutils.DesiredStateAbsent
	}

	if r.Config.Name == "istio-config" {
		log.Info("Reconciled")
		return nil
	}

	var drs = []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.multimeshEgressGateway, DesiredState: multimeshEgressGatewayDesiredState},
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

func (r *Reconciler) labels() map[string]string {
	return utils.MergeStringMaps(resourceLabels, r.Config.Spec.Gateways.EgressConfig.MeshGatewayConfiguration.Labels)
}
