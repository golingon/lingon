package sylt_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/golingon/lingon/pkg/terra"
	tu "github.com/golingon/lingon/pkg/testutil"
	"github.com/golingon/lingon/pkg/x/sylt"
	random "github.com/golingon/terraproviders/random/3.6.1"
	"github.com/golingon/terraproviders/random/3.6.1/random_string"
)

type randomStringStack struct {
	terra.Stack
	Provider     *random.Provider
	RandomString *random_string.Resource
}

func TestTerraWorkflow(t *testing.T) {
	t.Parallel()
	cleanupState := func(t *testing.T) {
		t.Cleanup(func() {
			stateFile := filepath.Join(
				".lingon",
				"terra",
				t.Name(),
				"terraform.tfstate",
			)
			t.Logf("cleanup: remove: %s", stateFile)
			if err := os.RemoveAll(stateFile); err != nil {
				t.Fatalf("cleanup failed: %v", err)
			}
		})
	}
	// This test uses (open)tofu and produces local state files.
	// We need to prune those test files before each run.
	ctx := context.Background()
	newRandomStringStack := func() *randomStringStack {
		return &randomStringStack{
			Provider: &random.Provider{},
			RandomString: &random_string.Resource{
				Name: "random_string",
				Args: random_string.Args{
					Length: terra.Number(8),
				},
			},
		}
	}
	t.Run("dry run not destroy", func(t *testing.T) {
		t.Parallel()
		cleanupState(t)
		wf := sylt.NewWorkflow()
		randomString := sylt.Terra(
			t.Name(),
			newRandomStringStack(),
		)
		err := wf.Run(ctx, randomString)
		tu.AssertNoError(t, err)
		t.Cleanup(func() {
			err := wf.Cleanup(ctx)
			tu.AssertNoError(t, err)
			tu.AssertEqual(t, true, randomString.HasChanges())
		})
	})
	t.Run("not dry run not destroy", func(t *testing.T) {
		t.Parallel()
		cleanupState(t)
		wf := sylt.NewWorkflow(sylt.WithWorkflowDryRun(false))
		randomString := sylt.Terra(
			t.Name(),
			newRandomStringStack(),
		)
		err := wf.Run(ctx, randomString)
		tu.AssertNoError(t, err)
		t.Cleanup(func() {
			err := wf.Cleanup(ctx)
			tu.AssertNoError(t, err)
			tu.AssertEqual(t, false, randomString.HasChanges())
		})
	})
	t.Run("dry run destroy", func(t *testing.T) {
		t.Parallel()
		cleanupState(t)
		wf := sylt.NewWorkflow(sylt.WithWorkflowDestroy(true))
		randomString := sylt.Terra(
			t.Name(),
			newRandomStringStack(),
		)
		err := wf.Run(ctx, randomString)
		tu.AssertNoError(t, err)
		t.Cleanup(func() {
			err := wf.Cleanup(ctx)
			tu.AssertNoError(t, err)
			tu.AssertEqual(t, true, randomString.HasChanges())
		})
	})
	t.Run("multi workflow plan apply destroy", func(t *testing.T) {
		t.Parallel()
		cleanupState(t)
		randomString := sylt.Terra(
			t.Name(),
			newRandomStringStack(),
		)
		t.Run("plan", func(t *testing.T) {
			wf := sylt.NewWorkflow()
			err := wf.Run(ctx, randomString)
			tu.AssertNoError(t, err)
			t.Cleanup(func() {
				err := wf.Cleanup(ctx)
				tu.AssertNoError(t, err)
				tu.AssertEqual(t, true, randomString.HasChanges())
			})
		})
		t.Run("apply", func(t *testing.T) {
			wf := sylt.NewWorkflow(sylt.WithWorkflowDryRun(false))
			err := wf.Run(ctx, randomString)
			tu.AssertNoError(t, err)
			t.Cleanup(func() {
				err := wf.Cleanup(ctx)
				tu.AssertNoError(t, err)
				tu.AssertEqual(t, false, randomString.HasChanges())
			})
		})
		t.Run("destroy", func(t *testing.T) {
			wf := sylt.NewWorkflow(
				sylt.WithWorkflowDryRun(false),
				sylt.WithWorkflowDestroy(true),
			)
			err := wf.Run(ctx, randomString)
			tu.AssertNoError(t, err)
			t.Cleanup(func() {
				err := wf.Cleanup(ctx)
				tu.AssertNoError(t, err)
				tu.AssertEqual(t, true, randomString.HasChanges())
			})
		})
	})
}
