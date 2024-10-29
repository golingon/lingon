// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terrajen

import (
	"fmt"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/veggiemonk/strcase"

	tfjson "github.com/hashicorp/terraform-json"
)

// ProviderGenerator is created for each provider and is used to generate the
// schema
// for each resource and data object, and the provider configuration.
// The schemas are used by the generator to create the Go files and sub
// packages.
type ProviderGenerator struct {
	// GeneratedPackageLocation is the directory on the filesystem where the
	// generated Go files will be created.
	// The GoProviderPkgPath path must match the location of the generated files
	// so that they can be imported correctly.
	// E.g. if we are in a Go module called "my-module" and we generate the
	// files in a "gen" directory within the root of "my-module", then
	// GoProviderPkgPath is "my-module/gen" and the GeneratedPackageLocation is
	// "./gen" assuming we are running from the root of
	// "my-module"
	GeneratedPackageLocation string
	// ProviderName is the local name of the provider.
	// E.g. aws
	// https://developer.hashicorp.com/terraform/language/providers/requirements#local-names
	ProviderName string
	// ProviderSource is the source address of the provider.
	// E.g. registry.terraform.io/hashicorp/aws
	// https://developer.hashicorp.com/terraform/language/providers/requirements#source-addresses
	ProviderSource string
	// ProviderVersion is the version of thr provider.
	// E.g. 4.49.0
	ProviderVersion string
}

type SchemaType string

const (
	SchemaTypeProvider   SchemaType = "provider"
	SchemaTypeResource   SchemaType = "resource"
	SchemaTypeDataSource SchemaType = "data"
)

// SchemaProvider creates a schema for the provider config block for the
// provider
// represented by ProviderGenerator
func (a *ProviderGenerator) SchemaProvider(sb *tfjson.SchemaBlock) *Schema {
	return &Schema{
		Description:          sb.Description,
		Deprecated:           sb.Deprecated,
		SchemaType:           SchemaTypeProvider,
		GeneratedPkgLocation: a.GeneratedPackageLocation, // gen/aws
		ProviderName:         a.ProviderName,             // aws
		ProviderSource:       a.ProviderSource,           // registry.terraform.io/hashicorp/aws
		ProviderVersion:      a.ProviderVersion,          // 4.49.0
		PackageName:          a.ProviderName,             // aws
		Type:                 "provider",
		StructName:           "Provider",
		ArgumentStructName:   "Provider", // Edge case for provider: args struct *is* the provider struct.
		StateStructName:      "n/a",      // Providers do not have a state.
		Receiver:             structReceiverFromName("provider"),

		SubPkgName: a.ProviderName,
		SubPkgPath: filepath.Join(
			a.GeneratedPackageLocation,
			"provider_types"+fileExtension,
		),
		FilePath: filepath.Join(
			a.GeneratedPackageLocation,
			"provider"+fileExtension,
		),
		graph: newGraph(sb),
	}
}

// SchemaResource creates a schema for the given resource for the provider
// represented by ProviderGenerator
func (a *ProviderGenerator) SchemaResource(
	name string,
	sb *tfjson.SchemaBlock,
) *Schema {
	rs := &Schema{
		Description:          sb.Description,
		Deprecated:           sb.Deprecated,
		SchemaType:           SchemaTypeResource,
		GeneratedPkgLocation: a.GeneratedPackageLocation, // gen/aws
		ProviderName:         a.ProviderName,             // aws
		ProviderSource:       a.ProviderSource,           // hashicorp/aws
		ProviderVersion:      a.ProviderVersion,          // 4.49.0
		PackageName:          name,                       // aws_iam_role
		Type:                 name,                       // aws_iam_role

		StructName:         "Resource",
		ArgumentStructName: suffixArgs, // Args
		AttributesStructName: strcase.Camel(
			name,
		) + suffixAttributes, // iam_role => awsIamRoleAttributes
		StateStructName: strcase.Camel(
			name,
		) + suffixState, // aws_iam_role => awsIamRoleState
		Receiver: structReceiverFromName(
			name,
		), // iam_role => ir

		SubPkgName: name, // aws_iam_role => aws_iam_role
		SubPkgPath: filepath.Join(
			a.GeneratedPackageLocation,
			name,
			name+"_types"+fileExtension,
		),
		FilePath: filepath.Join(
			a.GeneratedPackageLocation,
			name,
			name+fileExtension,
		),
		graph: newGraph(sb),
	}
	return rs
}

// SchemaData creates a schema for the given data object for the provider
// represented by ProviderGenerator
func (a *ProviderGenerator) SchemaData(
	name string,
	sb *tfjson.SchemaBlock,
) *Schema {
	dataName := "data_" + name
	ds := &Schema{
		Description:          sb.Description,
		Deprecated:           sb.Deprecated,
		SchemaType:           SchemaTypeDataSource,
		GeneratedPkgLocation: a.GeneratedPackageLocation, // gen/aws
		ProviderName:         a.ProviderName,             // aws
		ProviderSource:       a.ProviderSource,           // hashicorp/aws
		ProviderVersion:      a.ProviderVersion,          // 4.49.0
		PackageName:          name,                       // aws_iam_role
		Type:                 name,                       // aws_iam_role

		StructName:         "DataSource",
		ArgumentStructName: prefixStructDataSource + suffixArgs, // aws_iam_role => DataArgs
		AttributesStructName: strcase.Camel(
			dataName,
		) + suffixAttributes, // iam_role => dataAwsIamRoleAttributes
		StateStructName: "n/a", // Data sources do not have a state.
		Receiver: structReceiverFromName(
			name,
		), // iam_role => ir

		SubPkgName: name, // aws_iam_role => aws_iam_role
		SubPkgPath: filepath.Join(
			a.GeneratedPackageLocation,
			name,
			dataName+"_types"+fileExtension,
		), // gen/aws/aws_iam_role/data_aws_iam_role_types.go
		FilePath: filepath.Join(
			a.GeneratedPackageLocation,
			name,
			dataName+fileExtension,
		), // gen/aws/aws_iam_role/data_aws_iam_role.go
		graph: newGraph(sb),
	}

	return ds
}

// structReceiverFromName calculates a suitable receiver from the name of the
// object. It gets the first character of each word separated by underscores,
// e.g. iam_role => ir
func structReceiverFromName(name string) string {
	ss := strings.Split(name, "_")
	var receiver strings.Builder
	for _, s := range ss {
		receiver.WriteString(s[0:1])
	}
	r := receiver.String()
	// Avoid using keywords for the receiver!
	if token.Lookup(r).IsKeyword() || r == "nil" {
		r = "_" + r
	}
	return r
}

// Schema is used to store all the relevant information required for the Go
// code generator.
// A schema can represent a resource, a data object or the provider
// configuration.
type Schema struct {
	Description          string
	Deprecated           bool
	SchemaType           SchemaType // resource / provider / data
	GoProviderPkgPath    string     // github.com/golingon/lingon/gen/providers
	GeneratedPkgLocation string     // gen/providers/aws
	ProviderName         string     // aws
	ProviderSource       string     // hashicorp/aws
	ProviderVersion      string     // 4.49.0
	PackageName          string     // aws_iam_role
	Type                 string     // aws_iam_role

	// Structs
	StructName           string // Resource / DataSource / Provider
	ArgumentStructName   string // Args / DataArgs
	AttributesStructName string // iam_role => awsIamRoleAttributes
	StateStructName      string // iam_role => awsIamRoleState

	Receiver string // iam_role => ir

	SubPkgName string // iam_role => iamrole
	// SubPkgPath is the filepath for the schema entities types (args,
	// attributes, state).
	SubPkgPath string // aws_iam_role
	FilePath   string // gen/providers/aws/ xxx
	graph      *graph
}

func (s *Schema) Comment() string {
	str := strings.Builder{}

	if s.Description != "" {
		str.WriteString(s.Description + "\n\n")
	}
	switch s.SchemaType {
	case SchemaTypeProvider:
		str.WriteString(
			fmt.Sprintf(
				"%s is the provider for %s.",
				s.StructName,
				s.ProviderSource,
			),
		)
	case SchemaTypeResource:
		str.WriteString(
			fmt.Sprintf(
				"%s is the resource %s.",
				s.StructName,
				s.Type,
			),
		)
	case SchemaTypeDataSource:
		str.WriteString(
			fmt.Sprintf(
				"%s is the data source %s.",
				s.StructName,
				s.Type,
			),
		)
	}

	docsURL := fmt.Sprintf(
		"https://registry.terraform.io/providers/%s/%s/docs",
		s.ProviderSource,
		s.ProviderVersion,
	)
	str.WriteString("\n\nDocumentation: " + docsURL)

	if s.Deprecated {
		str.WriteString("\n\nDeprecated.")
	}

	return str.String()
}

func (s *Schema) SubPkgQualPath() string {
	return s.GoProviderPkgPath + "/" + s.SubPkgName
}
