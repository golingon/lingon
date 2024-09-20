// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terraform

//go:generate echo "\n>>>> docs/terraform: generating hashicorp/aws terra provider\n"
//go:generate go run -mod=readonly github.com/golingon/lingon/cmd/terragen -out ./out/aws -clean -provider aws=hashicorp/aws:5.44.0

//go:generate echo "\n>>>> generating terraform readme\n"
//go:generate go run github.com/dave/rebecca/cmd/becca@latest  -package=github.com/golingon/docsterraform -input readme.md.tpl
