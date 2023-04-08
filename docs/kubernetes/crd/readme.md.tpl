# Example CRD

All can be done in one step:

```bash
go generate -v ./...
```

## Import the manifest in Go

{{ "Example_import" | example }}


## Export the Go struct to YAML


{{ "Example_export" | example }}


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