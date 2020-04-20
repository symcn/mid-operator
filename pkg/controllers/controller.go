package controllers

import (
	"github.com/symcn/mid-operator/pkg/controllers/sidecar"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, opt *ControllersManagerOption) error {

	if opt.EnableSidecar {
		AddToManagerFuncs = append(AddToManagerFuncs, sidecar.Add)
	}

	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}

	return nil
}
