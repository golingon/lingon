// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package crd_test

import (
	"fmt"
	"os"
	"path/filepath"

	team "github.com/golingon/docskubernetes/crd/out"
	"github.com/golingon/lingon/pkg/kube"
)

var defaultOut = "out"

// Example_export to shows how to export to kubernetes manifests files in YAML.
func Example_export() {
	out := filepath.Join(defaultOut, "manifests")

	// Remove previously generated output directory
	_ = os.RemoveAll(out)

	tm := team.New()
	if err := kube.Export(
		tm,
		// directory where to export the manifests
		kube.WithExportOutputDirectory(out),

		// // Write the manifests to a bytes.Buffer.
		// // Note that it will be written as txtar format.
		// kube.WithExportWriter(buf),
		// // Add a kustomization.yaml file next to the manifests.
		// kube.WithExportKustomize(true),

		// // Write the manifest as a single file.
		// kube.WithExportAsSingleFile("manifest.yaml"),

		// // Write the manifests as JSON instead of YAML.
		// // Written as an JSON array if written as a single file.
		// kube.WithExportOutputJSON(true),

		// // Write to standard output instead of a file.
		// kube.WithExportStdOut(),

		// // Write the manifests as it would appear on the cluster.
		// // 1 resource per file and one folder per namespace.
		// kube.WithExportExplodeManifests(true),

		// // Define a hook to be called before exporting a secret.
		// // Note that this will remove the secret from the output.
		// kube.WithExportSecretHook(
		// 	func(s *corev1.Secret) error {
		// 		// do something with the secret
		// 		return nil
		// 	}),
	); err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Println("successfully exported CRDs to manifests")

	// Output:
	// successfully exported CRDs to manifests
}
