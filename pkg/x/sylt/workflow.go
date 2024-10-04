package sylt

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrDuplicateAction = errors.New("duplicate action name and type")

func NewWorkflow(opts ...WorkflowOption) *Workflow {
	opt := defaultWorkflowOpts()
	for _, o := range opts {
		o(&opt)
	}
	return &Workflow{
		opts: opt,
	}
}

// Workflow runs actions implementing the [Actioner] interface, to perform
// things like deploying terra stacks and kube apps.
// Use the [NewWorkflow] function to create a new workflow.
type Workflow struct {
	opts    workflowOpts
	actions []Actioner
	mu      sync.Mutex
}

type WorkflowOption func(*workflowOpts)

// WithWorkflowDryRun sets the dry run option for Workflow.
func WithWorkflowDryRun(b bool) WorkflowOption {
	return func(o *workflowOpts) { o.DryRun = b }
}

// WithWorkflowDestroy sets the destroy option for Workflow.
func WithWorkflowDestroy(b bool) WorkflowOption {
	return func(o *workflowOpts) { o.Destroy = b }
}

type workflowOpts struct {
	DryRun  bool
	Destroy bool
}

var defaultWorkflowOpts = func() workflowOpts {
	return workflowOpts{
		DryRun:  true,
		Destroy: false,
	}
}

// Run performs the given action.
// It adds the action to the workflow's list of ordered action.
// When Cleanup() is called on the workflow, if the destroy flag is set,
// the actions's Cleanup() function are called in reverse order.
func (w *Workflow) Run(ctx context.Context, action Actioner) error {
	if action.ActionName() == "" {
		return ErrMissingActionName
	}
	if action.ActionType() == "" {
		return ErrMissingActionType
	}
	if err := w.addAction(action); err != nil {
		return fmt.Errorf("adding action to client: %w", err)
	}
	if err := action.Run(ctx, RunOpts{
		DryRun:  w.opts.DryRun,
		Destroy: w.opts.Destroy,
	}); err != nil {
		return fmt.Errorf("running action %s: %w", action.ActionName(), err)
	}
	return nil
}

type CleanupOption func(*cleanupOpts)

type cleanupOpts struct {
	dryRun  bool
	destroy bool
}

// WithCleanupDryRun sets the dry run option for Cleanup.
func WithCleanupDryRun(b bool) CleanupOption {
	return func(o *cleanupOpts) { o.dryRun = b }
}

// WithCleanupDestroy sets the destroy option for Cleanup.
func WithCleanupDestroy(b bool) CleanupOption {
	return func(o *cleanupOpts) { o.destroy = b }
}

// Cleanup should be run after all actions have been processed.
// If the destroy option is set, Cleanup will call destroy on all actions, in
// the
// reverse order they were run.
//
// Calling Cleanup can be easily achieved with a Go defer statement, e.g.
//
//	defer func() {
//		if err := wf.Cleanup(ctx); err != nil {
//			// Handle error.
//		}
//	}()
//
// Or in a test using the t.Cleanup function, e.g.
//
//	t.Cleanup(func() {
//		if err := wf.Cleanup(
//			ctx,
//			sylt.WithCleanupDestroy(true),
//		); err != nil {
//			t.Fatalf("finishing: %s", err)
//		}
//	})
func (w *Workflow) Cleanup(ctx context.Context, opts ...CleanupOption) error {
	fOpts := cleanupOpts{
		dryRun:  w.opts.DryRun,
		destroy: w.opts.Destroy,
	}
	for _, opt := range opts {
		opt(&fOpts)
	}

	if !fOpts.destroy {
		return nil
	}

	// Iterate over actions in reverse.
	for i := len(w.actions) - 1; i >= 0; i-- {
		action := w.actions[i]
		if err := action.Cleanup(ctx, RunOpts{
			DryRun:  fOpts.dryRun,
			Destroy: fOpts.destroy,
		}); err != nil {
			return fmt.Errorf("destroying %s: %w", action.ActionName(), err)
		}
	}
	return nil
}

// addAction appends the given action to the client's list of action.
// This ensures all action name and type pairs are unique.
// The list of actions is used to destroy anything that the actions have
// created.
func (w *Workflow) addAction(action Actioner) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, exAction := range w.actions {
		if exAction.ActionName() == action.ActionName() &&
			exAction.ActionType() == action.ActionType() {
			return fmt.Errorf(
				"%w (name: %s, type: %s)",
				ErrDuplicateAction,
				action.ActionName(),
				action.ActionType(),
			)
		}
	}
	w.actions = append(w.actions, action)
	return nil
}
