package gateways

import (
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) podDisruptionBudget() runtime.Object {
	labels := r.labels()
	return &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: templates.ObjectMeta(r.pdbName(), labels, r.gw),
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable: utils.IntstrPointer(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: r.labels(),
			},
		},
	}
}
