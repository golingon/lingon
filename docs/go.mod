module github.com/golingon/lingon/docs

go 1.21

replace github.com/golingon/lingon => ../

require (
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/aws/karpenter v0.29.2
	github.com/aws/karpenter-core v0.29.2
	github.com/eidolon/wordwrap v0.0.0-20161011182207-e0f54129b8bb
	github.com/fatih/color v1.15.0
	github.com/go-playground/validator/v10 v10.15.1
	github.com/golingon/terraproviders/aws/4.60.0 v0.0.0-20230821062044-8717820dd713
	github.com/golingon/terraproviders/tls/4.0.4 v0.0.0-20230821062044-8717820dd713
	github.com/google/go-containerregistry v0.16.1
	github.com/hashicorp/hcl/v2 v2.17.0
	github.com/hashicorp/terraform-exec v0.18.1
	github.com/hashicorp/terraform-json v0.17.1
	github.com/hexops/valast v1.4.4
	github.com/invopop/yaml v0.2.0
	github.com/rogpeppe/go-internal v1.11.0
	github.com/stretchr/testify v1.8.4
	github.com/golingon/lingon v0.0.0-20230821085204-81fbacc93e96
	github.com/zegl/kube-score v1.17.0
	istio.io/api v1.19.0-beta.1
	istio.io/client-go v1.18.1
	k8s.io/api v0.28.0
	k8s.io/apiextensions-apiserver v0.28.0
	k8s.io/apimachinery v0.28.0
	k8s.io/client-go v0.28.0
	sigs.k8s.io/secrets-store-csi-driver v1.3.4
)

require (
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/aws/aws-sdk-go v1.44.329 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.14.3 // indirect
	github.com/dave/jennifer v1.7.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/cli v24.0.5+incompatible // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v24.0.5+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/samber/lo v1.38.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tidwall/gjson v1.16.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/vbatts/tar-split v0.11.5 // indirect
	github.com/veggiemonk/strcase v0.0.0-20230627213939-a882c834bcab // indirect
	github.com/zclconf/go-cty v1.13.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	golang.org/x/tools v0.12.1-0.20230815132531-74c255bcf846 // indirect
	google.golang.org/genproto v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b // indirect
	knative.dev/pkg v0.0.0-20230821102121-81e4ee140363 // indirect
	mvdan.cc/gofumpt v0.5.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.3.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
