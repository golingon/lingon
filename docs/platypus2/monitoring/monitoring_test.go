// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package monitoring

import (
	"os"
	"testing"

	"github.com/golingon/lingon/pkg/kube"
	tu "github.com/golingon/lingon/pkg/testutil"
	"github.com/golingon/lingoneks/monitoring/metricsserver"
	"github.com/golingon/lingoneks/monitoring/promcrd"
	"github.com/golingon/lingoneks/monitoring/vmcrd"
	"github.com/golingon/lingoneks/monitoring/vmk8s"
	"github.com/golingon/lingoneks/monitoring/vmop"
)

func TestMonitoringExport(t *testing.T) {
	tests := map[string]kube.Exporter{
		"out/1_promcrd":      promcrd.New(),
		"out/1_vmcrd":        vmcrd.New(),
		"out/2_vmop":         vmop.New(),
		"out/metrics-server": metricsserver.New(),
		"out/vmk8s":          vmk8s.New(),
	}
	for f, km := range tests {
		_ = os.RemoveAll(f)

		tu.AssertNoError(
			t,
			kube.Export(km, kube.WithExportOutputDirectory(f)),
			f,
		)
	}
}
