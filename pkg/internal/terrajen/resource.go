// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terrajen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/veggiemonk/strcase"
)

// ResourceFile generates a Go file for a Terraform resource configuration based
// on the given Schema
func ResourceFile(s *Schema) *jen.File {
	f := jen.NewFile(s.PackageName)
	f.ImportAlias(pkgHCL, pkgHCLAlias)
	f.ImportName(pkgTerra, pkgTerraAlias)
	f.HeaderComment(HeaderComment)
	f.Add(resourceStructCompileCheck(s))
	f.Add(resourceStruct(s))
	f.Add(argsStruct(s))
	f.Add(attributesStruct(s))
	f.Add(resourceStateStruct(s))

	return f
}

func resourceStructCompileCheck(s *Schema) *jen.Statement {
	return jen.Var().Op("_").Qual(pkgTerra, "Resource").Op("=").
		Params(
			jen.Op("*").Id(s.StructName),
		).
		Params(jen.Nil()).
		Line()
}

func resourceStruct(s *Schema) *jen.Statement {
	stmt := jen.Comment(
		fmt.Sprintf(
			"%s represents the Terraform resource %s.",
			s.StructName,
			s.Type,
		),
	).
		Line().
		Type().Id(s.StructName).Struct(
		jen.Id(idFieldName).String(),
		jen.Id(idFieldArgs).Id(s.ArgumentStructName),
		jen.Id(idFieldState).Op("*").Id(s.StateStructName),
		jen.Id(idFieldDependsOn).Add(qualTypeDependencies()),
		jen.Id(idFieldLifecycle).Op("*").Add(qualStructLifecycle()),
	)
	stmt.Line()
	stmt.Line()

	// Methods
	// Type
	stmt.Add(funcSchemaType(s, "Type"))
	stmt.Line()
	stmt.Line()
	// LocalName
	stmt.Add(funcLocalName(s))
	stmt.Line()
	stmt.Line()
	// Configuration
	stmt.Add(funcConfiguration(s))
	stmt.Line()
	stmt.Line()
	// DependOn
	stmt.Add(funcDependOn(s))
	stmt.Line()
	stmt.Line()
	// Dependencies
	stmt.Add(funcDependencies(s))
	stmt.Line()
	stmt.Line()
	// LifecycleManagement
	stmt.Add(funcLifecycleManagement(s))
	stmt.Line()
	stmt.Line()
	// Attributes
	stmt.Add(funcAttributes(s))
	stmt.Line()
	stmt.Line()
	// ImportState
	stmt.Add(funcResourceImportState(s))
	stmt.Line()
	stmt.Line()
	// State
	stmt.Add(funcResourceState(s))
	stmt.Line()
	stmt.Line()
	// StateMust
	stmt.Add(funcResourceStateMust(s))
	stmt.Line()
	stmt.Line()

	return stmt
}

func resourceStateStruct(s *Schema) *jen.Statement {
	fields := make([]jen.Code, 0)
	for _, attr := range s.graph.root.attributes {
		pan := strcase.Pascal(attr.name)
		stmt := jen.Id(pan)
		stmt.Add(ctyTypeToGoType(attr.ctyType, pan))
		stmt.Tag(
			map[string]string{
				tagJSON: attr.name,
			},
		)
		fields = append(fields, stmt)
	}

	for _, child := range s.graph.root.children {
		stmt := jen.Id(strcase.Pascal(child.name))
		if len(child.nestingPath) == 0 {
			stmt.Op("*")
		} else {
			stmt.Index()
		}

		stmt.Id(strcase.Pascal(child.uniqueName) + suffixState)
		// stmt.Qual(s.SubPkgQualPath(), strcase.Pascal(child.name)+suffixState)
		stmt.Tag(
			map[string]string{
				tagJSON: child.name,
			},
		)
		fields = append(fields, stmt)
	}
	return jen.
		Type().
		Id(s.StateStructName).
		Struct(fields...).
		Line().
		Line()
}
