// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package crd

//go:generate rm -rf out/
//go:generate go run github.com/dave/rebecca/cmd/becca@latest  -package=github.com/golingon/docskubernetes/crd -input readme.md.tpl
//go:generate go test -v example_import_test.go
//go:generate go test -v import_options_test.go
//go:generate go test -v example_export_test.go
//go:generate rm -rf out/manifest
