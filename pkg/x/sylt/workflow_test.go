package sylt_test

import (
	"context"
	"testing"

	"github.com/golingon/lingon/pkg/testutil"
	"github.com/golingon/lingon/pkg/x/sylt"
)

var _ sylt.Actioner = (*dummyAction)(nil)

type dummyAction struct {
	name      string
	runFn     func(context.Context, sylt.RunOpts) error
	cleanupFn func(context.Context, sylt.RunOpts) error
}

func (d *dummyAction) ActionName() string {
	return d.name
}

func (d *dummyAction) ActionType() sylt.ActionType {
	return sylt.ActionType("dummy")
}

func (d *dummyAction) Cleanup(ctx context.Context, opts sylt.RunOpts) error {
	if d.cleanupFn != nil {
		return d.cleanupFn(ctx, opts)
	}
	return nil
}

func (d *dummyAction) Run(ctx context.Context, opts sylt.RunOpts) error {
	if d.runFn != nil {
		return d.runFn(ctx, opts)
	}
	return nil
}

func TestWorkflow(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		opts     []sylt.WorkflowOption
		expCalls []string
	}{
		{
			name: "dry run not destroy",
			opts: []sylt.WorkflowOption{
				sylt.WithWorkflowDryRun(true),
				sylt.WithWorkflowDestroy(false),
			},
			expCalls: []string{"run"},
		},
		{
			name: "dry run destroy",
			opts: []sylt.WorkflowOption{
				sylt.WithWorkflowDryRun(true),
				sylt.WithWorkflowDestroy(true),
			},
			expCalls: []string{"run", "cleanup"},
		},
		{
			name: "not dry run not destroy",
			opts: []sylt.WorkflowOption{
				sylt.WithWorkflowDryRun(false),
				sylt.WithWorkflowDestroy(false),
			},
			expCalls: []string{"run"},
		},
		{
			name: "not dry run destroy",
			opts: []sylt.WorkflowOption{
				sylt.WithWorkflowDryRun(false),
				sylt.WithWorkflowDestroy(true),
			},
			expCalls: []string{"run", "cleanup"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wf := sylt.NewWorkflow()
			calls := []string{}
			act := dummyAction{
				name: "test",
				runFn: func(ctx context.Context, opts sylt.RunOpts) error {
					calls = append(calls, "run")
					return nil
				},
				cleanupFn: func(ctx context.Context, opts sylt.RunOpts) error {
					calls = append(calls, "cleanup")
					return nil
				},
			}
			func() {
				defer func() {
					err := wf.Cleanup(ctx)
					testutil.AssertNoError(t, err)
				}()
				err := wf.Run(ctx, &act)
				testutil.AssertNoError(t, err)
			}()
			testutil.AssertEqualSlice(t, []string{"run"}, calls)
		})
	}
}
