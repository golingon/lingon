package sylt

import (
	"context"
	"errors"
)

var (
	ErrMissingActionName = errors.New("missing action name")
	ErrMissingActionType = errors.New("missing stack type")
)

// Actioner defines the interface for actions.
// Actions should be designed to be runnable independently, or as part of a sylt
// workflow.
//
// For example, see [TerraAction] for deploying terra stacks.
// Custom actions can be defined.
// The benefit to this is a consistent way to manage (i.e. create, update,
// destroy) resources.
// For those things that don't fit into terra or kube, just define an action.
type Actioner interface {
	// ActionName is used to identify an action.
	// It also helps with logging and debugging.
	// It is called ActionName so that a struct can use a field member "Name".
	ActionName() string
	// ActionType is the type of the action.
	// It is called ActionType to be consistent with ActionName.
	// All actions must define an action type.
	ActionType() ActionType
	// Run the action.
	Run(ctx context.Context, opts RunOpts) error
	// Cleanup the action.
	// It is used to cleanup after the action.
	// It is only called if destroy is true.
	Cleanup(ctx context.Context, opts RunOpts) error
}

type ActionType string

type RunOpts struct {
	DryRun  bool
	Destroy bool
}
