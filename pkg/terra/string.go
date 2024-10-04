// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terra

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

func String(s string) StringValue {
	return StringValue{
		isInit: true,
		isRef:  false,
		value:  cty.StringVal(s),
	}
}

func ReferenceAsString(ref Reference) StringValue {
	return StringValue{
		isInit: true,
		isRef:  true,
		ref:    ref,
	}
}

func mustReferenceTokens(refr Referencer) string {
	ref, err := refr.InternalRef()
	if err != nil {
		panic(
			fmt.Sprintf(
				"non-reference value: %s",
				err.Error(),
			),
		)
	}
	toks, err := ref.InternalTokens()
	if err != nil {
		panic(fmt.Sprintf("getting internal tokens: %s", err.Error()))
	}
	return string(toks.Bytes())
}

// StringFormat can be used for string interpolation.
// It takes a string with %s placeholders and replaces those with the refs.
// E.g. given a variable `attr` which references an attribute
// `resource.id.attr`:
//
//	StringFormat("Hello ${%s}", attr) -> "Hello ${resource.id.attr}"
func StringFormat(s string, refs ...Referencer) StringValue {
	strRefs := make([]string, len(refs))
	for i, ref := range refs {
		strRefs[i] = mustReferenceTokens(ref)
	}
	// Slice must of type []interface{} to be used in fmt.Sprintf.
	anyRefs := make([]interface{}, len(strRefs))
	for i, ref := range strRefs {
		anyRefs[i] = ref
	}
	return String(fmt.Sprintf(s, anyRefs...))
}

var _ Value[StringValue] = (*StringValue)(nil)

type StringValue struct {
	isInit bool
	isRef  bool
	ref    Reference

	value cty.Value
}

func (v StringValue) AsBool() BoolValue {
	if v.isRef {
		return ReferenceAsBool(v.ref)
	}
	val, err := convert.Convert(v.value, cty.Bool)
	if err != nil {
		panic(fmt.Sprintf("converting string to bool: %s", err.Error()))
	}
	return BoolValue{
		value: val,
	}
}

func (v StringValue) AsNumber() NumberValue {
	if v.isRef {
		return ReferenceAsNumber(v.ref)
	}
	val, err := convert.Convert(v.value, cty.Number)
	if err != nil {
		panic(fmt.Sprintf("converting string to bool: %s", err.Error()))
	}
	return NumberValue{
		isInit: true,
		value:  val,
	}
}

func (v StringValue) InternalTokens() (hclwrite.Tokens, error) {
	if !v.isInit {
		return nil, nil
	}
	if v.isRef {
		return v.ref.InternalTokens()
	}
	// We need to support string interpolation, and using
	// hclwrite.TokensForValue(v.value) will escape `${` to `$${`.
	// Instead, we marshal the string value to JSON, remove the surrounding
	// quotes and convert it to hclwrite.Tokens.
	bStr, err := json.Marshal(v.value.AsString())
	if err != nil {
		return nil, fmt.Errorf("marshalling string value: %w", err)
	}
	unquotedStr := bStr[1 : len(bStr)-1]
	toks := hclwrite.Tokens{}
	toks = append(toks, &hclwrite.Token{
		Type:  hclsyntax.TokenOQuote,
		Bytes: []byte{'"'},
	})
	if len(unquotedStr) > 0 {
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenQuotedLit,
			Bytes: unquotedStr,
		})
	}
	toks = append(toks, &hclwrite.Token{
		Type:  hclsyntax.TokenCQuote,
		Bytes: []byte{'"'},
	})
	return toks, nil
}

func (v StringValue) InternalRef() (Reference, error) {
	if !v.isRef {
		return Reference{},
			errors.New("StringValue: cannot use value as reference")
	}
	return v.ref.copy(), nil
}

func (v StringValue) InternalWithRef(ref Reference) StringValue {
	return ReferenceAsString(ref)
}
