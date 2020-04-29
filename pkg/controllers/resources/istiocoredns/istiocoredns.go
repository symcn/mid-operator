package istiocoredns

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
	componentName          = "istiocoredns"
	deploymentName         = "istiocoredns"
	configMapName          = "istiocoredns"
	serviceAccountName     = "istiocoredns-service-account"
	clusterRoleName        = "istiocoredns"
	clusterRoleBindingName = "istio-istiocoredns-cluster-role-binding"
	serviceName            = "istiocoredns"
)

var labels = map[string]string{
	"app":   "istio-istiocoredns",
	"istio": "istiocoredns",
}

var labelSelector = map[string]string{
	"app": "istio-istiocoredns",
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

	log.Info("Reconciling")

	var desiredState k8sutils.DesiredState
	if utils.PointerToBool(r.Config.Spec.IstioCoreDNS.Enabled) {
		desiredState = k8sutils.DesiredStatePresent
	} else {
		desiredState = k8sutils.DesiredStateAbsent
	}

	for _, res := range []resources.Resource{
		r.serviceAccount,
		r.clusterRole,
		r.clusterRoleBinding,
		r.configMap,
		r.service,
		r.deployment,
	} {
		o := res()
		err := k8sutils.Reconcile(log, r.Client, o, desiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	err := r.reconcileCoreDNSConfigMap(log, desiredState)
	if err != nil {
		return emperror.WrapWith(err, "failed to update coredns configmap")
	}

	err = r.reconcileKubeDNSConfigMap(log, desiredState)
	if err != nil {
		return emperror.WrapWith(err, "failed to update kube-dns configmap")
	}

	log.Info("Reconciled")

	return nil
}
