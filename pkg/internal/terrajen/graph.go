// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terrajen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/veggiemonk/strcase"

	"github.com/dave/jennifer/jen"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

func newGraph(schema *tfjson.SchemaBlock) *graph {
	root := &node{}
	g := graph{
		root: root,
	}

	g.processAttributes(root, nil, schema.Attributes)

	for _, blockName := range sortMapKeys(schema.NestedBlocks) {
		blockType := schema.NestedBlocks[blockName]
		root.children = append(
			root.children,
			g.traverseBlockType(nil, blockName, blockType),
		)
	}

	g.calculateUniqueTypeNames()

	return &g
}

// graph is used to decouple the tfjson.SchemaBlock type from the code
// generator.
// A graph is created by supplying a tfjson.SchemaBlock.
type graph struct {
	// root node is the root of the schema block and does not have things like a
	// name.
	root *node
	// nodes contains all nodes in the schema block
	nodes []*node
}

type node struct {
	name        string
	description string
	deprecated  bool
	// path is the list of nodes from the root (resource/data/provider)
	// node to this node.
	// E.g. if a resource "x" has a block "y" which contains this node's
	// block "z", then
	// name = "z"
	// path = ["x", "y"]
	path []string
	// uniqueName is a calculated unique name within the subpkg for a
	// resource/data/provider object.
	// If the name is already unique, then the uniqueName is the name.
	// If the name is not unique, take the last element from the path (if any)
	// and prefix is to the name until the name is unique
	uniqueName string
	attributes []*attribute

	// isAttribute is true if the node is an attribute (not a block) in HCL.
	// A block is a nested object, like:
	// 	node {
	// 	  attribute = "value"}
	// 	}
	//
	// An attribute is a key-value pair, like:
	// 	node = {
	// 	  attribute = "value"
	// 	}
	isAttribute bool
	// isArg is true if the node can be passed as an argument in the terraform
	// configuration.
	isArg bool
	// nestingPath is the path from one Go struct to it's child.
	// For Terraform Schema Blocks this is only ever going to be
	// length 0 (for single block) or 1 (for list/set/map).
	// For Cty types, this is a bit more complicated as an object
	// can be nested within a list/set/map with no limit to the depth,
	// hence why we use a list here, and that list represents the
	// path from object --> object. E.g. if we have the cty pseudo type
	// list(set(map(object))) then the nestingPath would be:
	// [list, set, map]
	nestingPath []nodeNestingMode
	isRequired  bool
	// minItems is taken from the HCL block type, and is not set
	// when using Cty types.
	minItems uint64
	// maxItems is taken from the HCL block type, and is not set
	// when using Cty types.
	maxItems uint64

	children []*node

	receiver string
}

func (n *node) isSingularArg() bool {
	return len(n.nestingPath) == 0 || n.maxItems == 1
}

func (n *node) isSingularState() bool {
	return len(n.nestingPath) == 0
}

func (n *node) comment() string {
	str := strings.Builder{}

	str.WriteString(strcase.Pascal(n.uniqueName))

	if n.isSingularArg() {
		if n.isRequired {
			str.WriteString(" is required. ")
		} else {
			str.WriteString(" is optional. ")
		}
	} else {
		str.WriteString(" is " + nodeBlockListValidateTags(n) + ". ")
	}
	if n.description != "" {
		str.WriteString(strings.ReplaceAll(n.description, "*/", "\\*\\/"))
	}
	if n.deprecated {
		str.WriteString("\n\nDeprecated: see description.\n")
	}
	return str.String()
}

type nodeNestingMode int

const (
	nodeNestingModeList nodeNestingMode = 1
	nodeNestingModeSet  nodeNestingMode = 2
	nodeNestingModeMap  nodeNestingMode = 3
)

type attribute struct {
	name        string
	description string
	deprecated  bool
	ctyType     cty.Type
	// isArg is true if the attribute can be passed as an argument
	// to the node schema block, else it is false
	isArg bool
	// isRequired is true if the attribute can be passed as an argument
	// and is required, else it is false
	isRequired bool
}

func (a *attribute) comment() string {
	str := strings.Builder{}
	str.WriteString(strcase.Pascal(a.name))

	if a.isRequired {
		str.WriteString(" is required. ")
	} else {
		str.WriteString(" is optional. ")
	}
	if a.description != "" {
		str.WriteString(strings.ReplaceAll(a.description, "*/", "\\*\\/"))
	}
	if a.deprecated {
		str.WriteString("\n\nDeprecated: see description.\n")
	}
	return str.String()
}

func (g *graph) isEmpty() bool {
	return len(g.nodes) == 0
}

func (g *graph) processAttributes(
	node *node,
	path []string,
	attributes map[string]*tfjson.SchemaAttribute,
) {
	for _, atName := range sortMapKeys(attributes) {
		attr := attributes[atName]
		// Attributes which are objects will be treated as children.
		if attr.AttributeNestedType != nil {
			child := g.traverseCtyNestedType(
				path,
				atName,
				attr.AttributeNestedType,
				isAttributeArg(attr),
			)
			child.description = attr.Description
			node.children = append(
				node.children,
				child,
			)
			continue
		}
		if _, ok := ctyTypeElementObject(attr.AttributeType); ok {
			child := g.traverseCtyType(
				path,
				atName,
				attr.AttributeType,
				isAttributeArg(attr),
			)
			child.description = attr.Description
			node.children = append(node.children, child)
			continue
		}
		node.attributes = append(
			node.attributes, &attribute{
				name:        atName,
				description: attr.Description,
				deprecated:  attr.Deprecated,
				ctyType:     attr.AttributeType,
				isArg:       isAttributeArg(attr),
				isRequired:  attr.Required,
			},
		)
	}
}

func (g *graph) traverseBlockType(
	path []string,
	name string,
	blockType *tfjson.SchemaBlockType,
) *node {
	n := node{
		name:        name,
		description: blockType.Block.Description,
		deprecated:  blockType.Block.Deprecated,
		path:        path,
		uniqueName:  name,
		nestingPath: blockNodeNestingMode(blockType.NestingMode),
		isRequired:  isArgBlockRequired(blockType),
		minItems:    blockType.MinItems,
		maxItems:    blockType.MaxItems,
		// Blocks can always be arguments
		isArg:    true,
		receiver: structReceiverFromName(name),
	}
	g.nodes = append(g.nodes, &n)

	// First handle the attributes.
	g.processAttributes(&n, appendPath(path, name), blockType.Block.Attributes)

	for _, blockName := range sortMapKeys(blockType.Block.NestedBlocks) {
		blockType := blockType.Block.NestedBlocks[blockName]
		n.children = append(
			n.children,
			g.traverseBlockType(appendPath(path, name), blockName, blockType),
		)
	}
	return &n
}

func (g *graph) traverseCtyNestedType(
	path []string,
	name string,
	nestedType *tfjson.SchemaNestedAttributeType,
	isArg bool,
) *node {
	n := node{
		name:        name,
		path:        path,
		uniqueName:  name,
		nestingPath: blockNodeNestingMode(nestedType.NestingMode),
		isRequired:  false,
		isAttribute: true, // Nested types are always attributes.
		isArg:       isArg,
		receiver:    structReceiverFromName(name),
	}
	g.nodes = append(g.nodes, &n)

	g.processAttributes(&n, append(path, name), nestedType.Attributes)

	return &n
}

func (g *graph) traverseCtyType(
	path []string,
	name string,
	ct cty.Type,
	isArg bool,
) *node {
	n := node{
		name:        name,
		path:        path,
		uniqueName:  name,
		nestingPath: ctyNodeNestingMode(ct),
		isRequired:  false,
		isAttribute: true, // Cty types are always attributes.
		isArg:       isArg,
		receiver:    structReceiverFromName(name),
	}
	g.nodes = append(g.nodes, &n)

	// Get the underlying object
	obj, _ := ctyTypeElementObject(ct)
	for _, atName := range sortMapKeys(obj.AttributeTypes()) {
		at := obj.AttributeType(atName)
		// If there are objects within the attributes of this object, traverse
		// those objects
		// and make them children of this object
		if _, ok := ctyTypeElementObject(at); ok {
			n.children = append(
				n.children,
				g.traverseCtyType(appendPath(path, name), atName, at, isArg),
			)
			continue
		}
		n.attributes = append(
			// No information at this level on description or deprecation.
			// All we have is a cty.Type.
			n.attributes, &attribute{
				name:        atName,
				description: "",    // N/A
				deprecated:  false, // N/A
				ctyType:     at,
				isArg:       isArg,
				isRequired:  false,
			},
		)
	}
	return &n
}

// calculateUniqueTypeNames iterates over all nodes in the graph and calculates
// unique type names for each node.
// For most cases, this will be the path to the node with the node name itself
// appended on the end.
//
// To avoid generating incredibly long type names, if the node path is
// longer than 5 take the first 2 elements and generate a short hash of
// the rest as a suffix.
//
// This will provide strong uniqueness guarantees and should not affect
// the developer experience because Go's auto-complete type suggestions
// will find the right type for you.
// And if you accidently use the wrong type you will get a compile
// error.
//
// This was the lesser evil compared with struct names that are 50+ characters
// of words in PascalCase.
func (g *graph) calculateUniqueTypeNames() {
	for _, node := range g.nodes {
		nodePath := make([]string, len(node.path)+1)
		copy(nodePath, node.path)
		nodePath[len(node.path)] = node.name

		if len(nodePath) >= 5 {
			prefix := strings.Join(nodePath[:2], ".")
			suffix := strings.Join(nodePath[2:], ".")

			hash := shortHash(suffix)
			node.uniqueName = fmt.Sprintf("%s.%s", prefix, hash)
			continue

		}
		node.uniqueName = strings.Join(nodePath, ".")
	}
}

func shortHash(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))[0:8]
}

func appendPath(path []string, name string) []string {
	newPath := make([]string, len(path)+1)
	copy(newPath, path)
	newPath[len(path)] = name
	return newPath
}

func ctyTypeElementObject(ct cty.Type) (cty.Type, bool) {
	switch {
	case ct.IsObjectType():
		return ct, true
	case ct.IsCollectionType():
		return ctyTypeElementObject(ct.ElementType())
	default:
		return cty.NilType, false
	}
}

func blockNodeNestingMode(
	nestingMode tfjson.SchemaNestingMode,
) []nodeNestingMode {
	switch nestingMode {
	case tfjson.SchemaNestingModeSingle, tfjson.SchemaNestingModeGroup:
		return nil
	case tfjson.SchemaNestingModeList, tfjson.SchemaNestingModeMap:
		// Unintuitively, tfjson.SchemaNestingModeMap is not actually a map,
		// just a list,
		// but they get keyed by the block labels into a Map.
		// For our use case, we therefore treat it like a list.
		return []nodeNestingMode{nodeNestingModeList}
	case tfjson.SchemaNestingModeSet:
		return []nodeNestingMode{nodeNestingModeSet}
	default:
		panic(
			fmt.Sprintf(
				"unsupported SchemaNestingMode: %s",
				nestingMode,
			),
		)
	}
}

// ctyNodeNestingMode calculates the nesting mode for a given cty.Type that is
// known to contain a cty.Object.
// The nesting path is from parent to the child cty.Object,
// which may contain lists, sets or maps along the way.
// I.e. what is the path from parent --> cty.Object (child)
func ctyNodeNestingMode(ct cty.Type) []nodeNestingMode {
	if ct.IsObjectType() {
		return nil
	}
	switch {
	case ct.IsListType():
		return append(
			[]nodeNestingMode{nodeNestingModeList},
			ctyNodeNestingMode(ct.ElementType())...,
		)
	case ct.IsSetType():
		return append(
			[]nodeNestingMode{nodeNestingModeSet},
			ctyNodeNestingMode(ct.ElementType())...,
		)
	case ct.IsMapType(), ct.IsCollectionType():
		return append(
			[]nodeNestingMode{nodeNestingModeMap},
			ctyNodeNestingMode(ct.ElementType())...,
		)
	default:
		panic(
			fmt.Sprintf(
				"unsupported cty.Type nesting mode: %s",
				ct.FriendlyName(),
			),
		)
	}
}

func isArgBlockRequired(block *tfjson.SchemaBlockType) bool {
	if block.MaxItems == 1 && block.MinItems == 1 {
		return true
	}
	return false
}

func isAttributeArg(attr *tfjson.SchemaAttribute) bool {
	// If it is not computed, then it must be an argument.
	// If it is computed, and it is optional, then it can be an argument.
	// If it is required, then it must be an argument.
	switch {
	case !attr.Computed:
		return true
	case attr.Computed && attr.Optional:
		return true
	case attr.Required:
		return true
	default:
		return false
	}
}

// returnTypeFromNestingPath returns the jen statement for the type that
// node represents, e.g.
//
//	terra.ListValue[emrcluster.StepRef]
func returnTypeFromNestingPath(
	nestingPath []nodeNestingMode,
	qual *jen.Statement,
) *jen.Statement {
	fullQual := qual.Clone()
	// Iterate over nestingPath backwards to get the correct order.
	// E.g. if List-->Set-->Map is the nestingPath, then we can reverse this to
	// easily construct List[Set[Map[qual]]]
	for i := len(nestingPath) - 1; i >= 0; i-- {
		path := nestingPath[i]
		switch path {
		case nodeNestingModeList:
			fullQual = qualListValue().Types(fullQual.Clone())
		case nodeNestingModeSet:
			fullQual = qualSetValue().Types(fullQual.Clone())
		case nodeNestingModeMap:
			fullQual = qualMapValue().Types(fullQual.Clone())
		default:
			panic(fmt.Sprintf("unsupport node nesting type: %d", nestingPath))
		}
	}
	return fullQual
}

// jenNodeReturnValue returns the jen statement for the return value that
// node represents, e.g.
//
//	terra.ReferenceList[emrcluster.StepRef]
func jenNodeReturnValue(
	n *node, childQual *jen.Statement,
) *jen.Statement {
	// If the child is not nested in any lists/sets/maps then simply return
	// a reference to the single child qualifier
	if len(n.nestingPath) == 0 {
		return qualReferenceAsSingle().Types(childQual)
	}

	first := n.nestingPath[0]
	remainder := n.nestingPath[1:]
	subType := returnTypeFromNestingPath(remainder, childQual)

	var fullQual *jen.Statement
	switch first {
	case nodeNestingModeList:
		fullQual = qualReferenceAsList().Types(subType)
	case nodeNestingModeSet:
		fullQual = qualReferenceAsSet().Types(subType)
	case nodeNestingModeMap:
		fullQual = qualReferenceAsMap().Types(subType)
	default:
		panic(fmt.Sprintf("unsupport node nesting type: %d", n.nestingPath))
	}
	return fullQual
}

func sortMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
