// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terrajen

import (
	"github.com/dave/jennifer/jen"
)

// DataSourceFile generates a Go file for a Terraform data source configuration
// based on the given
// Schema
func DataSourceFile(s *Schema) *jen.File {
	f := jen.NewFile(s.PackageName)
	f.ImportAlias(pkgHCL, "hcl")
	f.ImportName(pkgTerra, pkgTerraAlias)
	f.HeaderComment(HeaderComment)
	f.Add(dataStructCompileCheck(s))
	f.Add(dataStruct(s))
	f.Add(argsStruct(s))
	f.Add(attributesStruct(s))

	return f
}

func dataStructCompileCheck(s *Schema) *jen.Statement {
	return jen.Var().Op("_").Qual(pkgTerra, "DataSource").Op("=").
		Params(
			jen.Op("*").Id(s.StructName),
		).
		Params(jen.Nil()).
		Line()
}

func dataStruct(s *Schema) *jen.Statement {
	stmt := jen.Comment(s.Comment()).
		Line().
		Type().Id(s.StructName).Struct(
		jen.Id(idFieldName).String(),
		jen.Id(idFieldArgs).Id(s.ArgumentStructName),
	)
	stmt.Line()
	stmt.Line()

	// DataSource
	stmt.Add(funcSchemaType(s, "DataSource"))
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
	// Attributes
	stmt.Add(funcAttributes(s))
	stmt.Line()
	stmt.Line()

	return stmt
}
