// Copyright (c) Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"io"
	"os"

	"github.com/golingon/lingon/pkg/kubeutil"
	corev1 "k8s.io/api/core/v1"
)

// ExportOption is used to configure conversion from Go code to kubernetes
// objects in YAML.
// Helpers function are provided to those field, see WithExportXXX functions
type ExportOption func(*goky)

// exportOption is used to configure the goky, all fields have sane defaults.
type exportOption struct {
	// OutputDir is the directory where the generated code will be written
	// (default: out)
	// ex: "./tekton"
	OutputDir string

	// Writer is used to write the exported manifests in txtar format
	// for more info on [golang.org/x/tools/txtar.Archive] format see:
	// https://pkg.go.dev/golang.org/x/tools/txtar Note that we are using
	// https://github.com/rogpeppe/go-internal/blob/master/txtar/ instead.
	//
	// ex: [os.Stdout], [bytes.Buffer]
	ManifestWriter io.Writer

	// NameFileFunc formats the name of the file containing the kubernetes
	// object
	NameFileFunc func(m *kubeutil.Metadata) string

	// SecretHook is used to process the secrets before they are exported.
	// The hook is called for each secret.
	// This is useful to redact the secrets in order not to save them in plain
	// text.
	SecretHook func(secret *corev1.Secret) error

	// Kustomize flag adds a kustomization.yaml file to the output
	Kustomize bool

	// Explode flag explodes files into multiple files according to their kind
	// and namespace.
	Explode bool

	// SingleFile flag will write all the manifests in a single file.
	// The file will be place in [OutputDir] and will be named as defined
	// in the parameter.
	// Note that this is not compatible with [Explode] flag.
	// It will ignore the use of [NameFileFunc].
	SingleFile string

	// OutputJSON sets the format of the output to JSON instead of YAML.
	// Not that in the case of exporting to a single file, the format will be
	// a JSON array of objects.
	OutputJSON bool
}

var exportDefaultOpts = exportOption{
	ManifestWriter: os.Stdout,
	OutputDir:      "out",
	NameFileFunc:   nil,
	SecretHook:     nil,
	Kustomize:      false,
	Explode:        false,
	SingleFile:     "out",
	OutputJSON:     false,
}

// WithExportNameFileFunc sets the function to format the name of the file
// containing the kubernetes object.
// Note that the files needs an extension to be added: ".yaml" or ".yml"
//
// Usage:
//
//	WithExportNameFileFunc(func(m *kubeutil.Metadata) string {
//		return fmt.Sprintf("%s-%s.yaml", strings.ToLower(m.Kind), m.Meta.Name)
//	})
func WithExportNameFileFunc(f func(m *kubeutil.Metadata) string) ExportOption {
	return func(g *goky) {
		g.o.NameFileFunc = f
	}
}

// WithExportExplodeManifests explodes the manifests into separate files
// organized by namespace to match closely the structure of the kubernetes
// cluster.
// See [Explode] for more info.
// Note that this option is incompatible [WithExportAsSingleFile].
func WithExportExplodeManifests(b bool) ExportOption {
	return func(g *goky) {
		g.o.Explode = b
		g.useSingleFile = false
	}
}

// WithExportWriter writes the generated manifests to [io.Writer].
// Note that the format is txtar, for more info on
// [golang.org/x/tools/txtar.Archive] format
// see: https://pkg.go.dev/golang.org/x/tools/txtar
//
// A txtar archive is zero or more comment lines and then a sequence of file
// entries. Each file entry begins with a file marker line of the form "--
// FILENAME --" and
// is followed by zero or more file content lines making up the file data.
// The comment or file content ends at the next file marker line.
// The file marker line must begin with the three-byte sequence "-- " and
// end with the three-byte sequence " --", but the enclosed file name can be
// surrounding by additional white space, all of which is stripped.
//
// If the txtar file is missing a trailing newline on the final line,
// parsers should consider a final newline to be present anyway.
//
// There are no possible syntax errors in a txtar archive.
func WithExportWriter(w io.Writer) ExportOption {
	return func(g *goky) {
		g.useWriter = true
		g.o.ManifestWriter = w
	}
}

// WithExportStdOut writes the generated manifests to [os.Stdout]
// Note that the format is txtar, for more info on
// [golang.org/x/tools/txtar.Archive] format
// see: https://pkg.go.dev/golang.org/x/tools/txtar
// See [WithExportWriter] for more info.
//
// If you want to write in the YAML format, use [WithExportAsSingleFile]
// instead.
func WithExportStdOut() ExportOption {
	return func(g *goky) {
		g.useWriter = true
		g.o.SingleFile = "stdout"
	}
}

// WithExportOutputJSON sets the format of the output to JSON instead of YAML.
// Not that in the case of exporting to a single file, the format will be
// a JSON array of objects.
func WithExportOutputJSON(b bool) ExportOption {
	return func(g *goky) {
		g.o.OutputJSON = b
	}
}

// WithExportAsSingleFile flag will write all the manifests in a single file
// Note that this is not compatible with [WithExportExplodeManifests] flag
//
// Usage:
//
//	err := Export(km,
//			WithExportOutputDirectory("./out"),
//			WithExportAsSingleFile("manifests.yaml"),
//			)
//
// the output file will be written to ./out/manifests.yaml
func WithExportAsSingleFile(name string) ExportOption {
	return func(g *goky) {
		g.useSingleFile = true
		if name != "" {
			g.o.SingleFile = name
		}
	}
}

// WithExportKustomize adds a kustomization.yaml file to the output.
func WithExportKustomize(b bool) ExportOption {
	return func(g *goky) {
		g.o.Kustomize = b
	}
}

// WithExportOutputDirectory sets the output directory for the generated
// manifests.
func WithExportOutputDirectory(dir string) ExportOption {
	return func(g *goky) {
		g.o.OutputDir = dir
	}
}

// WithExportSecretHook is used to process the secrets before they are exported.
// The hook is called for each secret.
// This is useful to send secret to a vault (pun intended)
// and not to save them in plain text.
// Base64 encoded secrets are not secure.
//
// NOTE: the secrets will *NOT* be written to the output directory or
// [io.Writer]
// if this option is used.
func WithExportSecretHook(f func(s *corev1.Secret) error) ExportOption {
	return func(g *goky) {
		g.o.SecretHook = f
	}
}
