// +build ignore

//go:generate go run generate.go

package main

import (
	"log"
	"net/http"
	"path"
	"path/filepath"
	"runtime"

	"github.com/shurcooL/vfsgen"
)

var IstioCRDs http.FileSystem = http.Dir(path.Join(getRepoRoot(), "pkg/static/istio-crds/assets"))

var MidCRDs http.FileSystem = http.Dir(path.Join(getRepoRoot(), "config/crd/bases"))

func main() {
	rootDir := getRepoRoot()
	log.Printf("rootDir: %s", rootDir)

	err := vfsgen.Generate(IstioCRDs, vfsgen.Options{
		Filename:     path.Join(rootDir, "pkg/static/istio-crds/generated/istio-crds.gogen.go"),
		PackageName:  "generated",
		VariableName: "IstioCRDs",
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = vfsgen.Generate(MidCRDs, vfsgen.Options{
		Filename:     path.Join(rootDir, "pkg/static/mid-crds/generated/mid-crds.gogen.go"),
		PackageName:  "generated",
		VariableName: "MidCRDs",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

// getRepoRoot returns the full path to the root of the repo
func getRepoRoot() string {
	// +nolint
	_, filename, _, _ := runtime.Caller(0)

	dir := filepath.Dir(filename)

	return filepath.Dir(path.Join(dir, ".."))
}
