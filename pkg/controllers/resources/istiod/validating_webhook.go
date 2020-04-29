package istiod

import (
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) webhooks() []admissionv1beta1.ValidatingWebhook {
	// ignore := admissionv1beta1.Fail
	se := admissionv1beta1.SideEffectClassNone
	return []admissionv1beta1.ValidatingWebhook{
		{
			Name: "validation.istio.io",
			ClientConfig: admissionv1beta1.WebhookClientConfig{
				Service: &admissionv1beta1.ServiceReference{
					Name:      ServiceNameIstiod,
					Namespace: r.Config.Namespace,
					Path:      utils.StrPointer("/validate"),
				},
				CABundle: nil,
			},
			Rules: []admissionv1beta1.RuleWithOperations{
				{
					Operations: []admissionv1beta1.OperationType{
						admissionv1beta1.Create,
						admissionv1beta1.Update,
					},
					Rule: admissionv1beta1.Rule{
						APIGroups:   []string{"config.istio.io", "rbac.istio.io", "security.istio.io", "authentication.istio.io", "networking.istio.io"},
						APIVersions: []string{"*"},
						Resources:   []string{"*"},
					},
				},
			},
			FailurePolicy: nil,
			SideEffects:   &se,
		},
	}
}

func (r *Reconciler) validatingWebhook() runtime.Object {
	return &admissionv1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: templates.ObjectMetaClusterScope(validatingWebhookName, utils.MergeStringMaps(istiodLabels, istiodLabelSelector), r.Config),
		Webhooks:   r.webhooks(),
	}
}
