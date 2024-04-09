// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terrajen

import (
	"testing"

	tu "github.com/golingon/lingon/pkg/testutil"
)

func TestProviderShortName(t *testing.T) {
	expectedName := "some_resource"
	trimmedName := providerShortName("aws_" + expectedName)
	tu.AssertEqual(t, expectedName, trimmedName)
}
