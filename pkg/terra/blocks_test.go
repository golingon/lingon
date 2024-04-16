// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terra

import (
	"io"
	"testing"

	tu "github.com/golingon/lingon/pkg/testutil"
)

func TestExtractBlocks_Simple(t *testing.T) {
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
	sb, err := objectsFromStack(&st)
	tu.AssertNoError(t, err)
	tu.IsEqual(t, len(sb.Resources), 1)
	tu.IsEqual[Resource](t, dr, sb.Resources[0])
	tu.IsEqual(t, len(sb.DataSources), 1)
	tu.IsEqual[DataSource](t, ddr, sb.DataSources[0])
}

func TestExtractBlocks_Complex(t *testing.T) {
	type DummyModule struct {
		Resource *dummyResource   `validate:"required"`
		Data     *dummyDataSource `validate:"required"`
	}
	type complexStack struct {
		DummyStack
		DummyModule
		SliceRes []*dummyResource    `validate:"required,dive,required"`
		OneRes   [1]*dummyResource   `validate:"required,dive,required"`
		OneData  [1]*dummyDataSource `validate:"required,dive,required"`
	}
	dr := &dummyResource{}
	ddr := &dummyDataSource{}
	st := complexStack{
		DummyStack: newDummyBaseStack(),
		DummyModule: DummyModule{
			Resource: dr,
			Data:     ddr,
		},
		SliceRes: []*dummyResource{dr, dr},
		OneRes:   [1]*dummyResource{dr},
		OneData:  [1]*dummyDataSource{ddr},
	}
	sb, err := objectsFromStack(&st)
	tu.AssertNoError(t, err)
	tu.IsEqual(t, len(sb.Resources), 4)
	tu.IsEqual(t, len(sb.DataSources), 2)
}

func TestExtractBlocks_UnknownField(t *testing.T) {
	type simpleStack struct {
		DummyStack
		UnknownString string
	}
	st := simpleStack{
		DummyStack: newDummyBaseStack(),
	}
	_, err := objectsFromStack(&st)
	tu.ErrorIs(t, err, ErrUnknownPublicField)
}

func TestExtractBlocks_PrivateField(t *testing.T) {
	type simpleStack struct {
		DummyStack
		privateField string //nolint:structcheck,unused
	}
	st := simpleStack{
		DummyStack: newDummyBaseStack(),
	}
	_, err := objectsFromStack(&st)
	tu.ErrorIs(t, err, ErrNotExportedField)
}

func TestExtractBlocks_EmbedPointer(t *testing.T) {
	type Composition struct{}
	type simpleStack struct {
		DummyStack
		*Composition
	}
	st := simpleStack{
		DummyStack:  newDummyBaseStack(),
		Composition: &Composition{},
	}
	_, err := objectsFromStack(&st)
	tu.AssertNoError(t, err)
}

func newDummyBaseStack() DummyStack {
	return DummyStack{
		Backend:  &dummyBackend{},
		Provider: &dummyProvider{},
	}
}

func TestExtractBlocks_IgnoredField(t *testing.T) {
	type simpleStack struct {
		DummyStack
		UnknownString string `lingon:"-"`
	}
	st := simpleStack{
		DummyStack: newDummyBaseStack(),
	}
	_, err := objectsFromStack(&st)
	tu.IsNil(t, err)
}

type DummyStack struct {
	Stack
	Backend  *dummyBackend
	Provider *dummyProvider
}

//
// Dummy Backend
//

var _ Backend = (*dummyBackend)(nil)

type dummyBackend struct{}

func (b dummyBackend) BackendType() string {
	return "dummy"
}

//
// Dummy Provider
//

var _ Provider = (*dummyProvider)(nil)

type dummyProvider struct{}

func (p *dummyProvider) LocalName() string {
	return "dummy"
}

func (p *dummyProvider) Source() string {
	return "dummy"
}

func (p *dummyProvider) Version() string {
	return "dummy"
}

func (p *dummyProvider) Configuration() interface{} {
	return dummyConfig
}

//
// Dummy Resources
//

var _ Resource = (*dummyResource)(nil)

type dummyResource struct{}

func (r *dummyResource) Type() string {
	return "dummy"
}

func (r *dummyResource) LocalName() string {
	return "dummy"
}

func (r *dummyResource) Configuration() interface{} {
	return dummyConfig
}

func (r *dummyResource) ImportState(av io.Reader) error {
	return nil
}

func (r *dummyResource) Dependencies() Dependencies {
	return nil
}

func (r *dummyResource) LifecycleManagement() *Lifecycle {
	return nil
}

//
// Dummy Data Resources
//

var _ DataSource = (*dummyDataSource)(nil)

type dummyDataSource struct{}

func (d *dummyDataSource) DataSource() string {
	return "dummy"
}

func (d *dummyDataSource) LocalName() string {
	return "dummy"
}

func (d *dummyDataSource) Configuration() interface{} {
	return dummyConfig
}

//
// Dummy Args / Configuration
//

var dummyConfig = dummyArgs{Name: "dummy"}

type dummyArgs struct {
	Name string `hcl:"name,attr"`
}
