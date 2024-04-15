// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terraform

import (
	"github.com/golingon/lingon/docs/terraform/out/aws/aws_vpc"
	"github.com/golingon/lingon/pkg/terra"
)

func Example_typesVars() {
	var (
		S = terra.String
		B = terra.Bool
	)

	_ = aws_vpc.New(
		"vpc", aws_vpc.Args{
			CidrBlock:        S("10.0.0.0/16"),
			EnableDnsSupport: B(true),
		},
	)
}
