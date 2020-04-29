package k8sutils

import (
	"context"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/symcn/mid-operator/pkg/utils"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CRDReconciler struct {
	crds       []*extensionsobj.CustomResourceDefinition
	runtimeCli client.Client
}

func NewCRDReconciler(cli client.Client, crds ...*extensionsobj.CustomResourceDefinition) *CRDReconciler {
	return &CRDReconciler{
		crds:       crds,
		runtimeCli: cli,
	}
}

func (r *CRDReconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", utils.ComponentNameCrd)
	for _, obj := range r.crds {
		crd := obj.DeepCopy()
		log := log.WithValues("kind", crd.Spec.Names.Kind)
		current := &extensionsobj.CustomResourceDefinition{}
		err := r.runtimeCli.Get(context.TODO(), client.ObjectKey{Name: crd.Name}, current)
		if err != nil && !apierrors.IsNotFound(err) {
			return emperror.WrapWith(err, "getting CRD failed", "kind", crd.Spec.Names.Kind)
		}
		if apierrors.IsNotFound(err) {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(crd); err != nil {
				log.Error(err, "Failed to set last applied annotation", "crd", crd)
			}
			if err := r.runtimeCli.Create(context.TODO(), crd); err != nil {
				return emperror.WrapWith(err, "creating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD created")
		} else {
			crd.ResourceVersion = current.ResourceVersion
			patchResult, err := patch.DefaultPatchMaker.Calculate(current, crd, patch.IgnoreStatusFields())
			if err != nil {
				log.Error(err, "could not match objects", "kind", crd.Spec.Names.Kind)
			} else if patchResult.IsEmpty() {
				log.V(1).Info("CRD is in sync")
				continue
			} else {
				log.V(1).Info("resource diffs",
					"patch", string(patchResult.Patch),
					"current", string(patchResult.Current),
					"modified", string(patchResult.Modified),
					"original", string(patchResult.Original))
			}

			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(crd); err != nil {
				log.Error(err, "Failed to set last applied annotation", "crd", crd)
			}

			if err := r.runtimeCli.Update(context.TODO(), crd); err != nil {
				if apierrors.IsConflict(err) || apierrors.IsInvalid(err) {
					err := r.runtimeCli.Delete(context.TODO(), crd)
					if err != nil {
						return emperror.WrapWith(err, "could not delete CRD", "kind", crd.Spec.Names.Kind)
					}
					crd.ResourceVersion = ""
					if err := r.runtimeCli.Create(context.TODO(), crd); err != nil {
						log.Info("resource needs to be re-created")
						return emperror.WrapWith(err, "creating CRD failed", "kind", crd.Spec.Names.Kind)
					}
					log.Info("CRD created")
				}

				return emperror.WrapWith(err, "updating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD updated")
		}
	}

	log.Info("Reconciled")
	return nil
}
