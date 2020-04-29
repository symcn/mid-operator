package ingressgateway

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
	ResourceName = "istio-ingressgateway"

	componentName         = "ingressgateway"
	k8sIngressGatewayName = "istio-autogenerated-k8s-ingress"
)

var (
	resourceLabels = map[string]string{
		"app":   "istio-ingressgateway",
		"istio": "ingressgateway",
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

	if utils.PointerToBool(r.Config.Spec.Gateways.Enabled) && utils.PointerToBool(r.Config.Spec.Gateways.IngressConfig.Enabled) {
		desiredState = k8sutils.DesiredStatePresent
	} else {
		desiredState = k8sutils.DesiredStateAbsent
	}

	spec := devopsv1beta1.MeshGatewaySpec{
		MeshGatewayConfiguration: r.Config.Spec.Gateways.IngressConfig.MeshGatewayConfiguration,
		Ports:                    r.Config.Spec.Gateways.IngressConfig.Ports,
		Type:                     devopsv1beta1.GatewayTypeIngress,
	}
	spec.Labels = r.labels()
	object := &devopsv1beta1.MeshGateway{
		ObjectMeta: templates.ObjectMeta(ResourceName, spec.Labels, r.Config),
		Spec:       spec,
	}

	err := k8sutils.Reconcile(log, r.Client, object, desiredState)
	if err != nil {
		return emperror.WrapWith(err, "failed to reconcile resource", "resource", object.GetObjectKind().GroupVersionKind())
	}

	var k8sIngressDesiredState k8sutils.DesiredState
	if utils.PointerToBool(r.Config.Spec.Gateways.Enabled) &&
		utils.PointerToBool(r.Config.Spec.Gateways.IngressConfig.Enabled) &&
		utils.PointerToBool(r.Config.Spec.Gateways.K8sIngress.Enabled) {
		k8sIngressDesiredState = k8sutils.DesiredStatePresent
	} else {
		k8sIngressDesiredState = k8sutils.DesiredStateAbsent
	}

	var meshExpansionDesiredState k8sutils.DesiredState
	if utils.PointerToBool(r.Config.Spec.MeshExpansion) {
		meshExpansionDesiredState = k8sutils.DesiredStatePresent
	} else {
		meshExpansionDesiredState = k8sutils.DesiredStateAbsent
	}

	var multimeshDesiredState k8sutils.DesiredState
	if utils.PointerToBool(r.Config.Spec.MultiMesh) {
		multimeshDesiredState = k8sutils.DesiredStatePresent
	} else {
		multimeshDesiredState = k8sutils.DesiredStateAbsent
	}

	if r.Config.Name == "istio-config" {
		log.Info("Reconciled")
		return nil
	}

	var drs = []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.k8sIngressGateway, DesiredState: k8sIngressDesiredState},
		{DynamicResource: r.meshExpansionGateway, DesiredState: meshExpansionDesiredState},
		{DynamicResource: r.clusterAwareGateway, DesiredState: meshExpansionDesiredState},
		{DynamicResource: r.multimeshIngressGateway, DesiredState: multimeshDesiredState},
		{DynamicResource: r.multimeshDestinationRule, DesiredState: multimeshDesiredState},
		{DynamicResource: r.multimeshEnvoyFilter, DesiredState: multimeshDesiredState},
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
	return utils.MergeStringMaps(resourceLabels, r.Config.Spec.Gateways.IngressConfig.MeshGatewayConfiguration.Labels)
}
