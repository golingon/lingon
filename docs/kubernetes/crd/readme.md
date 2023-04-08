# Example CRD

All can be done in one step:

```bash
go generate -v ./...
```

## Import the manifest in Go

```go
package crd_test

import (
	"fmt"
	"os"

	"github.com/volvo-cars/lingon/pkg/kube"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	secretsstorev1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
)

func defaultSerializer() runtime.Decoder {
	// Add CRDs to scheme
	// This is needed to be able to import CRDs from kubernetes manifests files.
	_ = apiextensions.AddToScheme(scheme.Scheme)
	_ = apiextensionsv1.AddToScheme(scheme.Scheme)
	_ = secretsstorev1.AddToScheme(scheme.Scheme)
	_ = istionetworkingv1beta1.AddToScheme(scheme.Scheme)
	return scheme.Codecs.UniversalDeserializer()
}

// Example_import to shows how to import CRDs from kubernetes manifests files.
func Example_import() {
	// Remove previously generated output directory
	_ = os.RemoveAll("./out")

	if err := kube.Import(
		kube.WithImportAppName("team"),
		kube.WithImportPackageName("team"),
		kube.WithImportManifestFiles([]string{"./manifest.yaml"}),
		kube.WithImportOutputDirectory("./out"),
		kube.WithImportSerializer(defaultSerializer()),
		// do not print verbose information
		kube.WithImportVerbose(false),
		// do not ignore errors
		kube.WithImportIgnoreErrors(false),
		// use the default logger even if unused with WithImportVerbose(false)
		kube.WithImportLogger(kube.Logger(os.Stderr)),
	); err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Println("successfully imported CRDs from manifest.yaml")

	// Output:
	// successfully imported CRDs from manifest.yaml
}
```


## Export the Go struct to YAML


```go
package crd_test

import (
	"fmt"
	"os"
	"path/filepath"

	team "github.com/volvo-cars/lingon/docs/kubernetes/crd/out"
	"github.com/volvo-cars/lingon/pkg/kube"
)

var defaultOut = "out"

// Example_export to shows how to export to kubernetes manifests files in YAML.
func Example_export() {
	out := filepath.Join(defaultOut, "manifests")

	// Remove previously generated output directory
	_ = os.RemoveAll(out)

	tm := team.New()
	if err := kube.Export(tm, kube.WithExportOutputDirectory(out)); err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Println("successfully exported CRDs to manifests")

	// Output:
	// successfully exported CRDs to manifests
}
```


## How to find CRD types

This is more of a heuristic than a rule.

- go to the repo of the project
- search for `AddToScheme` function

Often it is in the `pkg/apis` or `apis` folder of the project.

## List of well-known CRDs types

> This is best-effort

```go
import(
    certmanager "github.com/cert-manager/cert-manager/pkg/api"
    kservev1alpha1 "github.com/kserve/kserve/pkg/apis/serving/v1alpha1"
    servingv1alpha1 "github.com/kserve/modelmesh-serving/apis/serving/v1alpha1"
    profilev1 "github.com/kubeflow/kubeflow/components/profile-controller/api/v1"
    profilev1beta1 "github.com/kubeflow/kubeflow/components/profile-controller/api/v1beta1"
    metacontrolleralpha "github.com/metacontroller/metacontroller/pkg/apis/metacontroller/v1alpha1"
    tektonpipelinesv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
    tektontriggersv1alpha1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
    istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
    istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
    istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
    "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
    apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
    "k8s.io/apimachinery/pkg/runtime"
    utilruntime "k8s.io/apimachinery/pkg/util/runtime"
    knativecachingalpha1 "knative.dev/caching/pkg/apis/caching/v1alpha1"
    secretsstorev1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
)
```