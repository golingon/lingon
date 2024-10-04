package sylt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/golingon/lingon/pkg/terra"
	tu "github.com/golingon/lingon/pkg/testutil"
	tfjson "github.com/hashicorp/terraform-json"
)

func TestTerraRun(t *testing.T) {
	type test struct {
		name        string
		wfOpts      []WorkflowOption
		expCallArgs [][]string
	}

	tests := []test{
		{
			name: "default",
			wfOpts: []WorkflowOption{
				WithWorkflowDryRun(false),
				WithWorkflowDestroy(false),
			},
			expCallArgs: [][]string{
				terraCallInit,
				terraCallPlan,
				terraCallShowPlan,
				// How to simulate a plan diff?
				// We cannot create a exec.ExitError, because it's fields are
				// not exported.
				terraCallShowState,
			},
		},
		{
			name: "dry-run",
			wfOpts: []WorkflowOption{
				WithWorkflowDryRun(true),
				WithWorkflowDestroy(false),
			},
			expCallArgs: [][]string{
				terraCallInit,
				terraCallPlan,
				terraCallShowPlan,
				terraCallShowState,
			},
		},
		{
			name: "destroy",
			wfOpts: []WorkflowOption{
				WithWorkflowDryRun(false),
				WithWorkflowDestroy(true),
			},
			expCallArgs: [][]string{
				terraCallInit,
				terraCallShowState,
				terraCallInit,
				terraCallPlanDestroy,
				terraCallShowPlan,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type stack struct {
				terra.Stack
			}
			ctx := context.Background()
			terraRec := &TerraExecRecorder{}
			terraAct := Terra("test", &stack{})
			terraAct.cmd = terraRec

			wf := NewWorkflow(tt.wfOpts...)
			if err := wf.Run(ctx, terraAct); err != nil {
				t.Fatalf("running TerraAction: %s", err)
			}
			if err := wf.Cleanup(ctx); err != nil {
				t.Fatalf("finishing workflow: %s", err)
			}
			// Check the calls to Terraform.
			tu.AssertEqual(t, len(tt.expCallArgs), len(terraRec.calls))
			for i, expArgs := range tt.expCallArgs {
				t.Logf(
					"comparing index %d: %v == %v",
					i,
					expArgs,
					terraRec.calls[i].args,
				)
				tu.AssertEqualSlice(t, expArgs, terraRec.calls[i].args)
			}
		})
	}
}

var _ terraCmder = (*TerraExecRecorder)(nil)

type TerraExecRecorder struct {
	calls []terraCalls
}

// Run implements TerraExecutor.
func (t *TerraExecRecorder) Run(
	ctx context.Context,
	dir string,
	stdout io.Writer,
	stderr io.Writer,
	args ...string,
) error {
	t.calls = append(t.calls, terraCalls{
		dir:  dir,
		args: args,
	})
	// If command is show, then we need to write some JSON.
	if len(args) > 0 && args[0] == "show" {
		// Write some JSON to stdout.
		state := tfjson.State{
			FormatVersion:    "0.1",
			TerraformVersion: "1.0.0",
		}
		enc := json.NewEncoder(stdout)
		if err := enc.Encode(state); err != nil {
			return fmt.Errorf("encoding state: %w", err)
		}
	}
	return nil
}

type terraCalls struct {
	dir  string
	args []string
}
