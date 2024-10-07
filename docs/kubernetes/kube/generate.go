// Copyright (c) Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package kube

//go:generate rm -rf out/
//go:generate go run github.com/dave/rebecca/cmd/becca@latest  -package=github.com/golingon/docskubernetes/kube -input=readme.md.tpl
//go:generate go run github.com/golingon/lingon/cmd/explode -in ../../../pkg/kube/testdata/tekton.yaml -out out/explode
//go:generate go run github.com/golingon/lingon/cmd/kygo -in ../../../pkg/kube/testdata/tekton.yaml -out out/tekton -app=tekton -group -clean-name
//go:generate go test -v .
