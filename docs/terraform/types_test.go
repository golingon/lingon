// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terraform

import (
	"github.com/golingon/lingon/pkg/terra"
	aws "github.com/golingon/terraproviders/aws/4.60.0"
)

func Example_typesVars() {
	var (
		S = terra.String
		B = terra.Bool
	)

	_ = aws.NewVpc(
		"vpc", aws.VpcArgs{
			CidrBlock:        S("10.0.0.0/16"),
			EnableDnsSupport: B(true),
		},
	)
}
