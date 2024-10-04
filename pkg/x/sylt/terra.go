package sylt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/golingon/lingon/pkg/terra"
	tfjson "github.com/hashicorp/terraform-json"
)

const (
	terraPlanFile = "tfplan"
	terraMD5File  = "tf.md5"
)

const ActionTypeTerra ActionType = "terra"

type TerraOption func(*terraOpts)

func WithTerraCmd(cmd string) TerraOption {
	return func(o *terraOpts) {
		o.cmd = cmd
	}
}

type terraOpts struct {
	enableCache bool
	cmd         string
}

var defaultTerraOpts = func() terraOpts {
	return terraOpts{
		enableCache: false,
		cmd:         "tofu",
	}
}

// Terra creates a new TerraAction.
func Terra[T terra.Exporter](
	name string,
	stack T,
	opts ...TerraOption,
) *TerraAction[T] {
	opt := defaultTerraOpts()
	for _, o := range opts {
		o(&opt)
	}

	cmd := terraCmd{
		cmd: opt.cmd,
	}

	act := TerraAction[T]{
		Name:  name,
		Stack: stack,
		opts:  opt,
		cmd:   &cmd,
	}
	act.log = slog.With(
		"action_name",
		name,
		"action_type",
		ActionTypeTerra,
		"dir",
		act.dir(),
	)
	return &act
}

var _ Actioner = (*TerraAction[*terra.Stack])(nil)

// TerraAction is an action that performs terra commands on a stack.
// It implements the [Actioner] interface so can be used together with a
// [Workflow] or independently.
// Use the [Terra] function to create a TerraAction.
type TerraAction[T terra.Exporter] struct {
	// Name is the name of the action.
	// It is used to create a directory to store the terraform files.
	// The directory will be `.lingon/terra/<name>`.
	Name string
	// Stack is the terra stack to export (to hcl) and run terra commands on.
	Stack T

	opts terraOpts

	log         *slog.Logger
	cmd         terraCmder
	stateStatus StateStatus
	plan        *plan
}

func (a *TerraAction[T]) ActionName() string {
	return a.Name
}

func (a *TerraAction[T]) ActionType() ActionType {
	return ActionTypeTerra
}

func (a *TerraAction[T]) Run(ctx context.Context, opts RunOpts) error {
	if a.Name == "" {
		return ErrMissingActionName
	}

	runLog := a.log.With("run_opts", opts)

	runLog.Info("exporting stack")
	if err := a.Export(); err != nil {
		return err
	}

	var isCached bool
	if a.opts.enableCache {
		var err error
		isCached, err = a.compareLocalHash()
		if err != nil {
			return fmt.Errorf("checking local cache: %w", err)
		}
	}

	// Skip terraform init if the stack is cached.
	if !isCached {
		runLog.Info("initialising stack")
		if err := a.Init(ctx); err != nil {
			return fmt.Errorf(
				"initializing stack %s: %w",
				a.Name, err,
			)
		}
	}

	// If the action is marked for destruction, skip the plan and apply.
	// We only need to show the state.
	if opts.Destroy {
		runLog.Info("importing state into stack")
		if err := a.ImportState(ctx); err != nil {
			return fmt.Errorf(
				"getting state for stack %s: %w",
				a.Name, err,
			)
		}
		return nil
	}

	// Skip plan if the stack is cached.
	var diff bool
	if !isCached {
		runLog.Info("planning stack", "diff", diff)
		var err error
		diff, err = a.Plan(ctx)
		if err != nil {
			return fmt.Errorf(
				"planning stack %s: %w", a.Name, err,
			)
		}
		runLog.Info("planned stack", "diff", diff)
	}
	// Apply if there is a diff AND not dry run.
	if diff && !opts.DryRun {
		runLog.Info("applying stack")
		if err := a.Apply(ctx); err != nil {
			return fmt.Errorf(
				"applying stack %s: %w", a.Name, err,
			)
		}
	}

	runLog.Info("importing state into stack")
	if err := a.ImportState(ctx); err != nil {
		return fmt.Errorf(
			"getting state for stack %s: %w",
			a.Name, err,
		)
	}
	// Only write the md5 checksum if the stack has no changes and caching is
	// enabled.
	if !a.HasChanges() && a.opts.enableCache {
		if err := a.writeLocalHash(); err != nil {
			return fmt.Errorf("writing local cache: %w", err)
		}
	}
	runLog.Info("run finished", "has_changes", a.HasChanges())
	return nil
}

// Cleanup destroys the stack if the destroy option is set.
// It honours the dry run option.
func (a *TerraAction[T]) Cleanup(ctx context.Context, opts RunOpts) error {
	a.log.Info("running cleanup")
	if err := a.Init(ctx); err != nil {
		return fmt.Errorf(
			"initializing %s: %w",
			a.ActionName(), err,
		)
	}
	diff, err := a.PlanDestroy(ctx)
	if err != nil {
		return fmt.Errorf(
			"planning stack %s: %w", a.ActionName(), err,
		)
	}
	if opts.DryRun || !diff {
		return nil
	}
	// Apply the stack with above plan which passed the destroy flag.
	if err := a.Apply(ctx); err != nil {
		return fmt.Errorf(
			"destroying stack %s: %w", a.ActionName(), err,
		)
	}
	if err := a.ImportState(ctx); err != nil {
		return fmt.Errorf(
			"importing state for stack %s: %w",
			a.Name, err,
		)
	}
	// Cleanup the local cache (if exists), as the stack has been destroyed.
	cachePath := filepath.Join(a.dir(), terraMD5File)
	if _, err := os.Stat(cachePath); err == nil {
		if err := os.Remove(cachePath); err != nil {
			return fmt.Errorf("removing local cache: %w", err)
		}
	}
	return nil
}

// HasChanges returns true if the plan (if any) has no diff, or if we ran apply
// on
// the stack, and if the objects in the stack all have some state.
// If there is no plan, but the state is full, the stack is considered in sync.
// Being in sync means there is no drift.
// This is best effort: things can always change between the time terra plan and
// apply were run.
func (a *TerraAction[T]) HasChanges() bool {
	// TODO: what about if the state has things that are not in the stack??
	if a.stateStatus != StateStatusSync {
		return true
	}
	// If there is no plan, but the state is full, the best judgement is to say
	// that we are in sync.
	// This is primarily for the case where we are running destroy on the
	// workflow and we don't want to stop, because we don't run a plan during
	// destroy.
	if a.plan == nil {
		return false
	}
	if a.plan.diff() {
		return true
	}
	return false
}

// Export exports the stack to HCL.
func (a *TerraAction[T]) Export() error {
	if err := terra.Export(
		a.Stack,
		terra.WithExportOutputDirectory(a.dir()),
	); err != nil {
		return fmt.Errorf("exporting stack: %w", err)
	}
	return nil
}

// Apply runs the terra apply command.
func (a *TerraAction[T]) Apply(ctx context.Context) error {
	if err := a.cmd.Run(ctx, a.dir(), os.Stdout, os.Stderr, terraCallApply...); err != nil {
		return fmt.Errorf("running apply command: %w", err)
	}
	// Mark the plan as applied.
	a.plan.isApplied = true

	return nil
}

// Plan runs the terra plan command and imports the plan into the stack.
func (a *TerraAction[T]) Plan(
	ctx context.Context,
) (bool, error) {
	return a.planWithDestroy(ctx, false)
}

// PlanDestroy runs the terra plan command with the destroy flag and imports the
// plan into the stack.
func (a *TerraAction[T]) PlanDestroy(
	ctx context.Context,
) (bool, error) {
	return a.planWithDestroy(ctx, true)
}

func (a *TerraAction[T]) planWithDestroy(
	ctx context.Context,
	destroy bool,
) (bool, error) {
	planArgs := terraCallPlan
	if destroy {
		planArgs = terraCallPlanDestroy
	}
	doPlan := func() (bool, error) {
		if err := a.cmd.Run(ctx, a.dir(), os.Stdout, os.Stderr, planArgs...); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				// When passing the -detailed-exitcode flag to the plan command,
				// exit code 2 means no errors but there is a diff.
				// https://developer.hashicorp.com/terraform/cli/commands/plan#detailed-exitcode
				if exitError.ExitCode() == 2 {
					return true, nil
				}
			}
			return false, fmt.Errorf("running plan command: %w", err)
		}
		return false, nil
	}
	diff, err := doPlan()
	if err != nil {
		return false, err
	}

	if err := a.showPlan(ctx); err != nil {
		return false, fmt.Errorf("showing plan: %w", err)
	}

	return diff, nil
}

// Init runs the terra init command.
func (a *TerraAction[T]) Init(ctx context.Context) error {
	out := bytes.Buffer{}
	if err := a.cmd.Run(ctx, a.dir(), &out, &out, terraCallInit...); err != nil {
		fmt.Fprint(os.Stderr, out.String())
		return fmt.Errorf("running init command: %w", err)
	}
	return nil
}

func (a *TerraAction[T]) showPlan(ctx context.Context) error {
	var buf bytes.Buffer
	if err := a.cmd.Run(ctx, a.dir(), &buf, os.Stderr, terraCallShowPlan...); err != nil {
		return fmt.Errorf("running show command: %w", err)
	}

	var tfPlan tfjson.Plan
	dec := json.NewDecoder(&buf)
	dec.UseNumber()
	if err := dec.Decode(&tfPlan); err != nil {
		return fmt.Errorf("decoding plan JSON: %w", err)
	}
	if err := tfPlan.Validate(); err != nil {
		return fmt.Errorf("validating plan: %w", err)
	}
	a.plan = &plan{
		out:       &tfPlan,
		isApplied: false,
	}

	return nil
}

// ImportState runs `terra show` and imports the state into the stack.
func (a *TerraAction[T]) ImportState(ctx context.Context) error {
	var buf bytes.Buffer
	if err := a.cmd.Run(ctx, a.dir(), &buf, os.Stderr, "show", "-json"); err != nil {
		return fmt.Errorf("running show command: %w", err)
	}

	var tfState tfjson.State
	dec := json.NewDecoder(&buf)
	dec.UseNumber()
	if err := dec.Decode(&tfState); err != nil {
		return fmt.Errorf("decoding state JSON: %w", err)
	}
	if err := tfState.Validate(); err != nil {
		return fmt.Errorf("validating state: %w", err)
	}

	if err := a.importStateIntoStack(&tfState); err != nil {
		return fmt.Errorf("importing stack state: %w", err)
	}

	return nil
}

func (a *TerraAction[T]) dir() string {
	return filepath.Join(
		".lingon",
		"terra",
		a.Name,
	)
}

// importStateIntoStack imports the given state into the stack.
func (a *TerraAction[T]) importStateIntoStack(state *tfjson.State) error {
	stateStatus, err := StackImportState(a.Stack, state)
	if err != nil {
		return fmt.Errorf("importing state: %w", err)
	}
	a.stateStatus = stateStatus
	return nil
}

// compareLocalHash checks if the stack has been run before.
func (a *TerraAction[T]) compareLocalHash() (bool, error) {
	return false, errors.New("unimplemented")
}

// writeLocalHash writes the hash of the given stack.
func (a *TerraAction[T]) writeLocalHash() error {
	return errors.New("unimplemented")
}
