package sylt

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/golingon/lingon/pkg/terra"
	"github.com/golingon/lingon/pkg/testutil"
)

var _ ResourceStater[*struct{}] = (*dummyResource)(nil)

type dummyResource struct {
	state *struct{}
}

func (d *dummyResource) Configuration() interface{} {
	return nil
}

func (d *dummyResource) Dependencies() terra.Dependencies {
	return nil
}

func (d *dummyResource) ImportState(attributes io.Reader) error {
	return errors.New("unimplemented")
}

func (d *dummyResource) LifecycleManagement() *terra.Lifecycle {
	return nil
}

func (d *dummyResource) LocalName() string {
	return "dummy"
}

func (d *dummyResource) State() (*struct{}, bool) {
	if d.state == nil {
		return nil, false
	}
	return d.state, true
}

func (d *dummyResource) Type() string {
	return "dummy"
}

func ExampleRequireResourceState() {
	type MyStack struct {
		Resource1 *dummyResource
		Resource2 *dummyResource
	}
	stack := MyStack{
		Resource1: &dummyResource{
			state: &struct{}{},
		},
		Resource2: &dummyResource{},
	}
	var err error
	res1 := RequireResourceState(stack.Resource1, &err)
	res2 := RequireResourceState(stack.Resource2, &err)
	if err != nil {
		// handle error
	}
	fmt.Println(err.Error())
	fmt.Println(res1 == nil)
	fmt.Println(res2 == nil)
	// Output: missing state for resources: [dummy.dummy]
	// false
	// true
}

func TestMissingError(t *testing.T) {
	t.Run("with state", func(t *testing.T) {
		d := dummyResource{
			state: &struct{}{},
		}
		var err error
		state := RequireResourceState(&d, &err)
		testutil.AssertEqual(t, err, nil)
		testutil.AssertEqual(t, state, &struct{}{})
	})
	t.Run("without state", func(t *testing.T) {
		d := dummyResource{}
		var err error
		state := RequireResourceState(&d, &err)
		testutil.AssertErrorMsg(
			t,
			err,
			"missing state for resources: [dummy.dummy]",
		)
		testutil.AssertEqual(t, state, nil)
	})
	t.Run("without state multiple resources", func(t *testing.T) {
		d := dummyResource{}
		var err error
		_ = RequireResourceState(&d, &err)
		_ = RequireResourceState(&d, &err)
		testutil.AssertErrorMsg(
			t,
			err,
			"missing state for resources: [dummy.dummy,dummy.dummy]",
		)
	})
}
