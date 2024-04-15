// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terra

import (
	"bytes"
	"testing"

	tu "github.com/golingon/lingon/pkg/testutil"
)

func TestExport(t *testing.T) {
	type simpleStack struct {
		DummyStack
		DummyRes  *dummyResource   `validate:"required"`
		DummyData *dummyDataSource `validate:"required"`
	}
	dr := &dummyResource{}
	ddr := &dummyDataSource{}
	st := simpleStack{
		DummyStack: newDummyBaseStack(),
		DummyRes:   dr,
		DummyData:  ddr,
	}
	var b bytes.Buffer
	err := encodeStack(&st, &b)
	tu.IsNil(t, err)
	want := `terraform {
  backend "dummy" {
  }
  required_providers {
    dummy = {
      source  = "dummy"
      version = "dummy"
    }
  }
}

// Provider blocks
provider "dummy" {
  name = "dummy"
}

// Data blocks
data "dummy" "dummy" {
  name = "dummy"
}

// Resource blocks
resource "dummy" "dummy" {
  name = "dummy"
}

`
	tu.AssertEqual(t, want, b.String())
}
