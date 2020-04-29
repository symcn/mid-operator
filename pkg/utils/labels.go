package utils

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	ComponentNameCrd = "crds"
	CreatedByLabel   = "symcn.io/created-by"
	CreatedBy        = "mid-operator"
)

func GetWatchPredicateForCRDs() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Meta.GetLabels()[CreatedByLabel] == CreatedBy {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Meta.GetLabels()[CreatedByLabel] == CreatedBy {
				return true
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaOld.GetLabels()[CreatedByLabel] == CreatedBy || e.MetaNew.GetLabels()[CreatedByLabel] == CreatedBy {
				return true
			}
			return false
		},
	}
}
