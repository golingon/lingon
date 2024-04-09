// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terra

import (
	"fmt"
	"testing"

	tkihcl "github.com/golingon/lingon/pkg/internal/hcl"
	tu "github.com/golingon/lingon/pkg/testutil"
)

func TestReferenceCopy(t *testing.T) {
	ref := ReferenceResource(&dummyResource{})
	ref2 := ref.copy()
	tu.IsEqual(t, ref.underlyingType, ref2.underlyingType)
	tu.IsEqual(t, ref.res, ref2.res)
	tu.IsEqual(t, ref.data, ref2.data)
}

func TestReferenceTokens(t *testing.T) {
	ref := ReferenceResource(&dummyResource{})
	tu.IsEqual(t, "dummy.dummy", testTokensOrError(t, ref))

	appendRef := ref.Append("abc")
	tu.IsEqual(t, "dummy.dummy.abc", testTokensOrError(t, appendRef))

	keyRef := ref.key("a")
	tu.IsEqual(t, "dummy.dummy[\"a\"]", testTokensOrError(t, keyRef))

	indexRef := ref.index(3)
	tu.IsEqual(t, "dummy.dummy[3]", testTokensOrError(t, indexRef))

	splatRef := ref.splat()
	tu.IsEqual(t, "dummy.dummy[*]", testTokensOrError(t, splatRef))

	// 	Check original has not been updated
	tu.IsEqual(t, "dummy.dummy", testTokensOrError(t, ref))
}

func testTokensOrError(t *testing.T, value tkihcl.Tokenizer) string {
	toks, err := value.InternalTokens()
	if err != nil {
		t.Errorf("getting tokens: %s", err)
		t.Fail()
	}
	return string(toks.Bytes())
}

func exampleTokensOrError(value tkihcl.Tokenizer) string {
	toks, err := value.InternalTokens()
	if err != nil {
		return fmt.Sprintf("ERROR: getting tokens: %s", err)
	}
	return string(toks.Bytes())
}
