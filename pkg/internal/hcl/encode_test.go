// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package hcl

import (
	"bytes"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	tu "github.com/golingon/lingon/pkg/testutil"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type TerraformBlock struct {
	Backend           BackendBlock            `hcl:"backend,block"`
	RequiredProviders RequiredProvidersBlocks `hcl:"required_providers,block"`
}

type RequiredProvidersBlocks struct {
	Providers map[string]RequiredProvider `hcl:",remain"`
}

type RequiredProvider struct {
	Source  string `hcl:"source,attr" cty:"source"`
	Version string `hcl:"version,attr" cty:"version"`
}

type BackendBlock struct {
	Type string `hcl:",label"`

	CommonBlockConfig `hcl:",remain"`
}

type ProviderBlock struct {
	LocalName string `hcl:",label"`

	CommonBlockConfig `hcl:",remain"`
}

type DataSourceBlock struct {
	DataSource string `hcl:",label"`
	LocalName  string `hcl:",label"`

	CommonBlockConfig `hcl:",remain"`
}

type ResourceBlock struct {
	Type      string `hcl:",label"`
	LocalName string `hcl:",label"`

	CommonBlockConfig `hcl:",remain"`
}

type CommonBlockConfig struct {
	StringField string `hcl:"string_field,attr"`
	NumberField int    `hcl:"number_field,attr"`
}

// HCLFile is the test structure which we will use to encode HCL and also decode
// the resulting HCL
// into
type HCLFile struct {
	Terraform   TerraformBlock    `hcl:"terraform,block"`
	Providers   []ProviderBlock   `hcl:"provider,block"`
	DataSources []DataSourceBlock `hcl:"data,block"`
	Resources   []ResourceBlock   `hcl:"resource,block"`
}

// Create a common config which we reuse. We might want to make this a bit more
// diverse
// or create separate instance of this for greater coverage.
var cbcfg = CommonBlockConfig{
	StringField: "test",
	NumberField: 123,
}

// TestEncode tests the encoder.
// The approach is to
// 1. Create an expected HCL value
// 2. Encode HCL based on the expected value
// 3. Use the hcl library to decode the encoded HCL into a new instance of the
// expected value
// 4. Assert that they match
//
// This is a bit complex but much more convincing and easier to maintain in the
// long run than string-based comparisons.
func TestEncode(t *testing.T) {
	// Create the expected HCL values which we will use to create the encoder
	// arguments
	expectedHCL := HCLFile{
		Terraform: TerraformBlock{
			Backend: BackendBlock{
				Type:              "dummy",
				CommonBlockConfig: cbcfg,
			},
			RequiredProviders: RequiredProvidersBlocks{
				Providers: map[string]RequiredProvider{
					"aws": {
						Source:  "hashicorp/aws",
						Version: "4.49.0",
					},
				},
			},
		},
		Providers: []ProviderBlock{
			{
				LocalName:         "aws",
				CommonBlockConfig: cbcfg,
			},
		},
		DataSources: []DataSourceBlock{
			{
				DataSource:        "test_data_source",
				LocalName:         "test",
				CommonBlockConfig: cbcfg,
			},
		},
		Resources: []ResourceBlock{
			{
				Type:              "test_resource",
				LocalName:         "test",
				CommonBlockConfig: cbcfg,
			},
		},
	}

	// Populate the encoder arguments using the expected HCL value
	args := EncodeArgs{
		Backend: &Backend{
			Type:          expectedHCL.Terraform.Backend.Type,
			Configuration: expectedHCL.Terraform.Backend.CommonBlockConfig,
		},
	}
	for name, prov := range expectedHCL.Terraform.RequiredProviders.Providers {
		args.Providers = append(
			args.Providers, Provider{
				LocalName:     name,
				Source:        prov.Source,
				Version:       prov.Version,
				Configuration: cbcfg,
			},
		)
	}
	for _, data := range expectedHCL.DataSources {
		args.DataSources = append(
			args.DataSources, DataSource{
				DataSource:    data.DataSource,
				LocalName:     data.LocalName,
				Configuration: data.CommonBlockConfig,
			},
		)
	}
	for _, res := range expectedHCL.Resources {
		args.Resources = append(
			args.Resources, Resource{
				Type:          res.Type,
				LocalName:     res.LocalName,
				Configuration: res.CommonBlockConfig,
			},
		)
	}
	// Run the encoder
	var b bytes.Buffer
	err := Encode(&b, args)
	tu.AssertNoError(t, err, "Encode failed")

	// Decode the encoded HCL into a new instance of our test structure and
	// compare
	actualHCL := HCLFile{}
	err = hclsimple.Decode("test.hcl", b.Bytes(), nil, &actualHCL)
	tu.AssertNoError(t, err, "Decode failed")
	if diff := tu.Diff(actualHCL, expectedHCL); diff != "" {
		t.Error(tu.Callers(), diff)
	}
}

func TestEncodeRaw(t *testing.T) {
	expectedHCL := HCLFile{
		Terraform: TerraformBlock{
			Backend: BackendBlock{
				Type:              "sometype",
				CommonBlockConfig: cbcfg,
			},
			RequiredProviders: RequiredProvidersBlocks{
				Providers: map[string]RequiredProvider{
					"localprovider": {
						Source:  "somesource",
						Version: "someversion",
					},
					"another": {
						Source:  "anothersource",
						Version: "anotherversion",
					},
				},
			},
		},
		Providers: []ProviderBlock{
			{
				LocalName:         "localname",
				CommonBlockConfig: cbcfg,
			},
		},
		DataSources: []DataSourceBlock{
			{
				DataSource:        "some_data_source",
				LocalName:         "localname",
				CommonBlockConfig: cbcfg,
			},
		},
		Resources: []ResourceBlock{
			{
				Type:              "some_resource_type",
				LocalName:         "this",
				CommonBlockConfig: cbcfg,
			},
		},
	}
	assertEncodeRawAndDecode(t, expectedHCL)
}

func TestEncode_StructWithNil(t *testing.T) {
	type StructWithNil struct {
		ShouldBeNil *string `hcl:"should_be_nil"`
	}
	expectedHCL := StructWithNil{}
	assertEncodeRawAndDecode(t, expectedHCL)
	// Make explicit check that there are no blocks.
	block := hclwrite.NewBlock("block", nil)
	err := encodeStruct(reflect.ValueOf(expectedHCL), block,
		block.Body())
	tu.AssertNoError(t, err)

	tu.AssertEqual(t, len(block.Body().Attributes()), 0)
}

type StructWithChildAttribute struct {
	Child    *ChildStruct  `hcl:"child,attr"`
	Children []ChildStruct `hcl:"children,attr"`
}

type ChildStruct struct {
	ChildField string `hcl:"child_field,attr" cty:"child_field"`
}

func TestEncode_StructWithChildAttribute(t *testing.T) {
	expectedHCL := StructWithChildAttribute{
		Child: &ChildStruct{
			ChildField: "child_field_value",
		},
		Children: []ChildStruct{
			{
				ChildField: "child_field_value_1",
			},
			{
				ChildField: "child_field_value_2",
			},
		},
	}
	assertEncodeRawAndDecode(t, expectedHCL)
}

func assertEncodeRawAndDecode[T any](t *testing.T, in T) {
	var b bytes.Buffer
	err := EncodeRaw(&b, in)
	tu.AssertNoError(t, err, "EncodeRaw failed")

	var out T
	err = hclsimple.Decode("test.hcl", b.Bytes(), nil, &out)
	tu.AssertNoError(t, err, "Decode failed")

	if diff := tu.Diff(out, in); diff != "" {
		t.Error(tu.Callers(), diff)
	}
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

func TestEncodeRawGolden(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			err := EncodeRaw(&b, tt.value())
			tu.AssertNoError(t, err, "EncodeRaw failed")

			goldenFile := filepath.Join(goldenTestDir, tt.name+".hcl")
			golden, err := os.ReadFile(goldenFile)
			tu.AssertNoError(t, err, "reading golden file failed")

			tu.AssertEqual(t, b.String(), string(golden))
		})
	}
}

type test struct {
	name  string
	value func() interface{}
}

var tests = []test{
	{
		name: "empty_values",
		value: func() interface{} {
			type (
				Whatever    struct{}
				EmptyStruct struct {
					EmptySlice    []string          `hcl:"empty_slice,attr"`
					EmptyArray    [3]string         `hcl:"empty_array,attr"`
					EmptyMap      map[string]string `hcl:"empty_map,attr"`
					EmptyWhatever []Whatever        `hcl:"empty_whatever,attr"`
				}
			)
			return EmptyStruct{
				EmptySlice:    []string{},
				EmptyArray:    [3]string{},
				EmptyMap:      map[string]string{},
				EmptyWhatever: []Whatever{},
			}
		},
	},
	{
		name: "nil_values",
		value: func() interface{} {
			type (
				Whatever  struct{}
				NilStruct struct {
					NilSlice []Whatever          `hcl:"nil_slice,attr"`
					NilArray [3]Whatever         `hcl:"nil_array,attr"`
					NilMap   map[string]Whatever `hcl:"nil_map,attr"`
				}
			)
			return NilStruct{}
		},
	},
}

var goldenTestDir = filepath.Join("testdata", "golden")

func generatedGoldenFiles() error {
	if err := os.MkdirAll(goldenTestDir, 0o755); err != nil {
		return fmt.Errorf("creating golden directory: %w", err)
	}
	for _, tt := range tests {
		tt := tt
		var b bytes.Buffer
		err := EncodeRaw(&b, tt.value())
		if err != nil {
			return fmt.Errorf("encoding %q: %w", tt.name, err)
		}
		testFile := filepath.Join(goldenTestDir, tt.name+".hcl")
		if err := os.WriteFile(testFile, b.Bytes(), 0o644); err != nil {
			return fmt.Errorf("writing golden file %q: %w", testFile, err)
		}
	}
	return nil
}
