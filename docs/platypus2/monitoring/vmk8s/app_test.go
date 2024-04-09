// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package vmk8s

import (
	"os"
	"testing"

	"github.com/golingon/lingon/pkg/kube"
	tu "github.com/golingon/lingon/pkg/testutil"
)

func TestApp(t *testing.T) {
	os.RemoveAll("out")

	app := New()
	tu.AssertNoError(
		t,
		kube.Export(
			app,
			kube.WithExportOutputDirectory("out"),
			kube.WithExportExplodeManifests(true),
		),
		"export",
	)
}
