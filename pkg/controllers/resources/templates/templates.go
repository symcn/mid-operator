package templates

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type MeshNetworkEndpoint struct {
	FromCIDR     string `json:"fromCidr,omitempty"`
	FromRegistry string `json:"fromRegistry,omitempty"`
}

type MeshNetworkGateway struct {
	Address string `json:"address"`
	Port    uint   `json:"port"`
}

type MeshNetwork struct {
	Endpoints []*MeshNetworkEndpoint `json:"endpoints,omitempty"`
	Gateways  []*MeshNetworkGateway  `json:"gateways,omitempty"`
}

type MeshNetworks struct {
	Networks map[string]*MeshNetwork `json:"networks"`
}

func GetMeshNetworks(config *devopsv1beta1.Istio) *MeshNetworks {
	meshNetworks := make(map[string]*MeshNetwork)

	localNetwork := &MeshNetwork{
		Endpoints: []*MeshNetworkEndpoint{
			{
				FromRegistry: config.Spec.ClusterName,
			},
		},
	}

	if len(config.Status.GatewayAddress) > 0 {
		gateways := make([]*MeshNetworkGateway, 0)
		for _, address := range config.Status.GatewayAddress {
			gateways = append(gateways, &MeshNetworkGateway{
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

	return &MeshNetworks{Networks: meshNetworks}
}

func GetMeshNetworksHash(config *devopsv1beta1.Istio) string {
	hash := ""
	j, err := json.Marshal(GetMeshNetworks(config))
	if err != nil {
		return hash
	}

	hash = fmt.Sprintf("%x", md5.Sum(j))

	return hash
}

func ObjectMeta(name string, labels map[string]string, config runtime.Object) metav1.ObjectMeta {
	obj := config.DeepCopyObject()
	objMeta, _ := meta.Accessor(obj)
	ovk := config.GetObjectKind().GroupVersionKind()

	return metav1.ObjectMeta{
		Name:      name,
		Namespace: objMeta.GetNamespace(),
		Labels:    labels,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion:         ovk.GroupVersion().String(),
				Kind:               ovk.Kind,
				Name:               objMeta.GetName(),
				UID:                objMeta.GetUID(),
				Controller:         utils.BoolPointer(true),
				BlockOwnerDeletion: utils.BoolPointer(true),
			},
		},
	}
}

func ObjectMetaWithAnnotations(name string, labels map[string]string, annotations map[string]string, config runtime.Object) metav1.ObjectMeta {
	o := ObjectMeta(name, labels, config)
	o.Annotations = annotations
	return o
}

func ObjectMetaClusterScope(name string, labels map[string]string, config runtime.Object) metav1.ObjectMeta {
	obj := config.DeepCopyObject()
	objMeta, _ := meta.Accessor(obj)
	ovk := config.GetObjectKind().GroupVersionKind()

	return metav1.ObjectMeta{
		Name:   name,
		Labels: labels,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion:         ovk.GroupVersion().String(),
				Kind:               ovk.Kind,
				Name:               objMeta.GetName(),
				UID:                objMeta.GetUID(),
				Controller:         utils.BoolPointer(true),
				BlockOwnerDeletion: utils.BoolPointer(true),
			},
		},
	}
}

func ControlPlaneAuthPolicy(istiodEnabled, controlPlaneSecurityEnabled bool) string {
	if !istiodEnabled && controlPlaneSecurityEnabled {
		return "MUTUAL_TLS"
	}
	return "NONE"
}
