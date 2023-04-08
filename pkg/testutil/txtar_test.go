package testutil_test

import (
	"sort"
	"testing"

	tu "github.com/volvo-cars/lingon/pkg/testutil"
)

func TestFolder2Txtar(t *testing.T) {
	ar, err := tu.Folder2Txtar("../kube/testdata")
	tu.AssertNoError(t, err)
	tu.IsNotEqual(t, 0, len(ar.Files))
	want := []string{
		"../kube/testdata/argocd.yaml",
		"../kube/testdata/cilium.yaml",
		"../kube/testdata/external-secrets.yaml",
		"../kube/testdata/go/tekton/app.go",
		"../kube/testdata/go/tekton/cluster-role-binding.go",
		"../kube/testdata/go/tekton/cluster-role.go",
		"../kube/testdata/go/tekton/config-map.go",
		"../kube/testdata/go/tekton/custom-resource-definition.go",
		"../kube/testdata/go/tekton/deployment.go",
		"../kube/testdata/go/tekton/horizontal-pod-autoscaler.go",
		"../kube/testdata/go/tekton/mutating-webhook-configuration.go",
		"../kube/testdata/go/tekton/namespace.go",
		"../kube/testdata/go/tekton/role-binding.go",
		"../kube/testdata/go/tekton/role.go",
		"../kube/testdata/go/tekton/secret.go",
		"../kube/testdata/go/tekton/service-account.go",
		"../kube/testdata/go/tekton/service.go",
		"../kube/testdata/go/tekton/validating-webhook-configuration.go",
		"../kube/testdata/golden/cm-comment.golden",
		"../kube/testdata/golden/cm-comment.yaml",
		"../kube/testdata/golden/configmap.golden",
		"../kube/testdata/golden/configmap.yaml",
		"../kube/testdata/golden/deployment.golden",
		"../kube/testdata/golden/deployment.yaml",
		"../kube/testdata/golden/empty.golden",
		"../kube/testdata/golden/empty.yaml",
		"../kube/testdata/golden/log.golden",
		"../kube/testdata/golden/reader.yaml",
		"../kube/testdata/golden/secret.golden",
		"../kube/testdata/golden/secret.yaml",
		"../kube/testdata/golden/service.golden",
		"../kube/testdata/golden/service.yaml",
		"../kube/testdata/grafana.yaml",
		"../kube/testdata/istio.yaml",
		"../kube/testdata/karpenter.yaml",
		"../kube/testdata/spark.yaml",
		"../kube/testdata/tekton-updated.yaml",
		"../kube/testdata/tekton.yaml",
	}
	filenames := make([]string, 0, len(ar.Files))
	for _, f := range ar.Files {
		filenames = append(filenames, f.Name)
	}
	sort.Strings(want)
	tu.AssertEqualSlice(t, want, filenames)
}
