// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package kube_test

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/txtar"
	"github.com/volvo-cars/lingon/pkg/kube"
	"github.com/volvo-cars/lingon/pkg/kubeutil"
	tu "github.com/volvo-cars/lingon/pkg/testutil"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsbeta "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

const defaultImportOutputDir = "out/jamel"

func defaultSerializer() runtime.Decoder {
	// NEEDED FOR CRDS
	//

	_ = apiextensions.AddToScheme(scheme.Scheme)
	_ = apiextensionsbeta.AddToScheme(scheme.Scheme)
	return scheme.Codecs.UniversalDeserializer()
}

func TestImport(t *testing.T) {
	type args struct {
		Name     string
		OutDir   string
		Opts     []kube.ImportOption
		OutFiles []string
	}
	TT := []args{
		{
			Name:   "with CRDs and remove app name and group by kind",
			OutDir: filepath.Join(defaultImportOutputDir, "argocd"),
			Opts: []kube.ImportOption{
				kube.WithImportAppName("argocd"),
				kube.WithImportManifestFiles([]string{"testdata/argocd.yaml"}),
				kube.WithImportSerializer(defaultSerializer()),
				kube.WithImportRemoveAppName(true),
				kube.WithImportGroupByKind(true),
			},
			OutFiles: []string{
				"out/jamel/argocd/app.go",
				"out/jamel/argocd/cluster-role-binding.go",
				"out/jamel/argocd/cluster-role.go",
				"out/jamel/argocd/config-map.go",
				"out/jamel/argocd/custom-resource-definition.go",
				"out/jamel/argocd/deployment.go",
				"out/jamel/argocd/network-policy.go",
				"out/jamel/argocd/role-binding.go",
				"out/jamel/argocd/role.go",
				"out/jamel/argocd/secret.go",
				"out/jamel/argocd/service-account.go",
				"out/jamel/argocd/service.go",
				"out/jamel/argocd/stateful-set.go",
			},
		}, {
			Name: "with CRDs and remove app name containing dash and group by kind",
			OutDir: filepath.Join(
				defaultImportOutputDir,
				"external-secrets",
			),
			Opts: []kube.ImportOption{
				kube.WithImportAppName("external-secrets"),
				kube.WithImportManifestFiles([]string{"testdata/external-secrets.yaml"}),
				kube.WithImportSerializer(defaultSerializer()),
				kube.WithImportRemoveAppName(true),
				kube.WithImportGroupByKind(true),
			},
			OutFiles: []string{
				"out/jamel/external-secrets/app.go",
				"out/jamel/external-secrets/cluster-role-binding.go",
				"out/jamel/external-secrets/cluster-role.go",
				"out/jamel/external-secrets/custom-resource-definition.go",
				"out/jamel/external-secrets/deployment.go",
				"out/jamel/external-secrets/role-binding.go",
				"out/jamel/external-secrets/role.go",
				"out/jamel/external-secrets/secret.go",
				"out/jamel/external-secrets/service-account.go",
				"out/jamel/external-secrets/service.go",
				"out/jamel/external-secrets/validating-webhook-configuration.go",
			},
		}, {
			Name:   "with CRDs and remove app name and split by name",
			OutDir: filepath.Join(defaultImportOutputDir, "karpenter"),
			Opts: []kube.ImportOption{
				kube.WithImportAppName("karpenter"),
				kube.WithImportPackageName("karpenter"),
				kube.WithImportManifestFiles([]string{"testdata/karpenter.yaml"}),
				kube.WithImportSerializer(defaultSerializer()),
				kube.WithImportRemoveAppName(true),
			},
			OutFiles: []string{
				"out/jamel/karpenter/admin_cr.go",
				"out/jamel/karpenter/app.go",
				"out/jamel/karpenter/cert_secrets.go",
				"out/jamel/karpenter/config-logging_cm.go",
				"out/jamel/karpenter/core_cr.go",
				"out/jamel/karpenter/core_crb.go",
				"out/jamel/karpenter/cr.go",
				"out/jamel/karpenter/crb.go",
				"out/jamel/karpenter/defaulting.webhook..k8s.aws_mutatingwebhookconfigurations.go",
				"out/jamel/karpenter/defaulting.webhook..sh_mutatingwebhookconfigurations.go",
				"out/jamel/karpenter/deploy.go",
				"out/jamel/karpenter/dns_rb.go",
				"out/jamel/karpenter/dns_role.go",
				"out/jamel/karpenter/global-settings_cm.go",
				"out/jamel/karpenter/pdb.go",
				"out/jamel/karpenter/rb.go",
				"out/jamel/karpenter/role.go",
				"out/jamel/karpenter/sa.go",
				"out/jamel/karpenter/svc.go",
				"out/jamel/karpenter/validation.webhook..k8s.aws_validatingwebhookconfigurations.go",
				"out/jamel/karpenter/validation.webhook..sh_validatingwebhookconfigurations.go",
				"out/jamel/karpenter/validation.webhook.config..sh_validatingwebhookconfigurations.go",
			},
		}, {
			Name:   "with vanilla serializer and remove app name and group by kind",
			OutDir: filepath.Join(defaultImportOutputDir, "grafana"),
			Opts: []kube.ImportOption{
				kube.WithImportAppName("grafana"),
				kube.WithImportPackageName("grafana"),
				kube.WithImportManifestFiles([]string{"testdata/grafana.yaml"}),
				kube.WithImportRemoveAppName(true),
				kube.WithImportGroupByKind(true),
			},
			OutFiles: []string{
				"out/jamel/grafana/app.go",
				"out/jamel/grafana/cluster-role-binding.go",
				"out/jamel/grafana/cluster-role.go",
				"out/jamel/grafana/config-map.go",
				"out/jamel/grafana/deployment.go",
				"out/jamel/grafana/pod.go",
				"out/jamel/grafana/role-binding.go",
				"out/jamel/grafana/role.go",
				"out/jamel/grafana/secret.go",
				"out/jamel/grafana/service-account.go",
				"out/jamel/grafana/service.go",
			},
		}, {
			Name:   "with vanilla serializer and add methods",
			OutDir: filepath.Join(defaultImportOutputDir, "manifester"),
			Opts: []kube.ImportOption{
				kube.WithImportAppName("grafana"),
				kube.WithImportPackageName("grafana"),
				kube.WithImportManifestFiles([]string{"testdata/grafana.yaml"}),
				kube.WithImportRemoveAppName(true),
				kube.WithImportGroupByKind(true),
				kube.WithImportAddMethods(true),
			},
			OutFiles: []string{
				"out/jamel/manifester/app.go",
				"out/jamel/manifester/cluster-role-binding.go",
				"out/jamel/manifester/cluster-role.go",
				"out/jamel/manifester/config-map.go",
				"out/jamel/manifester/deployment.go",
				"out/jamel/manifester/pod.go",
				"out/jamel/manifester/role-binding.go",
				"out/jamel/manifester/role.go",
				"out/jamel/manifester/secret.go",
				"out/jamel/manifester/service-account.go",
				"out/jamel/manifester/service.go",
			},
		},
	}

	for _, tt := range TT {
		tc := tt
		t.Run(
			tt.Name, func(t *testing.T) {
				t.Parallel()
				var buf bytes.Buffer
				tc.Opts = append(
					tc.Opts,
					kube.WithImportOutputDirectory(tc.OutDir),
					kube.WithImportWriter(&buf),
				)
				err := kube.Import(tc.Opts...)
				tu.AssertNoError(t, err, "failed to import")
				ar := txtar.Parse(buf.Bytes())
				got := make([]string, 0, len(ar.Files))
				for _, f := range ar.Files {
					got = append(got, f.Name)
				}
				sort.Strings(got)
				want := tc.OutFiles
				tu.AssertEqualSlice(t, want, got)
			},
		)
	}
}

func TestImport_Error(t *testing.T) {
	err := kube.Import(
		kube.WithImportAppName("foo-app"),
		kube.WithImportPackageName("foo-package"),
		kube.WithImportManifestFiles([]string{"does-not-exists.yaml"}),
	)
	tu.IsNotNil(t, err)
	output := `import options: package name cannot contain a dash
file does not exist: does-not-exists.yaml`
	tu.AssertEqual(t, output, err.Error())
}

func TestImport_ErrorEmptyManifest(t *testing.T) {
	pathg := "testdata/golden/empty.golden"
	pathy := "testdata/golden/empty.yaml"
	golden, err := os.ReadFile(pathg)
	tu.AssertNoError(t, err, "read golden file")
	var buf bytes.Buffer
	err = kube.Import(
		kube.WithImportAppName("foo-app"),
		kube.WithImportPackageName("foopackage"),
		kube.WithImportManifestFiles([]string{pathy}),
		kube.WithImportWriter(&buf),
	)
	tu.AssertNoError(t, err, "failed to import")
	tu.AssertEqual(t, string(golden), buf.String())
}

func TestJamel_SaveFromReader(t *testing.T) {
	filename := "testdata/grafana.yaml"
	file, err := os.Open(filename)
	tu.AssertNoError(t, err, fmt.Sprintf("failed to open file: %s", filename))
	out := filepath.Join(defaultImportOutputDir, "reader")
	var buf bytes.Buffer
	err = kube.Import(
		kube.WithImportAppName("grafana"),
		kube.WithImportPackageName("grafana"),
		kube.WithImportOutputDirectory(out),
		kube.WithImportWriter(&buf),
		kube.WithImportReader(file),
		kube.WithImportNameFieldFunc(kube.NameFieldFunc),
		kube.WithImportNameVarFunc(kube.NameVarFunc),
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
	tu.AssertNoError(t, err, "failed to import")

	ar := txtar.Parse(buf.Bytes())
	got := make([]string, 0, len(ar.Files))
	for _, f := range ar.Files {
		got = append(got, f.Name)
	}
	sort.Strings(got)

	want := []string{
		"out/jamel/reader/app.go",
		"out/jamel/reader/clusterrole-grafana-clusterrole.go",
		"out/jamel/reader/clusterrolebinding-grafana-clusterrolebinding.go",
		"out/jamel/reader/configmap-grafana-dashboards-default.go",
		"out/jamel/reader/configmap-grafana-test.go",
		"out/jamel/reader/configmap-grafana.go",
		"out/jamel/reader/deployment-grafana.go",
		"out/jamel/reader/pod-grafana-test.go",
		"out/jamel/reader/role-grafana.go",
		"out/jamel/reader/rolebinding-grafana.go",
		"out/jamel/reader/secret-grafana.go",
		"out/jamel/reader/service-grafana.go",
		"out/jamel/reader/serviceaccount-grafana-test.go",
		"out/jamel/reader/serviceaccount-grafana.go",
	}

	tu.AssertEqualSlice(t, want, got)
}

func TestJamel_MissingCRDs(t *testing.T) {
	filename := "testdata/istio.yaml"
	file, err := os.Open(filename)
	tu.AssertNoError(t, err, fmt.Sprintf("failed to open file: %s", filename))

	var buf bytes.Buffer

	err = kube.Import(
		kube.WithImportAppName("istio"),
		kube.WithImportPackageName("istio"),
		kube.WithImportOutputDirectory("manifests/"),
		kube.WithImportReader(file),
		kube.WithImportWriter(&buf),
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
	errmsg := "generate go: stdin: " +
		"no kind \"EnvoyFilter\" is registered for version " +
		"\"networking.istio.io/v1alpha3\" in scheme \"pkg/runtime/scheme.go:100\""
	tu.AssertError(t, err, errmsg)
}

func TestJamel_ReaderWriter(t *testing.T) {
	filename := "testdata/grafana.yaml"
	file, err := os.Open(filename)
	tu.AssertNoError(t, err, fmt.Sprintf("failed to open file: %s", filename))

	var buf bytes.Buffer

	err = kube.Import(
		kube.WithImportAppName("grafana"),
		kube.WithImportPackageName("grafana"),
		kube.WithImportOutputDirectory("manifests/"),
		kube.WithImportReader(file),
		kube.WithImportWriter(&buf),
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
	tu.AssertNoError(t, err, "failed to import")

	want := []string{
		"manifests/app.go",
		"manifests/clusterrole-grafana-clusterrole.go",
		"manifests/clusterrolebinding-grafana-clusterrolebinding.go",
		"manifests/configmap-grafana-dashboards-default.go",
		"manifests/configmap-grafana-test.go",
		"manifests/configmap-grafana.go",
		"manifests/deployment-grafana.go",
		"manifests/pod-grafana-test.go",
		"manifests/role-grafana.go",
		"manifests/rolebinding-grafana.go",
		"manifests/secret-grafana.go",
		"manifests/service-grafana.go",
		"manifests/serviceaccount-grafana-test.go",
		"manifests/serviceaccount-grafana.go",
	}

	ar := txtar.Parse(buf.Bytes())
	got := make([]string, len(ar.Files))
	for i, f := range ar.Files {
		got[i] = f.Name
	}
	sort.Strings(got)
	tu.AssertEqualSlice(t, want, got)
}

func TestJamel_ConfigMapComments(t *testing.T) {
	out := filepath.Join(defaultImportOutputDir, "tekton")
	tu.AssertNoError(t, os.RemoveAll(out), "rm out dir")
	t.Cleanup(
		func() {
			tu.AssertNoError(t, os.RemoveAll(out), "rm out dir")
		},
	)

	err := kube.Import(
		kube.WithImportAppName("tekton"),
		kube.WithImportPackageName("tekton"),
		kube.WithImportOutputDirectory(out),
		kube.WithImportManifestFiles([]string{"testdata/tekton.yaml"}),
		kube.WithImportSerializer(defaultSerializer()),
		kube.WithImportRemoveAppName(true),
		kube.WithImportGroupByKind(true),
		kube.WithImportAddMethods(true),
	)
	tu.AssertNoError(t, err, "failed to import")

	got, err := kubeutil.ListGoFiles(out)
	tu.AssertNoError(t, err, "failed to list go files")
	sort.Strings(got)

	want := []string{
		"out/jamel/tekton/app.go",
		"out/jamel/tekton/cluster-role-binding.go",
		"out/jamel/tekton/cluster-role.go",
		"out/jamel/tekton/config-map.go",
		"out/jamel/tekton/custom-resource-definition.go",
		"out/jamel/tekton/deployment.go",
		"out/jamel/tekton/horizontal-pod-autoscaler.go",
		"out/jamel/tekton/mutating-webhook-configuration.go",
		"out/jamel/tekton/namespace.go",
		"out/jamel/tekton/role-binding.go",
		"out/jamel/tekton/role.go",
		"out/jamel/tekton/secret.go",
		"out/jamel/tekton/service-account.go",
		"out/jamel/tekton/service.go",
		"out/jamel/tekton/validating-webhook-configuration.go",
	}

	tu.IsEqual(t, len(want), len(got))
	tu.AssertEqualSlice(t, want, got)

	src, err := os.ReadFile("out/jamel/tekton/config-map.go")
	tu.AssertNoError(t, err, "reading config-map.go")

	comments := []string{
		"Contains pipelines version which can be queried by external\n\t\t   tools such as CLI. Elevated permissions are already given to\n\t\t   this ConfigMap such that even if we don't have access to\n\t\t   other resources in the namespace we still can have access to\n\t\t   this ConfigMap.",
		"Setting this flag to \"enforce\" will enforce verification of tasks/pipeline. Failing to verify\n\t\t   will fail the taskrun/pipelinerun. \"warn\" will only log the err message and \"skip\"\n\t\t   will skip the whole verification",
		"Setting this flag to \"false\" will stop Tekton from waiting for a\n\t\t   TaskRun's sidecar containers to be running before starting the first\n\t\t   step. This will allow Tasks to be run in environments that don't\n\t\t   support the DownwardAPI volume type, but may lead to unintended\n\t\t   behaviour if sidecars are used.\n\t\t   #\n\t\t   See https://github.com/tektoncd/pipeline/issues/4937 for more info.",
		"Setting this flag to \"true\" enables CloudEvents for CustomRuns and Runs, as long as a\n\t\t   CloudEvents sink is configured in the config-defaults config map",
	}
	set := token.NewFileSet()
	astFile, err := parser.ParseFile(set, "", src, parser.ParseComments)
	tu.AssertNoError(t, err, fmt.Sprintf("parsing %s", src))

	if len(astFile.Comments) < len(comments) {
		t.Errorf("not enough comments: %d", len(astFile.Comments))
	}
	cc := []string{}
	for _, comment := range astFile.Comments {
		cc = append(cc, comment.Text())
	}
	sort.SliceStable(cc, func(i, j int) bool { return cc[i] < cc[j] })
	for i, comment := range comments {
		tu.AssertEqual(t, cleanStr(comment), cleanStr(cc[i]))
	}
}

var clm = strings.NewReplacer(
	"\t", "",
	"\n", "",
)

func cleanStr(m string) string {
	return strings.TrimSpace(clm.Replace(m))
}

func TestJamel_VerboseLogger(t *testing.T) {
	out := filepath.Join(defaultImportOutputDir, "tekton")
	var bufLog, bufOut bytes.Buffer

	err := kube.Import(
		kube.WithImportAppName("tekton"),
		kube.WithImportPackageName("tekton"),
		kube.WithImportOutputDirectory(out),
		kube.WithImportManifestFiles([]string{"testdata/tekton.yaml"}),
		kube.WithImportSerializer(defaultSerializer()),
		kube.WithImportRemoveAppName(true),
		kube.WithImportGroupByKind(true),
		kube.WithImportAddMethods(true),
		kube.WithImportWriter(&bufOut),
		kube.WithImportVerbose(true),
		kube.WithImportIgnoreErrors(true),
		kube.WithImportLogger(kube.Logger(&bufLog)),
	)
	tu.AssertNoError(t, err, "failed to import")

	got := bufLog.String()
	want, err := os.ReadFile("testdata/golden/log.golden")
	tu.AssertNoError(t, err, "reading golden file")
	tu.AssertEqual(t, string(want), got)
}
