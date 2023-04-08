// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terrajen

import (
	"fmt"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	tu "github.com/volvo-cars/lingon/pkg/testutil"
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
	tu.IsEqual(t, len(g.attributes), 1)
	ac := g.attributes[0]
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
	tu.IsEqual(t, len(g.attributes), 10)
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
	tu.IsEqual(t, len(g.children), 1)
	tu.IsEqual(t, len(g.nodes), 1)

	ac := g.children[0]
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
