package base

import (
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/controllers/resources"
	"github.com/symcn/mid-operator/pkg/k8sutils"
)

const (
	componentName                 = "common"
	istioReaderName               = "istio-reader"
	istioReaderServiceAccountName = istioReaderName + "-service-account"
	IstioConfigMapName            = "istio"
)

var istioReaderLabel = map[string]string{
	"app": istioReaderName,
}

type Reconciler struct {
	resources.Reconciler
	remote bool
}

func New(client client.Client, config *devopsv1beta1.Istio, isRemote bool) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		remote: isRemote,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	for _, res := range []resources.Resource{
		r.serviceAccount,
		r.clusterRole,
		r.clusterRoleBinding,
		r.configMap,
	} {
		o := res()
		err := k8sutils.Reconcile(log, r.Client, o, k8sutils.DesiredStatePresent)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	log.Info("Reconciled")

	return nil
}
