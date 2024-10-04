package sylt

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
)

var (
	terraCallInit = []string{"init", "-upgrade"}
	terraCallPlan = []string{
		"plan",
		"-detailed-exitcode", // Return exit code 2 if there are changes.
		"-out=" + terraPlanFile,
	}
	terraCallPlanDestroy = []string{
		"plan",
		"-detailed-exitcode", // Return exit code 2 if there are changes.
		"-out=" + terraPlanFile,
		"-destroy",
	}
	terraCallApply = []string{
		"apply",
		"-input=false",
		terraPlanFile,
	}
	terraCallShowPlan  = []string{"show", "-json", terraPlanFile}
	terraCallShowState = []string{"show", "-json"}
)

// terraCmder is an interface for running Terraform commands.
// It is an interface for testing purposes.
// Maybe in the future we want to have specific methods here, like Init(),
// Plan(), Apply(), etc.
// And maybe then we export it so people can provide their own.
// Keep it private for now.
type terraCmder interface {
	Run(
		ctx context.Context,
		dir string,
		stdout io.Writer,
		stderr io.Writer,
		args ...string,
	) error
}

var _ terraCmder = (*terraCmd)(nil)

type terraCmd struct {
	cmd string
}

func (e *terraCmd) Run(
	ctx context.Context,
	dir string,
	stdout io.Writer,
	stderr io.Writer,
	args ...string,
) error {
	if e.cmd == "" {
		return errors.New("no command set")
	}
	cmd := exec.CommandContext(ctx, e.cmd, args...)
	cmd.Dir = dir
	// Inherit environment variables.
	cmd.Env = os.Environ()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
