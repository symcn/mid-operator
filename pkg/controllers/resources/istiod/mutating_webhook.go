package istiod

import (
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
)

const (
	ServiceNameInjector = "istio-sidecar-injector"
)

func (r *Reconciler) mutatingWebhook() runtime.Object {
	fail := admissionv1beta1.Fail
	unknownSideEffects := admissionv1beta1.SideEffectClassUnknown
	service := ServiceNameInjector
	if !utils.PointerToBool(r.Config.Spec.SidecarInjector.Enabled) && utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		service = ServiceNameIstiod
	}
	webhook := &admissionv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: templates.ObjectMetaClusterScope(ServiceNameInjector, sidecarInjectorLabels, r.Config),
		Webhooks: []admissionv1beta1.MutatingWebhook{
			{
				Name: "sidecar-injector.istio.io",
				ClientConfig: admissionv1beta1.WebhookClientConfig{
					Service: &admissionv1beta1.ServiceReference{
						Name:      service,
						Namespace: r.Config.Namespace,
						Path:      utils.StrPointer("/inject"),
					},
					CABundle: nil,
				},
				Rules: []admissionv1beta1.RuleWithOperations{
					{
						Operations: []admissionv1beta1.OperationType{
							admissionv1beta1.Create,
						},
						Rule: admissionv1beta1.Rule{
							Resources:   []string{"pods"},
							APIGroups:   []string{""},
							APIVersions: []string{"*"},
						},
					},
				},
				FailurePolicy: &fail,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"istio-injection": "enabled",
					},
				},
				SideEffects: &unknownSideEffects,
			},
		},
	}

	if utils.PointerToBool(r.Config.Spec.SidecarInjector.EnableNamespacesByDefault) {
		webhook.Webhooks[0].NamespaceSelector = &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "name",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{r.Config.Namespace},
				},
				{
					Key:      "istio-injection",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{"disabled"},
				},
			},
		}
	}

	if utils.PointerToBool(r.Config.Spec.Istiod.Enabled) {
		webhook.Webhooks[0].NamespaceSelector.MatchExpressions = append(webhook.Webhooks[0].NamespaceSelector.MatchExpressions, []metav1.LabelSelectorRequirement{
			{
				Key:      "istio-env",
				Operator: metav1.LabelSelectorOpDoesNotExist,
			},
			{
				Key:      "istio.io/rev",
				Operator: metav1.LabelSelectorOpDoesNotExist,
			},
		}...)
	}

	return webhook
}
