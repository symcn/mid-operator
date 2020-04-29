package istiocoredns

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/mholt/caddy/caddyfile"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/symcn/mid-operator/pkg/k8sutils"
)

// Add/Remove global:53 to/from coredns configmap
func (r *Reconciler) reconcileCoreDNSConfigMap(log logr.Logger, desiredState k8sutils.DesiredState) error {
	var cm corev1.ConfigMap

	err := r.Client.Get(context.Background(), types.NamespacedName{
		Name:      "coredns",
		Namespace: "kube-system",
	}, &cm)
	if k8serrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return emperror.Wrap(err, "could not get coredns configmap")
	}

	corefile := cm.Data["Corefile"]
	desiredCorefile := corefile
	clusterIP := ""

	if desiredState == k8sutils.DesiredStatePresent {
		var svc corev1.Service
		err = r.Client.Get(context.Background(), types.NamespacedName{
			Name:      serviceName,
			Namespace: r.Config.Namespace,
		}, &svc)
		if err != nil {
			return emperror.Wrap(err, "could not get Istio coreDNS service")
		}
		clusterIP = svc.Spec.ClusterIP
	}

	proxyOrForward := "proxy"
	if r.isProxyPluginDeprecated() {
		proxyOrForward = "forward"
	}
	config := caddyfile.EncodedServerBlock{
		Keys: []string{"global:53"},
		Body: [][]interface{}{
			{"errors"},
			{"cache", "30"},
			{proxyOrForward, ".", clusterIP},
		},
	}

	if desiredState == k8sutils.DesiredStatePresent {
		desiredCorefile, err = r.updateCorefile([]byte(corefile), config, false)
		if err != nil {
			return emperror.Wrap(err, "could not add config for .global to Corefile")
		}
	} else if desiredState == k8sutils.DesiredStateAbsent {
		desiredCorefile, err = r.updateCorefile([]byte(corefile), config, true)
		if err != nil {
			return emperror.Wrap(err, "could not remove config for .global from Corefile")
		}
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string, 0)
	}
	cm.Data["Corefile"] = desiredCorefile

	err = r.Client.Update(context.Background(), &cm)
	if err != nil {
		return emperror.Wrap(err, "could not update coredns configmap")
	}

	return nil
}

func (r *Reconciler) updateCorefile(corefile []byte, config caddyfile.EncodedServerBlock, remove bool) (string, error) {
	corefileJSONData, err := caddyfile.ToJSON(corefile)
	if err != nil {
		return "", emperror.Wrap(err, "could not convert Corefile to JSON data")
	}

	var corefileJSON caddyfile.EncodedCaddyfile
	err = json.Unmarshal(corefileJSONData, &corefileJSON)
	if err != nil {
		return "", emperror.Wrap(err, "could not unmarshal JSON to EncodedCaddyfile")
	}

	if len(config.Keys) < 1 {
		return "", errors.New("invalid .global config")
	}

	pos := -1
	for i, block := range corefileJSON {
		if len(block.Keys) < 1 {
			continue
		}
		if block.Keys[0] == config.Keys[0] {
			pos = i
			break
		}
	}

	if remove {
		if pos > 0 {
			corefileJSON = append(corefileJSON[:pos], corefileJSON[pos+1:]...)
		}
	} else {
		if pos > 0 {
			corefileJSON[pos] = config
		} else {
			corefileJSON = append(corefileJSON, config)
		}
	}

	corefileData, err := json.Marshal(&corefileJSON)
	if err != nil {
		return "", emperror.Wrap(err, "could not marshal EncodedCaddyfile to JSON")
	}

	corefile, err = caddyfile.FromJSON(corefileData)
	if err != nil {
		return "", emperror.Wrap(err, "could not convert JSON to Caddyfile")
	}

	// convert tabs to spaces for properly display content in ConfigMap
	return strings.Replace(string(corefile), "\t", "  ", -1), nil
}
