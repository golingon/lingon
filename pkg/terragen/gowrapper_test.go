// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terragen

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/golingon/lingon/pkg/internal/terrajen"
	tu "github.com/golingon/lingon/pkg/testutil"
	tfjson "github.com/hashicorp/terraform-json"
	"golang.org/x/tools/txtar"
)

var update = flag.Bool("update", false, "update golden files")

type ProviderTestCase struct {
	Name string

	ProviderName    string
	ProviderSource  string
	ProviderVersion string

	FilterResources   []string
	FilterDataSources []string
}

// TestGenerateProvider tests the generation of Terraform provider schemas into
// Go code.
//
// When using the -update flag, it updates the golden files which are used as a
// baseline for detecting drift in the generated code.
// It is quite challenging to verify the correctness of the generated code,
// and it is out of scope of this pkg and test.
func TestGenerateProvider(t *testing.T) {
	goldenTestDir := filepath.Join("testdata", "golden")

	tests := []ProviderTestCase{
		{
			Name:            "aws_emr_cluster",
			ProviderName:    "aws",
			ProviderSource:  "hashicorp/aws",
			ProviderVersion: "5.44.0",

			FilterResources:   []string{"aws_emr_cluster"},
			FilterDataSources: []string{"aws_emr_cluster"},
		},
		{
			Name:            "aws_iam_role",
			ProviderName:    "aws",
			ProviderSource:  "hashicorp/aws",
			ProviderVersion: "5.44.0",

			FilterResources:   []string{"aws_iam_role"},
			FilterDataSources: []string{"aws_iam_role"},
		},
	}
	if *update {
		t.Log("running update")
		// Generate "golden" files. Start be deleting the directory.
		err := os.RemoveAll(goldenTestDir)
		tu.AssertNoError(t, err, "removing golden test dir")

		for _, test := range tests {
			ctx := context.Background()
			ps, err := GenerateProviderSchema(
				ctx, Provider{
					Name:    test.ProviderName,
					Source:  test.ProviderSource,
					Version: test.ProviderVersion,
				},
			)
			tu.AssertNoError(t, err, "generating provider schema:", test.Name)

			// Filter resources and data sources.
			for rName := range ps.ResourceSchemas {
				if !slices.Contains(test.FilterResources, rName) {
					delete(ps.ResourceSchemas, rName)
				}
			}
			for dName := range ps.DataSourceSchemas {
				if !slices.Contains(test.FilterDataSources, dName) {
					delete(ps.DataSourceSchemas, dName)
				}
			}

			providerGenerator := terrajen.ProviderGenerator{
				GoProviderPkgPath:        "test/out",
				GeneratedPackageLocation: "out",
				ProviderName:             test.ProviderName,
				ProviderSource:           test.ProviderSource,
				ProviderVersion:          test.ProviderVersion,
			}

			ar, err := generateProviderTxtar(providerGenerator, ps)
			tu.AssertNoError(t, err, "generating provider txtar")

			testDir := filepath.Join(goldenTestDir, test.Name)
			err = os.MkdirAll(testDir, os.ModePerm)
			tu.AssertNoError(t, err, "creating golden test dir: ", testDir)

			schemaPath := filepath.Join(testDir, "schema.json")
			schemaFile, err := os.Create(schemaPath)
			tu.AssertNoError(t, err, "creating schema file")
			err = json.NewEncoder(schemaFile).Encode(ps)
			tu.AssertNoError(t, err, "writing schema file")

			txtarPath := filepath.Join(testDir, "provider.txtar")
			err = os.WriteFile(txtarPath, txtar.Format(ar), 0o644)
			tu.AssertNoError(t, err, "writing txtar file")
		}

		t.SkipNow()
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			schemaPath := filepath.Join(goldenTestDir, test.Name, "schema.json")
			schemaFile, err := os.Open(schemaPath)
			tu.AssertNoError(t, err, "opening schema file")
			var ps tfjson.ProviderSchema
			err = json.NewDecoder(schemaFile).Decode(&ps)
			tu.AssertNoError(t, err, "decoding schema file")

			txtarPath := filepath.Join(
				goldenTestDir,
				test.Name,
				"provider.txtar",
			)
			txtarContents, err := os.ReadFile(txtarPath)
			tu.AssertNoError(t, err, "opening txtar file")
			expectedAr := txtar.Parse(txtarContents)

			providerGenerator := terrajen.ProviderGenerator{
				GoProviderPkgPath:        "test/out",
				GeneratedPackageLocation: "out",
				ProviderName:             "aws",
				ProviderSource:           "hashicorp/aws",
				ProviderVersion:          "5.44.0",
			}

			actualAr, err := generateProviderTxtar(providerGenerator, &ps)
			tu.AssertNoError(t, err, "generating provider txtar")

			if diff := tu.DiffTxtar(actualAr, expectedAr); diff != "" {
				t.Fatal(tu.Callers(), diff)
			}
		})
	}
}

func TestParseProvider(t *testing.T) {
	type test struct {
		providerStr string
		provider    Provider
		expectErr   bool
		errmsg      string
	}

	tests := []test{
		{
			providerStr: "aws=hashicorp/aws:4.60.0",
			provider: Provider{
				Name:    "aws",
				Source:  "hashicorp/aws",
				Version: "4.60.0",
			},
			expectErr: false,
		},
		{
			providerStr: "awshashicorp/aws:4.60.0",
			expectErr:   true,
			errmsg:      "provider format incorrect: missing `=`",
		},
		{
			providerStr: "aws=hashicorp/aws",
			expectErr:   true,
			errmsg:      "provider format incorrect: missing `:` in `source:version`",
		},
		{
			providerStr: "aws=hashicorp/aws",
			expectErr:   true,
			errmsg:      "provider format incorrect: missing `:` in `source:version`",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.providerStr, func(t *testing.T) {
				ap, err := ParseProvider(tt.providerStr)
				if tt.expectErr {
					tu.AssertErrorMsg(t, err, tt.errmsg)
				} else {
					tu.AssertNoError(t, err, "parsing provider")
					tu.AssertEqual(t, tt.provider, ap)
				}
			},
		)
	}
}
