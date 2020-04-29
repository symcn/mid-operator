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
	"github.com/symcn/mid-operator/pkg/controllers/resources"
	"github.com/symcn/mid-operator/pkg/controllers/resources/base"
	"github.com/symcn/mid-operator/pkg/controllers/resources/cni"
	"github.com/symcn/mid-operator/pkg/controllers/resources/egressgateway"
	"github.com/symcn/mid-operator/pkg/controllers/resources/ingressgateway"
	"github.com/symcn/mid-operator/pkg/controllers/resources/istiocoredns"
	"github.com/symcn/mid-operator/pkg/controllers/resources/istiod"
	"github.com/symcn/mid-operator/pkg/controllers/resources/proxywasm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"time"

	"github.com/goph/emperror"
	"github.com/pkg/errors"

	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/static"
	"github.com/symcn/mid-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// IstioReconciler reconciles a Istio object
type IstioReconciler struct {
	client.Client
	dynamic        dynamic.Interface
	Log            logr.Logger
	Mgr            manager.Manager
	Scheme         *runtime.Scheme
	CrdsReconciler *k8sutils.CRDReconciler
	recorder       record.EventRecorder
}

func Add(mgr manager.Manager) error {
	dy, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return emperror.Wrap(err, "failed to create dynamic client")
	}

	crds, err := static.LoadIstioCRDs()
	if err != nil {
		return errors.Wrapf(err, "unable to load Istio crd controller")
	}

	for i, crd := range crds {
		klog.Infof("index: %d crd name: %s", i, crd.Name)
	}

	reconciler := &IstioReconciler{
		Client:         mgr.GetClient(),
		dynamic:        dy,
		Mgr:            mgr,
		Log:            ctrl.Log.WithName("controllers").WithName("Istio"),
		Scheme:         mgr.GetScheme(),
		recorder:       mgr.GetEventRecorderFor("istio-controller"),
		CrdsReconciler: k8sutils.NewCRDReconciler(mgr.GetClient(), crds...),
	}

	errc := reconciler.SetupWithManager(mgr)
	if errc != nil {
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
	logger := r.Log.WithValues("key", req.NamespacedName)

	config := &devopsv1beta1.Istio{}
	err := r.Client.Get(ctx, req.NamespacedName, config)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get istio")
		return reconcile.Result{}, err
	}

	// Set default values where not set
	devopsv1beta1.SetDefaults(config)

	if config.Status.Status == "" {
		err := r.updateStatus(config, devopsv1beta1.Created, "", logger)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err = r.CrdsReconciler.Reconcile(logger)
	if err != nil {
		logger.Error(err, "failed to Reconcile istio crd")
	}

	meshNetworks, err := r.getMeshNetworks(config, logger)
	if err != nil {
		return reconcile.Result{}, err
	}
	config.Spec.SetMeshNetworks(meshNetworks)

	reconcilers := []resources.ComponentReconciler{
		base.New(r.Client, config, false),
		istiod.New(r.Client, r.dynamic, config),
		cni.New(r.Client, config),
		istiocoredns.New(r.Client, config),
		proxywasm.New(r.Client, r.dynamic, config),
		ingressgateway.New(r.Client, r.dynamic, config),
		egressgateway.New(r.Client, r.dynamic, config),
	}

	for _, rec := range reconcilers {
		err = rec.Reconcile(logger)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	if utils.PointerToBool(config.Spec.Gateways.Enabled) && utils.PointerToBool(config.Spec.Gateways.IngressConfig.Enabled) {
		config.Status.GatewayAddress, err = GetMeshGatewayAddress(r.Client, client.ObjectKey{
			Name:      ingressgateway.ResourceName,
			Namespace: config.Namespace,
		})
		if err != nil {
			logger.Error(err, "ingress gateway address pending")
			r.updateStatus(config, devopsv1beta1.ReconcileFailed, err.Error(), logger)
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Duration(30) * time.Second,
			}, nil
		}
	}

	err = r.updateStatus(config, devopsv1beta1.Available, "", logger)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}
	logger.Info("reconcile finished")

	// vsList := &networkingv1alpha3.VirtualServiceList{}
	// err = r.Client.List(ctx, vsList)
	// if err != nil {
	// 	if apierrors.IsNotFound(err) {
	// 		return reconcile.Result{}, nil
	// 	}
	//
	// 	logger.Error(err, "failed to get VirtualServiceList")
	// 	return reconcile.Result{}, err
	// }
	//
	// for i := range vsList.Items {
	// 	vs := &vsList.Items[i]
	// 	logger.Info("VirtualService", "ns", vs.Namespace, "name", vs.Name)
	// }
	//
	// envoyFilters := &networkingv1alpha3.EnvoyFilterList{}
	// err = r.Client.List(ctx, envoyFilters)
	// if err != nil {
	// 	if apierrors.IsNotFound(err) {
	// 		return reconcile.Result{}, nil
	// 	}
	//
	// 	logger.Error(err, "failed to get EnvoyFilterList")
	// 	return reconcile.Result{}, err
	// }
	//
	// for i := range envoyFilters.Items {
	// 	envoyFilter := &envoyFilters.Items[i]
	// 	logger.Info("VirtualService", "ns", envoyFilter.Namespace, "name", envoyFilter.Name)
	// }

	return ctrl.Result{}, nil
}

func (r *IstioReconciler) updateStatus(config *devopsv1beta1.Istio, status devopsv1beta1.ConfigState, errorMessage string, logger logr.Logger) error {
	typeMeta := config.TypeMeta
	config.Status.Status = status
	config.Status.ErrorMessage = errorMessage
	err := r.Client.Status().Update(context.Background(), config)
	if apierrors.IsNotFound(err) {
		err = r.Client.Update(context.Background(), config)
	}
	if err != nil {
		if !apierrors.IsConflict(err) {
			return emperror.Wrapf(err, "could not update Istio state to '%s'", status)
		}
		var actualConfig devopsv1beta1.Istio
		err := r.Client.Get(context.TODO(), types.NamespacedName{
			Namespace: config.Namespace,
			Name:      config.Name,
		}, &actualConfig)
		if err != nil {
			return emperror.Wrap(err, "could not get config for updating status")
		}
		actualConfig.Status.Status = status
		actualConfig.Status.ErrorMessage = errorMessage
		err = r.Client.Status().Update(context.Background(), &actualConfig)
		if apierrors.IsNotFound(err) {
			err = r.Client.Update(context.Background(), &actualConfig)
		}
		if err != nil {
			return emperror.Wrapf(err, "could not update Istio state to '%s'", status)
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	config.TypeMeta = typeMeta
	logger.Info("Istio state updated", "status", status)
	return nil
}

func (r *IstioReconciler) getMeshNetworks(config *devopsv1beta1.Istio, logger logr.Logger) (*devopsv1beta1.MeshNetworks, error) {
	meshNetworks := make(map[string]devopsv1beta1.MeshNetwork)

	localNetwork := devopsv1beta1.MeshNetwork{
		Endpoints: []devopsv1beta1.MeshNetworkEndpoint{
			{
				FromRegistry: config.Spec.ClusterName,
			},
		},
	}

	if len(config.Status.GatewayAddress) > 0 {
		gateways := make([]devopsv1beta1.MeshNetworkGateway, 0)
		for _, address := range config.Status.GatewayAddress {
			gateways = append(gateways, devopsv1beta1.MeshNetworkGateway{
				Address: address, Port: 443,
			})
		}
		localNetwork.Gateways = gateways
	}

	meshNetworks[config.Spec.NetworkName] = localNetwork

	// remoteIstios := remoteistioCtrl.GetRemoteIstiosByOwnerReference(r.Mgr, config, logger)
	// for _, remoteIstio := range remoteIstios {
	// 	gateways := make([]devopsv1beta1.MeshNetworkGateway, 0)
	// 	if len(remoteIstio.Status.GatewayAddress) > 0 {
	// 		for _, address := range remoteIstio.Status.GatewayAddress {
	// 			gateways = append(gateways, istiov1beta1.MeshNetworkGateway{
	// 				Address: address, Port: 443,
	// 			})
	// 		}
	// 	} else {
	// 		continue
	// 	}
	//
	// 	meshNetworks[remoteIstio.Name] = devopsv1beta1.MeshNetwork{
	// 		Endpoints: []istiov1beta1.MeshNetworkEndpoint{
	// 			{
	// 				FromRegistry: remoteIstio.Name,
	// 			},
	// 		},
	// 		Gateways: gateways,
	// 	}
	// }

	return &devopsv1beta1.MeshNetworks{
		Networks: meshNetworks,
	}, nil
}
