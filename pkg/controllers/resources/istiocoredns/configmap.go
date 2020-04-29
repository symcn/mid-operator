package istiocoredns

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/symcn/mid-operator/pkg/controllers/resources/templates"
)

func (r *Reconciler) data() map[string]string {
	var data map[string]string
	if r.isProxyPluginDeprecated() {
		data = map[string]string{
			"Corefile": `.:53 {
          errors
          health
          grpc global 127.0.0.1:8053
          forward . /etc/resolv.conf {
            except global
          }
          prometheus :9153
          cache 30
          reload
        }
`,
		}
	} else {
		data = map[string]string{
			"Corefile": `.:53 {
    errors
    health
    proxy global 127.0.0.1:8053 {
        protocol grpc insecure
    }
    prometheus :9153
    proxy . /etc/resolv.conf
    cache 30
    reload
}
`,
		}
	}

	return data
}

func (r *Reconciler) configMap() runtime.Object {
	return &corev1.ConfigMap{
		ObjectMeta: templates.ObjectMeta(configMapName, labels, r.Config),
		Data:       r.data(),
	}
}
