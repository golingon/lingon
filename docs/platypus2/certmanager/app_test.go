// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package certmanager

import (
	"os"
	"testing"

	"github.com/golingon/lingon/pkg/kube"
	tu "github.com/golingon/lingon/pkg/testutil"
)

func TestCertManagerExport(t *testing.T) {
	tu.AssertNoError(t, os.RemoveAll("out"))
	tu.AssertNoError(
		t,
		kube.Export(New(), kube.WithExportOutputDirectory("out")),
	)
}
