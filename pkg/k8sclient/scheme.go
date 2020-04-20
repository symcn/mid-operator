package k8sclient

import (
	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	//  monitorv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	configv1alpha2 "istio.io/client-go/pkg/apis/config/v1alpha2"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = apiextensionsv1beta1.AddToScheme(scheme)
	_ = devopsv1beta1.AddToScheme(scheme)

	// _ = monitorv1.AddToScheme(scheme)
	_ = networkingv1beta1.AddToScheme(scheme)
	_ = configv1alpha2.AddToScheme(scheme)
}

// GetScheme gets an initialized runtime.Scheme with k8s core added by default
func GetScheme() *runtime.Scheme {
	return scheme
}
