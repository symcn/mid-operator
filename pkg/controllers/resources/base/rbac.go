package base

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *Reconciler) serviceAccount() runtime.Object {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      istioReaderServiceAccountName,
			Namespace: r.Config.Namespace,
			Labels:    istioReaderLabel,
		},
	}
}

func (r *Reconciler) clusterRole() runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   r.istioReaderNameWithNamespace(),
			Labels: istioReaderLabel,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"config.istio.io", "rbac.istio.io", "security.istio.io", "networking.istio.io", "authentication.istio.io"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes", "pods", "services", "endpoints", "replicationcontrollers"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"extensions", "apps"},
				Resources: []string{"replicasets"},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}
}

func (r *Reconciler) clusterRoleBinding() runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   r.istioReaderNameWithNamespace(),
			Labels: istioReaderLabel,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     r.istioReaderNameWithNamespace(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      istioReaderServiceAccountName,
				Namespace: r.Config.Namespace,
			},
		},
	}
}

func (r *Reconciler) istioReaderNameWithNamespace() string {
	return fmt.Sprintf("%s-%s", istioReaderName, r.Config.Namespace)
}
