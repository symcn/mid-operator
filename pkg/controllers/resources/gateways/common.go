package gateways

import (
	"fmt"

	"github.com/symcn/mid-operator/pkg/utils"
)

func (r *Reconciler) gatewayName() string {
	return r.gw.Name
}

func (r *Reconciler) serviceAccountName() string {
	return fmt.Sprintf("%s-service-account", r.gw.Name)
}

func (r *Reconciler) labels() map[string]string {
	return utils.MergeStringMaps(map[string]string{
		"gateway-name": r.gatewayName(),
		"gateway-type": string(r.gw.Spec.Type),
	}, r.gw.Spec.Labels)
}

func (r *Reconciler) clusterRoleName() string {
	return fmt.Sprintf("%s-cluster-role", r.gw.Name)
}

func (r *Reconciler) clusterRoleBindingName() string {
	return fmt.Sprintf("%s-cluster-role-binding", r.gw.Name)
}

func (r *Reconciler) roleName() string {
	return fmt.Sprintf("%s-role-sds", r.gw.Name)
}

func (r *Reconciler) roleBindingName() string {
	return fmt.Sprintf("%s-role-binding-sds", r.gw.Name)
}

func (r *Reconciler) labelSelector() map[string]string {
	return r.labels()
}

func (r *Reconciler) pdbName() string {
	return r.gw.Name
}

func (r *Reconciler) hpaName() string {
	return fmt.Sprintf("%s-autoscaler", r.gw.Name)
}
