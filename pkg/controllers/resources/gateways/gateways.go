package gateways

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/controllers/resources"
	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	componentName             = "meshgateway"
	defaultIngressgatewayName = "istio-ingressgateway"
)

type Reconciler struct {
	resources.Reconciler
	gw      *devopsv1beta1.MeshGateway
	dynamic dynamic.Interface
}

func New(client client.Client, dc dynamic.Interface, config *devopsv1beta1.Istio, gw *devopsv1beta1.MeshGateway) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		gw:      gw,
		dynamic: dc,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	err := r.waitForIstiod()
	if err != nil {
		return err
	}

	pdbDesiredState := k8sutils.DesiredStateAbsent
	if utils.PointerToBool(r.Config.Spec.DefaultPodDisruptionBudget.Enabled) {
		pdbDesiredState = k8sutils.DesiredStatePresent
	}

	sdsDesiredState := k8sutils.DesiredStateAbsent
	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		sdsDesiredState = k8sutils.DesiredStatePresent
	}

	hpaDesiredState := k8sutils.DesiredStateAbsent
	if r.gw.Spec.MinReplicas != nil && r.gw.Spec.MaxReplicas != nil && *r.gw.Spec.MinReplicas > 1 && *r.gw.Spec.MinReplicas != *r.gw.Spec.MaxReplicas {
		hpaDesiredState = k8sutils.DesiredStatePresent
	}

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: k8sutils.DesiredStatePresent},
		{Resource: r.clusterRole, DesiredState: k8sutils.DesiredStatePresent},
		{Resource: r.clusterRoleBinding, DesiredState: k8sutils.DesiredStatePresent},
		{Resource: r.deployment, DesiredState: k8sutils.DesiredStatePresent},
		{Resource: r.service, DesiredState: k8sutils.DesiredStatePresent},
		{Resource: r.horizontalPodAutoscaler, DesiredState: hpaDesiredState},
		{Resource: r.podDisruptionBudget, DesiredState: pdbDesiredState},
		{Resource: r.role, DesiredState: sdsDesiredState},
		{Resource: r.roleBinding, DesiredState: sdsDesiredState},
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

func (r *Reconciler) waitForIstiod() error {
	if !utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		return nil
	}

	opt := &client.ListOptions{
		Namespace: r.Config.Namespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "istiod",
		}),
	}

	var pods v1.PodList
	err := r.Client.List(context.Background(), &pods, opt)
	if err != nil {
		return emperror.Wrap(err, "could not list pods")
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			readyContainers := 0
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Ready {
					readyContainers++
				}
			}
			if readyContainers == len(pod.Status.ContainerStatuses) {
				return nil
			}
		}
	}

	return errors.Errorf("Istiod is not running yet")
}
