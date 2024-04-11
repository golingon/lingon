// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package main

//go:generate echo "\n>>>> docs/platypus2: generating hashicorp/aws terra provider\n"
//go:generate go run -mod=readonly github.com/golingon/lingon/cmd/terragen -out ./out/aws -pkg github.com/golingon/lingon/docs/terraform/out/aws -clean -provider aws=hashicorp/aws:5.44.0

//go:generate echo "\n>>>> docs/platypus2: generating hashicorp/tls terra provider\n"
//go:generate go run -mod=readonly github.com/golingon/lingon/cmd/terragen -out ./out/tls -pkg github.com/golingon/lingon/docs/terraform/out/tls -clean -provider tls=hashicorp/tls:4.0.5
