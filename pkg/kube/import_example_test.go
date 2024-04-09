// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package kube_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golingon/lingon/pkg/kube"
	"github.com/golingon/lingon/pkg/kubeutil"
	"golang.org/x/tools/txtar"
)

func ExampleImport() {
	out := filepath.Join("out", "tekton")
	_ = os.RemoveAll(out)
	defer os.RemoveAll(out)
	err := kube.Import(
		// name of the application
		kube.WithImportAppName("tekton"),
		// name of the Go package where the code is generated to
		kube.WithImportPackageName("tekton"),
		// the directory to write the generated code to
		kube.WithImportOutputDirectory(out),
		// the list of manifest files to read and convert
		kube.WithImportManifestFiles([]string{"testdata/tekton.yaml"}),
		// define the types for the CRDs
		kube.WithImportSerializer(defaultSerializer()),
		// will try to remove "tekton" from the name of the variable in the Go
		// code, make them shorter
		kube.WithImportRemoveAppName(true),
		// group all the resources from the same kind into one file each
		// example: 10 ConfigMaps => 1 file "config-map.go" containing 10
		// variables storing ConfigMap, etc...
		kube.WithImportGroupByKind(true),
		// add convenience methods to the App struct
		kube.WithImportAddMethods(true),
		// do not print verbose information
		kube.WithImportVerbose(false),
		// do not ignore errors
		kube.WithImportIgnoreErrors(false),
		// just for example purposes
		// how to create a logger (see
		// [golang.org/x/tools/slog](https://golang.org/x/tools/slog))
		// this has no effect with WithImportVerbose(false)
		kube.WithImportLogger(kube.Logger(os.Stderr)),
		// remove the status field and
		// other output-only fields from the manifest code before importing it.
		// Note that ConfigMap are not cleaned up as the comments will be lost.
		kube.WithImportCleanUp(true),
	)
	if err != nil {
		panic(fmt.Errorf("import: %w", err))
	}
	got, err := kubeutil.ListGoFiles(out)
	if err != nil {
		panic(fmt.Errorf("list go files: %w", err))
	}
	// sort the files to make the output deterministic
	sort.Strings(got)

	for _, f := range got {
		fmt.Println(f)
	}

	// Output:
	//
	// out/tekton/app.go
	// out/tekton/cluster-role-binding.go
	// out/tekton/cluster-role.go
	// out/tekton/config-map.go
	// out/tekton/custom-resource-definition.go
	// out/tekton/deployment.go
	// out/tekton/horizontal-pod-autoscaler.go
	// out/tekton/mutating-webhook-configuration.go
	// out/tekton/namespace.go
	// out/tekton/role-binding.go
	// out/tekton/role.go
	// out/tekton/secret.go
	// out/tekton/service-account.go
	// out/tekton/service.go
	// out/tekton/validating-webhook-configuration.go
}

func ExampleImport_withWriter() {
	filename := "testdata/grafana.yaml"
	file, _ := os.Open(filename)

	var buf bytes.Buffer

	err := kube.Import(
		kube.WithImportAppName("grafana"),
		kube.WithImportPackageName("grafana"),
		// will prefix all files with path "manifests/"
		kube.WithImportOutputDirectory("manifests/"),
		// we could just use kube.WithImportManifestFiles([]string{filename})
		// but this is just an example to show how to use WithImportReader
		// and WithImportWriter
		kube.WithImportReader(file),
		kube.WithImportWriter(&buf),
		// We don't want to group the resources by kind,
		// each file will contain a single resource
		kube.WithImportGroupByKind(false),
		// We rename the files to avoid name collisions.
		// Tip: use the Kind and Name of the resource to
		// create a unique name and avoid collision.
		//
		// Here, we didn't use WithImportGroupByKind,
		// each file will contain a single resource.
		kube.WithImportNameFileFunc(
			func(m kubeutil.Metadata) string {
				return fmt.Sprintf(
					"%s-%s.go",
					strings.ToLower(m.Kind),
					m.Meta.Name,
				)
			},
		),
	)
	if err != nil {
		panic("failed to import")
	}

	// the output contained in bytes.Buffer is in the txtar format
	// for more details, see https://pkg.go.dev/golang.org/x/tools/txtar
	ar := txtar.Parse(buf.Bytes())

	// sort the files to make the output deterministic
	sort.SliceStable(
		ar.Files, func(i, j int) bool {
			return ar.Files[i].Name < ar.Files[j].Name
		},
	)
	for _, f := range ar.Files {
		fmt.Println(f.Name)
	}
	// Output:
	//
	// manifests/app.go
	// manifests/clusterrole-grafana-clusterrole.go
	// manifests/clusterrolebinding-grafana-clusterrolebinding.go
	// manifests/configmap-grafana-dashboards-default.go
	// manifests/configmap-grafana-test.go
	// manifests/configmap-grafana.go
	// manifests/deployment-grafana.go
	// manifests/pod-grafana-test.go
	// manifests/role-grafana.go
	// manifests/rolebinding-grafana.go
	// manifests/secret-grafana.go
	// manifests/service-grafana.go
	// manifests/serviceaccount-grafana-test.go
	// manifests/serviceaccount-grafana.go
}
