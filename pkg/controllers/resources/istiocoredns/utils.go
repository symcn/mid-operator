package istiocoredns

import (
	"strings"

	"github.com/coreos/go-semver/semver"

	"github.com/symcn/mid-operator/pkg/utils"
)

// Removed support for the proxy plugin: https://coredns.io/2019/03/03/coredns-1.4.0-release/
func (r *Reconciler) isProxyPluginDeprecated() bool {
	imageParts := strings.Split(utils.PointerToString(r.Config.Spec.IstioCoreDNS.Image), ":")
	tag := imageParts[1]

	v140 := semver.New("1.4.0")
	vCoreDNSTag := semver.New(tag)

	if v140.LessThan(*vCoreDNSTag) {
		return true
	}

	return false
}
