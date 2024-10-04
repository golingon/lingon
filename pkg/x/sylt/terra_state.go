package sylt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/golingon/lingon/pkg/terra"
	tfjson "github.com/hashicorp/terraform-json"
)

var _ error = (*MissingStateError)(nil)

// MissingStateError is returned when the state is missing for resources.
type MissingStateError struct {
	// Resources that are missing state.
	Resources []terra.Resource
}

func (se *MissingStateError) Error() string {
	strRes := make([]string, len(se.Resources))
	for i, res := range se.Resources {
		strRes[i] = fmt.Sprintf("%s.%s", res.Type(), res.LocalName())
	}
	return fmt.Sprintf(
		"missing state for resources: [%s]",
		strings.Join(strRes, ","),
	)
}

type ResourceStater[T any] interface {
	terra.Resource
	State() (T, bool)
}

// RequireResourceState checks if the given [terra.Resource] has state.
// If the resource has state, it is returned.
// If the resource does not have state, an error is added to the given error.
// If the error is a [MissingStateError], the resource is added to the list of
// resources that are missing state.
//
//	var err error
//	res1 := RequireResourceState(stack.Resource1, &err)
//	res2 := RequireResourceState(stack.Resource2, &err)
//	if err != nil {
//		return err
//	}
func RequireResourceState[V any, T ResourceStater[V]](res T, err *error) V {
	if state, ok := res.State(); ok {
		return state
	}
	var emptyState V
	var stateErr *MissingStateError
	if errors.As(*err, &stateErr) {
		stateErr.Resources = append(stateErr.Resources, res)
		return emptyState
	}
	*err = errors.Join(*err, &MissingStateError{
		Resources: []terra.Resource{res},
	})
	return emptyState
}

// StateStatus defines how complete the state is for a stack.
type StateStatus int

const (
	// StateStatusUnknown the state mode has not been determined yet
	// (e.g. no plan/apply).
	StateStatusUnknown StateStatus = 0
	// StateStatusEmpty there is no state (e.g. no apply yet).
	StateStatusEmpty StateStatus = 1
	// StateStatusPartial there is a state, but there are resources in the stack
	// that are not in the state yet (e.g. need to be applied).
	StateStatusPartial StateStatus = 2
	// StateStatusSync all resources in the stack have state and the state does
	// not have resources that are not in the stack.
	StateStatusSync StateStatus = 3
	// StateStatusOverflow all resources in the stack have state but the state
	// has more resources than the stack.
	StateStatusOverflow StateStatus = 4
)

// StackImportState imports the Terraform state into the Terraform Stack.
// A [StateStatus] is returned indicating how complete the state of the
// resources is.
func StackImportState(
	stack terra.Exporter,
	state *tfjson.State,
) (StateStatus, error) {
	sb, err := terra.ObjectsFromStack(stack)
	if err != nil {
		return StateStatusUnknown, fmt.Errorf("getting stack objects: %w", err)
	}
	// Note: ideally we would always set the state to nil for each resource
	// before importing the current state.
	// However, there is currently no east way to do this with the current
	// [terra.Resource] interface.
	// Probably better not to reuse the same stack for multiple runs anyway,
	// so this is a bit of an edge case.
	if state.Values == nil || len(state.Values.RootModule.Resources) == 0 {
		return StateStatusEmpty, nil
	}
	isFullState := true
	stateResources := state.Values.RootModule.Resources
	// Iterate over the resources in the Stack and try to find the corresponding
	// resource in the state.
	// If it exists, import the state into the Stack.
	for _, res := range sb.Resources {
		resFound := false
		for _, sr := range stateResources {
			// Find the resource in the state. It is the same resource if the
			// resource type and resource local name match because that is how
			// Terraform uniquely identifies resources in its state.
			if res.Type() == sr.Type && res.LocalName() == sr.Name {
				resFound = true
				var b bytes.Buffer
				if err := json.NewEncoder(&b).Encode(sr.AttributeValues); err != nil {
					return StateStatusUnknown, fmt.Errorf(
						"encoding attribute values for resource %s.%s: %w",
						res.Type(), res.LocalName(), err,
					)
				}
				if err := res.ImportState(&b); err != nil {
					return StateStatusUnknown, fmt.Errorf(
						"importing state into resource %s.%s: %w",
						res.Type(), res.LocalName(), err,
					)
				}
				break
			}
		}
		if !resFound {
			isFullState = false
		}
	}
	if isFullState {
		// If all the stack resources have state, check that the state does not
		// have more resources than the stack. If it does, it means that the
		// state has resources that are not in the stack.
		if len(stateResources) > len(sb.Resources) {
			return StateStatusOverflow, nil
		}
		return StateStatusSync, nil
	}
	return StateStatusPartial, nil
}
