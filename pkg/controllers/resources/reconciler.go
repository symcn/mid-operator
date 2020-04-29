package resources

import (
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/k8sutils"
	"github.com/symcn/mid-operator/pkg/utils"
)

type Reconciler struct {
	client.Client
	Config *devopsv1beta1.Istio
}

type ComponentReconciler interface {
	Reconcile(log logr.Logger) error
}

type Resource func() runtime.Object

type ResourceVariation func(t string) runtime.Object

func ResolveVariations(t string, v []ResourceVariationWithDesiredState, desiredState k8sutils.DesiredState) []ResourceWithDesiredState {
	var state k8sutils.DesiredState
	resources := make([]ResourceWithDesiredState, 0)
	for i := range v {
		i := i

		if v[i].DesiredState == k8sutils.DesiredStateAbsent || desiredState == k8sutils.DesiredStateAbsent {
			state = k8sutils.DesiredStateAbsent
		} else {
			state = k8sutils.DesiredStatePresent
		}

		resource := ResourceWithDesiredState{
			func() runtime.Object {
				return v[i].ResourceVariation(t)
			},
			state,
		}
		resources = append(resources, resource)
	}

	return resources
}

type DynamicResource func() *k8sutils.DynamicObject

type DynamicResourceWithDesiredState struct {
	DynamicResource DynamicResource
	DesiredState    k8sutils.DesiredState
}

type ResourceWithDesiredState struct {
	Resource     Resource
	DesiredState k8sutils.DesiredState
}

type ResourceVariationWithDesiredState struct {
	ResourceVariation ResourceVariation
	DesiredState      k8sutils.DesiredState
}

func GetDiscoveryPort(in *devopsv1beta1.Istio) string {
	if utils.PointerToBool(in.Spec.Istiod.Enabled) {
		return "15012"
	}
	if in.Spec.ControlPlaneSecurityEnabled {
		return "15011"
	}
	return "15010"
}

func GetDiscoveryAddress(in *devopsv1beta1.Istio, svcNames ...string) string {
	svcName := "istio-pilot"
	if len(svcNames) == 1 {
		svcName = svcNames[0]
	}
	return fmt.Sprintf("%s.%s.svc:%s", svcName, in.Namespace, GetDiscoveryPort(in))
}
