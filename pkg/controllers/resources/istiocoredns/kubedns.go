package istiocoredns

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/symcn/mid-operator/pkg/k8sutils"
)

// Add/Remove .global to/from kube-dns configmap
func (r *Reconciler) reconcileKubeDNSConfigMap(log logr.Logger, desiredState k8sutils.DesiredState) error {
	var cm corev1.ConfigMap

	err := r.Client.Get(context.Background(), types.NamespacedName{
		Name:      "kube-dns",
		Namespace: "kube-system",
	}, &cm)
	if k8serrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return emperror.Wrap(err, "could not get kube-dns configmap")
	}

	stubDomains := make(map[string][]string, 0)
	if cm.Data["stubDomains"] != "" {
		err = json.Unmarshal([]byte(cm.Data["stubDomains"]), &stubDomains)
		if err != nil {
			return emperror.Wrap(err, "could not unmarshal stubDomains")
		}
	}

	if desiredState == k8sutils.DesiredStatePresent {
		var svc corev1.Service
		err = r.Client.Get(context.Background(), types.NamespacedName{
			Name:      serviceName,
			Namespace: r.Config.Namespace,
		}, &svc)
		if err != nil {
			return emperror.Wrap(err, "could not get Istio coreDNS service")
		}
		stubDomains["global"] = []string{svc.Spec.ClusterIP}
	} else if desiredState == k8sutils.DesiredStateAbsent {
		_, ok := stubDomains["global"]
		if ok {
			delete(stubDomains, "global")
		}
	}

	stubDomainsData, err := json.Marshal(&stubDomains)
	if err != nil {
		return emperror.Wrap(err, "could not marshal updated stub domains")
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string, 0)
	}
	cm.Data["stubDomains"] = string(stubDomainsData)

	err = r.Client.Update(context.Background(), &cm)
	if err != nil {
		return emperror.Wrap(err, "could not update kube-dns configmap")
	}

	return nil
}
