// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package kubernetes

//go:generate echo "\n>>>> KUBERNETES: generating kubernetes readme\n"
//go:generate go run github.com/dave/rebecca/cmd/becca@latest  -package=github.com/golingon/docskubernetes/crd -input readme.md.tpl
//go:generate go run github.com/dave/rebecca/cmd/becca@latest  -package=github.com/golingon/docskubernetes/crd -input=optionsimport.md.tpl -output=options-import.md
//go:generate go run github.com/dave/rebecca/cmd/becca@latest  -package=github.com/golingon/docskubernetes/crd -input=optionsexport.md.tpl -output=options-export.md
