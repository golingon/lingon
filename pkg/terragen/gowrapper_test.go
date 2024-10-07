// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terragen

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/golingon/lingon/pkg/internal/terrajen"
	tu "github.com/golingon/lingon/pkg/testutil"
	tfjson "github.com/hashicorp/terraform-json"
	"golang.org/x/tools/txtar"
)

var goldenTestDir = filepath.Join("testdata", "golden")

type ProviderTestCase struct {
	Name string

	ProviderName    string
	ProviderSource  string
	ProviderVersion string

	FilterResources   []string
	FilterDataSources []string
}

var providerTests = []ProviderTestCase{
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
	{
		Name:            "aws_securitylake_subscriber",
		ProviderName:    "aws",
		ProviderSource:  "hashicorp/aws",
		ProviderVersion: "5.44.0",

		FilterResources:   []string{"aws_securitylake_subscriber"},
		FilterDataSources: []string{"aws_securitylake_subscriber"},
	},
	{
		Name:            "aws_globalaccelerator_cross_account_attachment",
		ProviderName:    "aws",
		ProviderSource:  "hashicorp/aws",
		ProviderVersion: "5.47.0",

		FilterResources: []string{
			"aws_globalaccelerator_cross_account_attachment",
		},
		FilterDataSources: []string{
			"aws_globalaccelerator_cross_account_attachment",
		},
	},
	{
		Name:            "cidaas",
		ProviderName:    "cidaas",
		ProviderSource:  "cidaas/cidaas",
		ProviderVersion: "3.1.2",

		FilterResources: []string{
			"cidaas_app",
			"cidaas_custom_provider",
			"cidaas_registration_field",
		},
		FilterDataSources: []string{},
	},
}

func TestMain(m *testing.M) {
	update := flag.Bool("update", false, "update golden files")
	flag.Parse()
	if *update {
		slog.Info("updating golden files and skipping tests")
		if err := generatedGoldenFiles(); err != nil {
			slog.Error("generating golden files", "error", err)
			os.Exit(1)
		}
		// Skip running tests if updating golden files.
		return
	}
	os.Exit(m.Run())
}

func generatedGoldenFiles() error {
	if err := os.RemoveAll(goldenTestDir); err != nil {
		return fmt.Errorf("removing golden test dir: %w", err)
	}

	for _, test := range providerTests {
		ctx := context.Background()
		ps, err := GenerateProviderSchema(
			ctx, Provider{
				Name:    test.ProviderName,
				Source:  test.ProviderSource,
				Version: test.ProviderVersion,
			},
		)
		if err != nil {
			return fmt.Errorf("generating provider schema: %w", err)
		}

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
			GeneratedPackageLocation: "out",
			ProviderName:             test.ProviderName,
			ProviderSource:           test.ProviderSource,
			ProviderVersion:          test.ProviderVersion,
		}

		ar, err := generateProviderTxtar(providerGenerator, ps)
		if err != nil {
			return fmt.Errorf("generating provider txtar: %w", err)
		}

		testDir := filepath.Join(goldenTestDir, test.Name)
		if err := os.MkdirAll(testDir, os.ModePerm); err != nil {
			return fmt.Errorf("creating golden test dir: %w", err)
		}

		schemaPath := filepath.Join(testDir, "schema.json")
		schemaFile, err := os.Create(schemaPath)
		if err != nil {
			return fmt.Errorf("creating schema file: %w", err)
		}
		if err := json.NewEncoder(schemaFile).Encode(ps); err != nil {
			return fmt.Errorf("writing schema file: %w", err)
		}

		txtarPath := filepath.Join(testDir, "provider.txtar")
		if err := os.WriteFile(txtarPath, txtar.Format(ar), 0o644); err != nil {
			return fmt.Errorf("writing txtar file: %w", err)
		}
	}
	return nil
}

// TestGenerateProvider tests that the generated Go code for a Terraform
// provider matches the golden tests.
// This is to ensure that the generated code is consistent and does not change,
// unless we want it to.
//
// It is quite challenging to verify the correctness of the generated code,
// and it is out of scope of this pkg and test.
func TestGenerateProvider(t *testing.T) {
	goldenTestDir := filepath.Join("testdata", "golden")

	for _, test := range providerTests {
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
				GeneratedPackageLocation: "out",
				ProviderName:             test.ProviderName,
				ProviderSource:           test.ProviderSource,
				ProviderVersion:          test.ProviderVersion,
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

var (
	// gotype is for linting code
	gotypeCmd     = "golang.org/x/tools/cmd/gotype"
	gotypeVersion = "@v0.25.0"
	gotype        = gotypeCmd + gotypeVersion
)

// TestCompileGenGoCode tests that the generated Go code compiles.
// This test is slow because compiles the generated Go code using `gotype`.
// https://pkg.go.dev/golang.org/x/tools/cmd/gotype
//
// This test has a dependency on `gotype`. It uses `go run ...` to execute
// `gotype`.
func TestCompileGenGoCode(t *testing.T) {
	ctx := tu.WithTimeout(t, context.Background(), time.Minute*10)
	for _, test := range providerTests {
		t.Run(test.Name, func(t *testing.T) {
			outDir := filepath.Join("out", test.Name)
			schemaPath := filepath.Join(goldenTestDir, test.Name, "schema.json")
			schemaFile, err := os.Open(schemaPath)
			tu.AssertNoError(t, err, "opening schema file")
			var providerSchema tfjson.ProviderSchema
			err = json.NewDecoder(schemaFile).Decode(&providerSchema)
			tu.AssertNoError(t, err, "decoding schema file")

			if err := GenerateGoCode(
				GenerateGoArgs{
					ProviderName:    test.ProviderName,
					ProviderSource:  test.ProviderSource,
					ProviderVersion: test.ProviderVersion,
					OutDir:          outDir,
					Force:           false,
					Clean:           true,
				},
				&providerSchema,
			); err != nil {
				tu.AssertNoError(t, err)
			}
			dirs := dirsForProvider(outDir, &providerSchema)
			for _, dir := range dirs {
				t.Logf("running gotype on %s", dir)
				goTypeExec(t, ctx, dir)
			}
		})
	}
}

// goTypeExec executes `gotype` in the given directory.
func goTypeExec(t *testing.T, ctx context.Context, dir string) {
	cmd := exec.CommandContext(ctx, "go", "run", gotype, "-v", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("gotype output:\n%s", string(out))
		tu.AssertNoError(t, err)
	}
}

// dirsForProvider returns a list of directories that contain generated Go code
// based on a terraform provider.
func dirsForProvider(root string, schema *tfjson.ProviderSchema) []string {
	dirsMap := map[string]struct{}{
		// Include root directory because it contains the provider.go file.
		root: {},
	}
	for key := range schema.ResourceSchemas {
		dirsMap[filepath.Join(root, key)] = struct{}{}
	}
	for key := range schema.DataSourceSchemas {
		dirsMap[filepath.Join(root, key)] = struct{}{}
	}
	dirs := make([]string, 0, len(dirsMap))
	for dir := range dirsMap {
		dirs = append(dirs, dir)
	}
	return dirs
}
