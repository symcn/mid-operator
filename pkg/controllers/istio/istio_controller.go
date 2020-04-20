/*
Copyright 2020 The symcn authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package istio

import (
	"context"

	"github.com/go-logr/logr"
	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// IstioReconciler reconciles a Istio object
type IstioReconciler struct {
	client.Client
	Log    logr.Logger
	Mgr    manager.Manager
	Scheme *runtime.Scheme
}

func Add(mgr manager.Manager) error {
	reconciler := &IstioReconciler{
		Client: mgr.GetClient(),
		Mgr:    mgr,
		Log:    ctrl.Log.WithName("controllers").WithName("Istio"),
		Scheme: mgr.GetScheme(),
	}

	err := reconciler.SetupWithManager(mgr)
	if err != nil {
		return errors.Wrapf(err, "unable to create Istio controller")
	}

	return nil
}

func (r *IstioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1beta1.Istio{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=devops.symcn.com,resources=istios,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.symcn.com,resources=istios/status,verbs=get;update;patch

func (r *IstioReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("istio", req.NamespacedName)

	ctx := context.Background()
	logger := r.Log.WithValues("key", req.NamespacedName, "id", uuid.Must(uuid.NewV4()).String())

	istio := &devopsv1beta1.Istio{}
	err := r.Client.Get(ctx, req.NamespacedName, istio)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get istio")
		return reconcile.Result{}, err
	}

	// test
	vsList := &networkingv1beta1.VirtualServiceList{}
	err = r.Client.List(ctx, vsList)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get VirtualServiceList")
		return reconcile.Result{}, err
	}

	for i := range vsList.Items {
		vs := &vsList.Items[i]
		logger.Info("VirtualService", "ns", vs.Namespace, "name", vs.Name)
	}

	return ctrl.Result{}, nil
}
