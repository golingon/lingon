// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terra

import (
	"fmt"
)

func ExampleString() {
	s := String("hello world")
	fmt.Println(exampleTokensOrError(s))
	// 	Output: "hello world"
}

func ExampleStringFormat() {
	// Create a dummy resource and create a StringValue reference to it.
	// Ignore this and pretend you have your reference attribute as you
	// normally would, e.g.
	// 	ref.Attributes().Name()
	res := dummyResource{}
	ref := ReferenceAsSingle[StringValue](ReferenceResource(&res))
	// Create a StringValue with a format string and the reference.
	s := StringFormat("${%s}", ref)
	fmt.Println(exampleTokensOrError(s))
	// 	Output: "${dummy.dummy}"
}
