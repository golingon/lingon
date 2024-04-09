// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package karpenter

import (
	"os"
	"testing"

	"github.com/golingon/lingon/pkg/kube"
	"github.com/stretchr/testify/require"
)

func TestExport(t *testing.T) {
	_ = os.RemoveAll("out")

	app := New(
		Opts{
			ClusterName:            "CLUSTER_NAME",
			ClusterEndpoint:        "CLUSTER_ENDPOINT",
			IAMRoleArn:             "ROLE_ARN",
			DefaultInstanceProfile: "DEFAULT_INSTANCE_PROFILE",
			InterruptQueue:         "INTERRUPT_QUEUE",
		},
	)
	err := kube.Export(app, kube.WithExportOutputDirectory("out"))
	require.NoError(t, err)
}
