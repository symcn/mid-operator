package static

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/symcn/mid-operator/pkg/k8sclient"
	istioCrds "github.com/symcn/mid-operator/pkg/static/istio-crds/generated"
	midCrds "github.com/symcn/mid-operator/pkg/static/mid-crds/generated"
	"github.com/symcn/mid-operator/pkg/utils"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func load(f io.Reader) ([]*extensionsobj.CustomResourceDefinition, error) {
	var b bytes.Buffer

	var yamls []string

	crds := make([]*extensionsobj.CustomResourceDefinition, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			yamls = append(yamls, b.String())
			b.Reset()
		} else {
			if _, err := b.WriteString(line); err != nil {
				return crds, err
			}
			if _, err := b.WriteString("\n"); err != nil {
				return crds, err
			}
		}
	}
	if s := strings.TrimSpace(b.String()); s != "" {
		yamls = append(yamls, s)
	}

	for _, yaml := range yamls {
		s := json.NewSerializerWithOptions(json.DefaultMetaFactory,
			k8sclient.GetScheme(), k8sclient.GetScheme(), json.SerializerOptions{Yaml: true})

		obj, _, err := s.Decode([]byte(yaml), nil, nil)
		if err != nil {
			continue
		}

		var crd *extensionsobj.CustomResourceDefinition
		var ok bool
		if crd, ok = obj.(*extensionsobj.CustomResourceDefinition); !ok {
			continue
		}

		crd.Status = extensionsobj.CustomResourceDefinitionStatus{}
		crd.SetGroupVersionKind(schema.GroupVersionKind{})
		labels := crd.GetLabels()
		labels[utils.CreatedByLabel] = utils.CreatedBy
		crds = append(crds, crd)
	}

	return crds, nil
}

func LoadIstioCRDs() ([]*extensionsobj.CustomResourceDefinition, error) {
	crds := make([]*extensionsobj.CustomResourceDefinition, 0)
	dir, err := istioCrds.IstioCRDs.Open("/")
	if err != nil {
		return crds, err
	}

	dirFiles, err := dir.Readdir(-1)
	if err != nil {
		return crds, err
	}
	for _, file := range dirFiles {
		f, err := istioCrds.IstioCRDs.Open(file.Name())
		if err != nil {
			return crds, err
		}

		tmp, err := load(f)
		if err != nil {
			return crds, err
		}

		crds = append(crds, tmp...)
	}

	return crds, nil
}

func LoadMidCRDs() ([]*extensionsobj.CustomResourceDefinition, error) {
	crds := make([]*extensionsobj.CustomResourceDefinition, 0)
	dir, err := midCrds.MidCRDs.Open("/")
	if err != nil {
		return crds, err
	}

	dirFiles, err := dir.Readdir(-1)
	if err != nil {
		return crds, err
	}
	for _, file := range dirFiles {
		f, err := midCrds.MidCRDs.Open(file.Name())
		if err != nil {
			return crds, err
		}

		tmp, err := load(f)
		if err != nil {
			return crds, err
		}

		crds = append(crds, tmp...)
	}

	return crds, nil
}
