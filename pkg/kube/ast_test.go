// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"os"
	"strings"
	"testing"

	"github.com/volvo-cars/lingon/pkg/kubeutil"
	tu "github.com/volvo-cars/lingon/pkg/testutil"
)

func TestKube2GoJen(t *testing.T) {
	type TT struct {
		name     string
		manifest string
		golden   string
		redact   bool
	}
	tests := []TT{
		{
			name:     "deployment",
			manifest: "testdata/golden/deployment.yaml",
			golden:   "testdata/golden/deployment.golden",
		},
		{
			name:     "service",
			manifest: "testdata/golden/service.yaml",
			golden:   "testdata/golden/service.golden",
		},
		{
			name:     "secret",
			manifest: "testdata/golden/secret.yaml",
			golden:   "testdata/golden/secret.golden",
			redact:   true,
		},
		{
			name:     "empty configmap",
			manifest: "testdata/golden/configmap.yaml",
			golden:   "testdata/golden/configmap.golden",
		},
		{
			name:     "configmap with comments",
			manifest: "testdata/golden/cm-comment.yaml",
			golden:   "testdata/golden/cm-comment.golden",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := convert(t, tt.manifest, tt.redact)
				want := tu.ReadGolden(t, tt.golden)
				tu.AssertEqual(t, want, got)
			},
		)
	}
}

func convert(
	t *testing.T,
	path string,
	redact bool,
) string {
	t.Helper()
	data, err := os.ReadFile(path)
	tu.AssertNoError(t, err, "read manifest")
	j := jamel{
		o: importOption{
			Serializer:    defaultSerializer(),
			RedactSecrets: redact,
		},
	}
	m, err := kubeutil.ExtractMetadata(data)
	tu.AssertNoError(t, err, "extract metadata")
	code, err := j.yaml2GoJen(data, m)
	tu.AssertNoError(t, err, "convert yaml to go")
	var b strings.Builder
	err = code.Render(&b)
	tu.AssertNoError(t, err, "render code")
	return b.String()
}
