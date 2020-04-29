package cni

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/controllers/resources"
	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

const (
	componentName                = "cni"
	serviceAccountName           = "istio-cni"
	clusterRoleName              = "istio-cni"
	clusterRoleRepairName        = "istio-cni-repair-role"
	clusterRoleBindingName       = "istio-cni"
	clusterRoleBindingRepairName = "istio-cni-repair-rolebinding"
	daemonSetName                = "istio-cni-node"
	configMapName                = "istio-cni-config"
)

var cniLabels = map[string]string{
	"k8s-app": "istio-cni-node",
}

var cniRepairLabels = map[string]string{
	"k8s-app": "istio-cni-repair",
}

var labelSelector = map[string]string{
	"k8s-app": "istio-cni-node",
}

type Reconciler struct {
	resources.Reconciler
}

func New(client client.Client, config *devopsv1beta1.Istio) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	desiredState := k8sutils.DesiredStatePresent
	desiredStateRepair := k8sutils.DesiredStatePresent
	if !utils.PointerToBool(r.Config.Spec.SidecarInjector.InitCNIConfiguration.Enabled) {
		desiredState = k8sutils.DesiredStateAbsent
		desiredStateRepair = k8sutils.DesiredStateAbsent
	}
	if !utils.PointerToBool(r.Config.Spec.SidecarInjector.InitCNIConfiguration.Repair.Enabled) {
		desiredStateRepair = k8sutils.DesiredStateAbsent
	}

	log.Info("Reconciling")

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: desiredState},
		{Resource: r.clusterRole, DesiredState: desiredState},
		{Resource: r.clusterRoleRepair, DesiredState: desiredStateRepair},
		{Resource: r.clusterRoleBinding, DesiredState: desiredState},
		{Resource: r.clusterRoleBindingRepair, DesiredState: desiredStateRepair},
		{Resource: r.configMap, DesiredState: desiredState},
		{Resource: r.daemonSet, DesiredState: desiredState},
	} {
		o := res.Resource()
		err := k8sutils.Reconcile(log, r.Client, o, res.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	log.Info("Reconciled")

	return nil
}
