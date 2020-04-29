package istiod

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
	componentName                = "istiod"
	serviceAccountName           = "istiod-service-account"
	clusterRoleNameIstiod        = "istiod-cluster-role"
	clusterRoleBindingNameIstiod = "istiod-cluster-role-binding"
	configMapNameEnvoy           = "pilot-envoy-config"
	deploymentName               = "istiod"
	ServiceNameIstiod            = "istiod"
	ServiceNamePilot             = "istio-pilot"
	hpaName                      = "istiod-autoscaler"
	pdbName                      = "istiod"
	validatingWebhookName        = "istiod-istio-system"
)

var pilotLabels = map[string]string{
	"app": "istio-pilot",
}

var istiodLabels = map[string]string{
	"app": "istiod",
}

var istiodLabelSelector = map[string]string{
	"istio": "istiod",
}

var pilotLabelSelector = map[string]string{
	"istio": "pilot",
}

var sidecarInjectorLabels = map[string]string{
	"app": "istio-sidecar-injector",
}

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

	var istiodDesiredState k8sutils.DesiredState
	var pdbDesiredState k8sutils.DesiredState
	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		istiodDesiredState = k8sutils.DesiredStatePresent
		if utils.PointerToBool(r.Config.Spec.DefaultPodDisruptionBudget.Enabled) {
			pdbDesiredState = k8sutils.DesiredStatePresent
		} else {
			pdbDesiredState = k8sutils.DesiredStateAbsent
		}
	} else {
		istiodDesiredState = k8sutils.DesiredStateAbsent
		pdbDesiredState = k8sutils.DesiredStateAbsent
	}

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: istiodDesiredState},
		{Resource: r.clusterRole, DesiredState: istiodDesiredState},
		{Resource: r.clusterRoleBinding, DesiredState: istiodDesiredState},
		{Resource: r.configMapEnvoy, DesiredState: istiodDesiredState},
		{Resource: r.deployment, DesiredState: istiodDesiredState},
		{Resource: r.service, DesiredState: istiodDesiredState},
		{Resource: r.horizontalPodAutoscaler, DesiredState: istiodDesiredState},
		{Resource: r.podDisruptionBudget, DesiredState: pdbDesiredState},
		{Resource: r.validatingWebhook, DesiredState: istiodDesiredState},
	} {
		o := res.Resource()
		err := k8sutils.Reconcile(log, r.Client, o, res.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	var meshExpansionDesiredState k8sutils.DesiredState
	var meshExpansionDestinationRuleDesiredState k8sutils.DesiredState
	if utils.PointerToBool(r.Config.Spec.MeshExpansion) {
		meshExpansionDesiredState = k8sutils.DesiredStatePresent
		if r.Config.Spec.ControlPlaneSecurityEnabled {
			meshExpansionDestinationRuleDesiredState = k8sutils.DesiredStatePresent
		}
	} else {
		meshExpansionDesiredState = k8sutils.DesiredStateAbsent
		meshExpansionDestinationRuleDesiredState = k8sutils.DesiredStateAbsent
	}

	for _, dr := range []resources.DynamicResourceWithDesiredState{
		{DynamicResource: r.meshExpansionDestinationRule, DesiredState: meshExpansionDestinationRuleDesiredState},
		{DynamicResource: r.meshExpansionVirtualService, DesiredState: meshExpansionDesiredState},
	} {
		o := dr.DynamicResource()
		err := o.Reconcile(log, r.dynamic, dr.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile dynamic resource", "resource", o.Gvr)
		}
	}

	log.Info("Reconciled")

	return nil
}
