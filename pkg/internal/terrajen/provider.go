// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terrajen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

// ProviderFile generates a Go file for a Terraform provider configuration based
// on the given Schema
func ProviderFile(s *Schema) *jen.File {
	f := jen.NewFile(s.ProviderName)
	f.ImportName(pkgTerra, pkgTerraAlias)
	f.HeaderComment(HeaderComment)
	f.Add(providerStructCompileCheck(s))
	f.Add(providerStruct(s))

	return f
}

func providerStructCompileCheck(s *Schema) *jen.Statement {
	return jen.Var().Op("_").Qual(pkgTerra, "Provider").Op("=").
		Params(
			jen.Op("*").Id(s.StructName),
		).
		Params(jen.Nil()).
		Line()
}

func providerStruct(s *Schema) *jen.Statement {
	// Use the args struct as the main struct, because there is nothing else to
	// go in the provider.
	stmt := argsStruct(s)
	stmt.Line()
	stmt.Line()

	// LocalName
	stmt.Add(funcProviderLocalName(s))
	stmt.Line()
	stmt.Line()
	// Source
	stmt.Add(funcProviderSource(s))
	stmt.Line()
	stmt.Line()
	// Version
	stmt.Add(funcProviderVersion(s))
	stmt.Line()
	stmt.Line()
	// Configuration
	stmt.Add(funcProviderConfiguration(s))
	stmt.Line()
	stmt.Line()

	return stmt
}

func funcProviderConfiguration(s *Schema) *jen.Statement {
	return jen.Comment(
		fmt.Sprintf(
			"%s returns the provider configuration for [%s].",
			idFuncConfiguration,
			s.StructName,
		),
	).
		Line().
		Func().
		// Receiver
		Params(jen.Id(s.Receiver).Op("*").Id(s.StructName)).
		// Name
		Id(idFuncConfiguration).Call().
		// Return type
		Interface().
		// Body
		Block(
			jen.Return(
				jen.Id(s.Receiver),
			),
		)
}
