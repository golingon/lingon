// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"os"
	"testing"
)

func ReadGolden(t *testing.T, path string) string {
	t.Helper()
	golden, err := os.ReadFile(path)
	AssertNoError(t, err, "read golden file")
	return string(golden)
}
