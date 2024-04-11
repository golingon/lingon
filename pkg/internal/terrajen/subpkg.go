// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terrajen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/veggiemonk/strcase"
)

// SubPkgFile generates a Go file for the given schema.
// The schema should represent a sub-package or be the sub-types of a top-level
// provider/resource/data source.
//
// For example, the AWS provider has a top-level provider config, with many
// nested subtypes.
// SubPkgFile would generate a file containing all the subtypes.
func SubPkgFile(s *Schema) (*jen.File, bool) {
	// Skip sub pkg if there are no blocks to render
	if s.graph.isEmpty() {
		return nil, false
	}
	f := jen.NewFile(s.SubPkgName)
	f.ImportAlias(pkgHCL, "hcl")
	f.HeaderComment(HeaderComment)
	for _, n := range s.graph.nodes {
		if n.isArg {
			f.Add(subPkgArgStruct(n, s.SchemaType))
		}
	}
	for _, n := range s.graph.nodes {
		f.Add(subPkgAttributeStruct(n, s.SchemaType))
	}
	for _, n := range s.graph.nodes {
		f.Add(subPkgStateStruct(n, s.SchemaType))
	}

	return f, true
}

func subPkgArgStruct(n *node, schemaType SchemaType) *jen.Statement {
	fields := make([]jen.Code, 0)

	for _, attr := range n.attributes {
		// Skip attributes that are not arguments.
		if !attr.isArg {
			continue
		}
		stmt := jen.Comment(attr.comment()).Line()
		stmt.Add(jen.Id(strcase.Pascal(attr.name)))
		stmt.Add(ctyTypeReturnType(attr.ctyType))
		// Add tags
		tags := map[string]string{
			tagHCL: attr.name + ",attr",
		}
		if attr.isRequired {
			tags[tagValidate] = "required"
		}
		stmt.Tag(tags)
		fields = append(fields, stmt)
	}

	for _, child := range n.children {
		// Skip attributes that are not arguments.
		if !child.isArg {
			continue
		}
		stmt := jen.Comment(child.comment()).Line()
		stmt.Add(jen.Id(strcase.Pascal(child.name)))
		tags := map[string]string{
			tagHCL: child.name + ",block",
		}
		if child.isSingularArg() {
			stmt.Op("*")
			if child.isRequired {
				tags[tagValidate] = "required"
			}
		} else {
			// For children the nesting type cannot be a map
			for _, path := range child.nestingPath {
				switch path {
				case nodeNestingModeList, nodeNestingModeSet:
					stmt.Index()
				default:
					panic(
						fmt.Sprintf(
							"unsupported nesting path %d for child",
							path,
						),
					)
				}
			}
			tags[tagValidate] = nodeBlockListValidateTags(child)
		}
		stmt.Id(subPkgArgStructName(child, schemaType))
		stmt.Tag(tags)
		fields = append(fields, stmt)
	}

	stmt := jen.
		Type().Id(subPkgArgStructName(n, schemaType)).
		Struct(fields...)
	stmt.Line()
	stmt.Line()

	return stmt
}

func subPkgArgStructName(n *node, schemaType SchemaType) string {
	if schemaType == SchemaTypeDataSource {
		return prefixStructDataSource + strcase.Pascal(n.uniqueName)
	}
	return strcase.Pascal(n.uniqueName)
}

func subPkgAttributeStruct(n *node, schemaType SchemaType) *jen.Statement {
	structName := subPkgAttributeStructName(n, schemaType)

	structFieldRef := "ref"
	refArg := "ref"
	stmt := jen.Type().Id(structName).
		Struct(
			jen.Id(structFieldRef).Add(qualReferenceValue()),
		)
	stmt.Line()
	stmt.Line()

	// Methods
	// Override InternalRef, e.g.
	//
	// 	func (i OidcRef) InternalRef() (terra.Reference, error) {
	// 		return i.ref, nil
	// 	}
	stmt.Add(
		jen.Func().
			// Receiver
			Params(jen.Id(n.receiver).Id(structName)).
			// Name
			Id(idFuncInternalRef).Call().
			// Return type
			Params(qualReferenceValue(), jen.Error()).
			// Body
			Block(
				jen.Return(
					jen.Id(n.receiver).Dot(
						structFieldRef,
					),
					jen.Nil(),
				),
			),
	)
	stmt.Line()
	stmt.Line()

	// Override InternalWithRef, e.g.
	//
	// 	func (i OidcRef) InternalWithRef(ref terra.Reference) OidcRef {
	// 		return terra.ReferenceSingle[OidcRef](ref)
	// 	}
	stmt.Add(
		jen.Func().
			// Receiver
			Params(jen.Id(n.receiver).Id(structName)).
			// Name
			Id(idFuncInternalWithRef).Call(
			jen.Id(refArg).Add(qualReferenceValue()),
		).
			// Return type
			Id(structName).
			// Body
			Block(
				jen.Return(
					jen.Id(structName).Values(
						jen.Dict{
							jen.Id(structFieldRef): jen.Id(refArg),
						},
					),
				),
			),
	)
	stmt.Line()
	stmt.Line()
	// Override InternalTokens
	stmt.Add(
		jen.Func().
			// Receiver
			Params(jen.Id(n.receiver).Id(structName)).
			// Name
			Id(idFuncInternalTokens).
			Call().
			// Return type
			Params(qualHCLWriteTokens(), jen.Error()).
			// Body
			Block(
				jen.Return(
					jen.Id(n.receiver).Dot(
						structFieldRef,
					).Dot(idFuncInternalTokens).Call(),
				),
			),
	)
	stmt.Line()
	stmt.Line()

	for _, attr := range n.attributes {
		appendRef := jen.Id(n.receiver).
			Dot(refArg).
			Dot("Append").
			Call(jen.Lit(attr.name))
		stmt.Add(
			jen.Func().
				// Receiver
				Params(jen.Id(n.receiver).Id(structName)).
				// Name
				Id(strcase.Pascal(attr.name)).Call().
				//	Return type
				Add(ctyTypeReturnType(attr.ctyType)).
				// Body
				Block(
					jen.Return(
						funcReferenceByCtyType(attr.ctyType).
							Call(appendRef),
					),
				),
		)
		stmt.Line()
		stmt.Line()
	}

	for _, child := range n.children {
		childStructName := subPkgAttributeStructName(child, schemaType)
		appendRef := jen.Id(n.receiver).
			Dot(refArg).
			Dot("Append").
			Call(jen.Lit(child.name))

		stmt.Add(
			jen.Func().
				// Receiver
				Params(jen.Id(n.receiver).Id(structName)).
				// Name
				Id(strcase.Pascal(child.name)).Call().
				// Return type
				Add(
					returnTypeFromNestingPath(
						child.nestingPath,
						jen.Id(childStructName),
					),
				).Block(
				jen.Return(
					jenNodeReturnValue(child, jen.Id(childStructName)).
						Call(appendRef),
				),
			),
		)
		stmt.Line()
		stmt.Line()
	}

	return stmt
}

func subPkgAttributeStructName(n *node, schemaType SchemaType) string {
	structName := strcase.Pascal(n.uniqueName) + suffixAttributes
	if schemaType == SchemaTypeDataSource {
		return prefixStructDataSource + structName
	}
	return structName
}

func subPkgStateStruct(n *node, schemaType SchemaType) *jen.Statement {
	fields := make([]jen.Code, 0)

	for _, attr := range n.attributes {
		pan := strcase.Pascal(attr.name)
		stmt := jen.Id(pan)
		stmt.Add(ctyTypeToGoType(attr.ctyType, pan))
		// Add tags
		stmt.Tag(
			map[string]string{
				tagJSON: attr.name,
			},
		)
		fields = append(fields, stmt)
	}

	for _, child := range n.children {
		stmt := jen.Id(strcase.Pascal(child.name))
		if child.isSingularState() {
			stmt.Op("*")
		} else {
			stmt.Index()
		}
		stmt.Id(subPkgStateStructName(child, schemaType))
		stmt.Tag(
			map[string]string{
				tagJSON: child.name,
			},
		)
		fields = append(fields, stmt)
	}

	stmt := jen.
		Type().Id(subPkgStateStructName(n, schemaType)).
		Struct(fields...)
	stmt.Line()
	stmt.Line()

	return stmt
}

func subPkgStateStructName(n *node, schemaType SchemaType) string {
	structName := strcase.Pascal(n.uniqueName) + suffixState
	if schemaType == SchemaTypeDataSource {
		return prefixStructDataSource + structName
	}
	return structName
}
