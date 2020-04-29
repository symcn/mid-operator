package istio

import (
	"context"
	"errors"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetMeshGatewayAddress(client client.Client, key client.ObjectKey) ([]string, error) {
	var mgw devopsv1beta1.MeshGateway

	ips := make([]string, 0)

	err := client.Get(context.TODO(), key, &mgw)
	if err != nil && !k8serrors.IsNotFound(err) {
		return ips, err
	}

	if mgw.Status.Status != devopsv1beta1.Available {
		return ips, errors.New("gateway is pending")
	}

	ips = mgw.Status.GatewayAddress

	return ips, nil
}
