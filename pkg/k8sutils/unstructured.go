package k8sutils

import (
	"reflect"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/symcn/mid-operator/pkg/k8sclient"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type DesiredState string

const (
	DesiredStatePresent DesiredState = "present"
	DesiredStateAbsent  DesiredState = "absent"
)

type DynamicObject struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Spec      map[string]interface{}
	Gvr       schema.GroupVersionResource
	Kind      string
	Owner     metav1.Object
}

func (d *DynamicObject) Reconcile(log logr.Logger, client dynamic.Interface, desiredState DesiredState) error {
	if desiredState == "" {
		desiredState = DesiredStatePresent
	}
	desired := d.unstructured()
	desiredType := reflect.TypeOf(desired)
	log = log.WithValues("type", reflect.TypeOf(d), "name", d.Name)
	current, err := client.Resource(d.Gvr).Namespace(d.Namespace).Get(d.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "name", d.Name, "kind", desiredType)
	}
	if apierrors.IsNotFound(err) {
		if desiredState == DesiredStatePresent {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "Failed to set last applied annotation", "desired", desired)
			}
			if _, err := client.Resource(d.Gvr).Namespace(d.Namespace).Create(desired, metav1.CreateOptions{}); err != nil {
				return emperror.WrapWith(err, "creating resource failed", "name", d.Name, "kind", desiredType)
			}
			log.Info("resource created", "kind", d.Gvr.Resource)
		}
	} else {
		if desiredState == DesiredStatePresent {
			patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired, patch.IgnoreStatusFields())
			if err != nil {
				log.Error(err, "could not match objects", "kind", desiredType)
			} else if patchResult.IsEmpty() {
				log.V(1).Info("resource is in sync")
				return nil
			} else {
				log.V(1).Info("resource diffs",
					"patch", string(patchResult.Patch),
					"current", string(patchResult.Current),
					"modified", string(patchResult.Modified),
					"original", string(patchResult.Original))
			}
			// Need to set this before resourceversion is set, as it would constantly change otherwise
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "Failed to set last applied annotation", "desired", desired)
			}
			desired.SetResourceVersion(current.GetResourceVersion())
			if _, err := client.Resource(d.Gvr).Namespace(d.Namespace).Update(desired, metav1.UpdateOptions{}); err != nil {
				return emperror.WrapWith(err, "updating resource failed", "name", d.Name, "kind", desiredType)
			}
			log.Info("resource updated", "kind", d.Gvr.Resource)
		} else if desiredState == DesiredStateAbsent {
			if err := client.Resource(d.Gvr).Namespace(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{}); err != nil {
				return emperror.WrapWith(err, "deleting resource failed", "name", d.Name, "kind", desiredType)
			}
			log.Info("resource deleted", "kind", d.Gvr.Resource)
		}
	}
	return nil
}

func (d *DynamicObject) unstructured() *unstructured.Unstructured {
	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": d.Spec,
		},
	}
	u.SetName(d.Name)
	if len(d.Namespace) > 0 {
		u.SetNamespace(d.Namespace)
	}
	if d.Labels != nil {
		u.SetLabels(d.Labels)
	}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   d.Gvr.Group,
		Version: d.Gvr.Version,
		Kind:    d.Kind,
	})

	ro, ok := d.Owner.(runtime.Object)
	if !ok {
		klog.Errorf("is not a %T a runtime.Object, cannot call SetControllerReference", d.Owner)
		return nil
	}

	gvk, err := apiutil.GVKForObject(ro, k8sclient.GetScheme())
	if err != nil {
		klog.Errorf("cannot get gvk runtime.Object %T, err: %+v", ro, err)
		return nil
	}

	// Create a new ref
	ref := *metav1.NewControllerRef(d.Owner, schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind})
	u.SetOwnerReferences([]metav1.OwnerReference{ref})
	return u
}
