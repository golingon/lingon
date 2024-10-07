// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terrajen

import (
	"encoding/json"
	"fmt"
	"testing"

	tu "github.com/golingon/lingon/pkg/testutil"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
)

type schemaAttribute struct {
	Name string
	tfjson.SchemaAttribute
}

type schemaBlockType struct {
	Name string
	tfjson.SchemaBlockType
}

func TestGraph_AttributeString(t *testing.T) {
	ex := schemaAttribute{
		Name: "string",
		SchemaAttribute: tfjson.SchemaAttribute{
			AttributeType: cty.String,
			Description:   "string",
			Required:      true,
		},
	}
	sb := tfjson.SchemaBlock{
		Attributes: map[string]*tfjson.SchemaAttribute{
			ex.Name: &ex.SchemaAttribute,
		},
	}

	g := newGraph(&sb)
	tu.IsEqual(t, len(g.root.attributes), 1)
	ac := g.root.attributes[0]
	tu.IsEqual(t, ac.name, ex.Name)
	tu.True(t, ex.AttributeType.Equals(ac.ctyType), "attribute type mismatch")
	tu.True(t, ac.isArg, "ac.isArg")
	tu.True(t, ac.isRequired, "ac.isRequired")
}

func TestGraph_Attributes(t *testing.T) {
	count := 10
	attrs := make(map[string]*tfjson.SchemaAttribute, count)
	for i := 0; i < count; i++ {
		attrs[fmt.Sprintf("%d", i)] = &tfjson.SchemaAttribute{
			AttributeType: cty.String,
			Description:   "string",
			Required:      true,
		}
	}
	sb := tfjson.SchemaBlock{
		Attributes: attrs,
	}

	g := newGraph(&sb)
	tu.IsEqual(t, len(g.root.attributes), 10)
}

func TestGraph_Blocks(t *testing.T) {
	exAttr := schemaAttribute{
		Name: "string",
		SchemaAttribute: tfjson.SchemaAttribute{
			AttributeType: cty.String,
			Description:   "string",
			Required:      true,
		},
	}
	ex := schemaBlockType{
		Name: "block",
		SchemaBlockType: tfjson.SchemaBlockType{
			NestingMode: tfjson.SchemaNestingModeSingle,
			Block: &tfjson.SchemaBlock{
				Attributes: map[string]*tfjson.SchemaAttribute{
					exAttr.Name: &exAttr.SchemaAttribute,
				},
			},
			// Make it required
			MinItems: 1,
			MaxItems: 1,
		},
	}
	sb := tfjson.SchemaBlock{
		NestedBlocks: map[string]*tfjson.SchemaBlockType{
			ex.Name: &ex.SchemaBlockType,
		},
	}

	g := newGraph(&sb)
	tu.IsEqual(t, len(g.root.children), 1)
	tu.IsEqual(t, len(g.nodes), 1)

	ac := g.root.children[0]
	tu.IsEqual(t, len(ac.attributes), 1)
	tu.IsEqual(t, ex.Name, ac.name)
	tu.True(t, ac.isRequired, "ac.isRequired")
	tu.True(t, ac.isArg, "ac.isArg")
	acc := ac.attributes[0]
	tu.IsEqual(t, acc.name, exAttr.Name)
	tu.True(
		t,
		exAttr.AttributeType.Equals(acc.ctyType),
		"AttributeType mismatch",
	)
	tu.True(t, acc.isArg, "acc.isArg")
	tu.True(t, acc.isRequired, "acc.isRequired")
}

func TestGraph_ListNestedAttribute_HostedPages(t *testing.T) {
	js := `
{
	"hosted_pages": {
		"nested_type": {
			"attributes": {
				"content": {
					"type": "string",
					"description": "The conent of the hosted page.",
					"description_kind": "markdown",
					"optional": true
				},
				"hosted_page_id": {
					"type": "string",
					"description_kind": "markdown",
					"required": true
				},
				"locale": {
					"type": "string",
					"description_kind": "markdown",
					"optional": true,
					"computed": true
				},
				"url": {
					"type": "string",
					"description": "The URL for the hosted page.",
					"description_kind": "markdown",
					"required": true
				}
			},
			"nesting_mode": "list"
		},
		"description": "List of hosted pages with their respective attributes",
		"description_kind": "markdown",
		"required": true
	}
}
	`
	attrs := map[string]*tfjson.SchemaAttribute{}
	err := json.Unmarshal([]byte(js), &attrs)
	tu.AssertNoError(t, err, "unmarshalling json")

	newGraph(&tfjson.SchemaBlock{
		Attributes: attrs,
	})
}

func TestGraph_SchemaBlock(t *testing.T) {
	type test struct {
		name    string
		block   func() *tfjson.SchemaBlock
		expNode node
	}
	tests := []test{
		{
			name: "nested_attributes_hosted_pages",
			block: func() *tfjson.SchemaBlock {
				js := `
				{
					"hosted_pages": {
						"nested_type": {
							"attributes": {
								"content": {
									"type": "string",
									"description": "The conent of the hosted page.",
									"description_kind": "markdown",
									"optional": true
								},
								"hosted_page_id": {
									"type": "string",
									"description_kind": "markdown",
									"required": true
								},
								"locale": {
									"type": "string",
									"description_kind": "markdown",
									"optional": true,
									"computed": true
								},
								"url": {
									"type": "string",
									"description": "The URL for the hosted page.",
									"description_kind": "markdown",
									"required": true
								}
							},
							"nesting_mode": "list"
						},
						"description": "List of hosted pages with their respective attributes",
						"description_kind": "markdown",
						"required": true
					}
				}`
				attrs := map[string]*tfjson.SchemaAttribute{}
				err := json.Unmarshal([]byte(js), &attrs)
				tu.AssertNoError(t, err, "unmarshalling json")

				return &tfjson.SchemaBlock{
					Attributes: attrs,
				}
			},
			expNode: node{
				name:       "hosted_pages",
				path:       []string{},
				uniqueName: "hosted_pages",
				attributes: []*attribute{
					{
						name:       "content",
						ctyType:    cty.String,
						isArg:      true,
						isRequired: false,
					},
					{
						name:       "hosted_page_id",
						ctyType:    cty.String,
						isArg:      true,
						isRequired: true,
					},
					{
						name:       "locale",
						ctyType:    cty.String,
						isArg:      true,
						isRequired: false,
					},
					{
						name:       "url",
						ctyType:    cty.String,
						isArg:      true,
						isRequired: true,
					},
				},
				isAttribute: true,
				isArg:       true,
				isRequired:  false,
				nestingPath: []nodeNestingMode{nodeNestingModeList},
				receiver:    "hp",
			},
		},
		{
			name: "nested_attributes_scopes",
			block: func() *tfjson.SchemaBlock {
				js := `
				{
					"scopes": {
						"nested_type": {
							"attributes": {
								"recommended": {
									"type": "bool",
									"description": "Indicates if the scope is recommended.",
									"description_kind": "markdown",
									"optional": true,
									"computed": true
								},
								"required": {
									"type": "bool",
									"description": "Indicates if the scope is required.",
									"description_kind": "markdown",
									"optional": true,
									"computed": true
								},
								"scope_name": {
									"type": "string",
									"description": "The name of the scope, e.g., ` + "`openid`" + `, ` + "`profile`" + `.",
									"description_kind": "markdown",
									"optional": true
								}
							},
							"nesting_mode": "list"
						},
						"description": "List of scopes of the provider with details",
						"description_kind": "markdown",
						"optional": true
					}
				}`
				attrs := map[string]*tfjson.SchemaAttribute{}
				err := json.Unmarshal([]byte(js), &attrs)
				tu.AssertNoError(t, err, "unmarshalling json")

				return &tfjson.SchemaBlock{
					Attributes: attrs,
				}
			},
			expNode: node{
				name:       "scopes",
				path:       []string{},
				uniqueName: "scopes",
				attributes: []*attribute{
					{
						name:       "recommended",
						ctyType:    cty.Bool,
						isArg:      true,
						isRequired: false,
					},
					{
						name:       "required",
						ctyType:    cty.Bool,
						isArg:      true,
						isRequired: false,
					},
					{
						name:       "scope_name",
						ctyType:    cty.String,
						isArg:      true,
						isRequired: false,
					},
				},
				isAttribute: true,
				isArg:       true,
				isRequired:  false,
				nestingPath: []nodeNestingMode{nodeNestingModeList},
				receiver:    "s",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr := newGraph(tt.block())
			tu.AssertEqual(t, len(gr.root.children), 1)
			actualNode := gr.root.children[0]
			if diff := tu.Diff(*actualNode, tt.expNode); diff != "" {
				t.Fatal(tu.Callers(), diff)
			}
		})
	}
}
