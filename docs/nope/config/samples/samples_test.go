// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package samples_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golingon/lingon/pkg/kube"
	"github.com/golingon/lingon/pkg/kubeutil"
	tu "github.com/golingon/lingon/pkg/testutil"
	"github.com/volvo-cars/nope/config/samples"
)

func TestExport(t *testing.T) {
	outDir := "out"

	rmErr := os.RemoveAll(outDir)
	tu.AssertNoError(t, rmErr, "removing out directory")

	app := samples.NewApp()
	err := kube.Export(
		app,
		kube.WithExportOutputDirectory(outDir),
		kube.WithExportNameFileFunc(func(m *kubeutil.Metadata) string {
			return strings.ToLower(
				fmt.Sprintf("%s_%s.yaml", m.Kind, m.Meta.Name),
			)
		}),
	)
	tu.AssertNoError(t, err, "exporting samples")
}
