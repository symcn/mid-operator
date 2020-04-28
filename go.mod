module github.com/symcn/mid-operator

go 1.14

require (
	github.com/go-logr/logr v0.1.0
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	istio.io/client-go v0.0.0-20200316192452-065c59267750
	k8s.io/api v0.17.2
	k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.0
)
