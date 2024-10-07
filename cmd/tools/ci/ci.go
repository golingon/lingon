// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"dario.cat/mergo"
	wf "github.com/golingon/lingon/pkg/workflow"
)

const (
	// OSVScanner is the OSV Scanner to find vulnerabilities
	osvScannerRepo    = "github.com/google/osv-scanner/cmd/osv-scanner"
	osvScannerVersion = "@v1.8.2"
	osvScanner        = osvScannerRepo + osvScannerVersion
	osvScannerBin     = "osv-scanner"

	// goVuln to find vulnerabilities
	vulnRepo    = "golang.org/x/vuln/cmd/govulncheck"
	vulnVersion = "@latest"
	goVuln      = vulnRepo + vulnVersion
	goVulnBin   = "govulncheck"

	// goCILint is for linting code
	goCILintRepo    = "github.com/golangci/golangci-lint/cmd/golangci-lint"
	goCILintVersion = "@v1.60.3"
	goCILint        = goCILintRepo + goCILintVersion
	goCILintBin     = "golangci-lint"

	// goFumpt is mvdan.cc/gofumpt to format code
	goFumptRepo    = "mvdan.cc/gofumpt"
	goFumptVersion = "@v0.7.0"
	goFumpt        = goFumptRepo + goFumptVersion
	goFumptBin     = "gofumpt"

	dirK8s   = "./docs/kubernetes"
	dirTerra = "./docs/terraform"
)

var binz = map[string]string{
	osvScanner: osvScannerBin,
	goVuln:     goVulnBin,
	goCILint:   goCILintBin,
	goFumpt:    goFumptBin,
}

type Result struct {
	Tasks []Task
}

func (r Result) Output() string {
	if len(r.Tasks) > 0 {
		return r.Tasks[len(r.Tasks)-1].Output
	}
	return "none"
}

func (r Result) String() string {
	if len(r.Tasks) == 0 {
		return "none"
	}
	return fmt.Sprintf("tasks:[..., %s]", r.Tasks[len(r.Tasks)-1])
}

type Task struct {
	Cmd    string
	Output string
	Dir    string
}

func (t Task) String() string {
	return fmt.Sprintf("Task{Dir:'%s', Command: '%s'}", t.Dir, t.Cmd)
}

var errUnrecoverable = errors.New("unrecoverable")

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				a = slog.Attr{Key: "time", Value: slog.StringValue(t.Format("15:04:05"))}
			}
			return a
		},
	})).With("ci", "lingon")
	if err := Main(logger); err != nil {
		logger.Error("main", "err", err)
		os.Exit(1)
	}
}

func Main(logger *slog.Logger) error {
	var cover, lint, generate, examples, nodiff, pr, scan, release, update, V bool
	flag.BoolVar(&cover, "cover", false, "tests with coverage")
	flag.BoolVar(&lint, "lint", false, "linting and formatting code (gofumpt, golangci-lint)")
	flag.BoolVar(&generate, "generate", false, "generate all docs and readme")
	flag.BoolVar(&examples, "examples", false, "generate and tests all docs examples")
	flag.BoolVar(&nodiff, "nodiff", false, "error if git diff is not empty")
	flag.BoolVar(&pr, "pr", false, "run pull request checks: lint + go test + examples /!\\")
	flag.BoolVar(&scan, "scan", false, "scan for vulnerabilities")
	flag.BoolVar(&release, "release", false, "create a new release")
	flag.BoolVar(&update, "update", false, "update dependencies")
	flag.BoolVar(&V, "verbose", false, "verbose logging")

	flag.Parse()

	genargs := []string{"generate", "./..."}
	vargs := ""
	if V {
		genargs = slices.Insert(genargs, 1, "-v", "-x")
		vargs = "-v"
	}

	mid := []wf.Middleware[Result]{
		wf.StartTimeInCtxMiddleware[Result](),
		ErrorMiddleware[Result](func(err error) bool { return errors.Is(err, errUnrecoverable) }),
		LoggerMiddleware[Result](logger),
	}
	p := wf.NewPipeline(mid...)

	if update {
		p.Steps = append(p.Steps, p.Series(
			run(V, ".", "go", "get", "-u", "./..."),
			run(V, ".", "go", "mod", "tidy"),
			run(V, dirK8s, "go", "get", "-u", "./..."),
			run(V, dirK8s, "go", "mod", "tidy"),
			run(V, dirTerra, "go", "get", "-u", "./..."),
			run(V, dirTerra, "go", "mod", "tidy"),
		))
	}

	if cover {
		coverOut := "cover.out"
		p.Steps = append(p.Steps, p.Series(
			run(V, ".", "go", "test", "-coverprofile="+coverOut, "-covermode=count", "./pkg/..."),
			run(V, ".", "go", "tool", "cover", "-func="+coverOut),
			wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
				var buf bytes.Buffer
				if err := coverPct(&buf, coverOut); err != nil {
					return r, err
				}
				pct, err := parseCovPct(&buf)
				if err != nil {
					return r, err
				}
				s, err := badgetpl(pct)
				if err != nil {
					return r, err
				}
				if err := writeBadge(".github/coverage.svg", []byte(s)); err != nil {
					return r, err
				}
				return r, nil
			}),
		))
	}

	if lint {
		p.Steps = append(p.Steps, p.Series(
			run(V, ".", "go", "mod", "tidy"),
			run(V, ".", "go", "run", goFumpt, "-w", "-extra", "."),
			run(V, ".", "go", "run", goCILint, vargs, "run", "./..."),
		))
	}

	if generate {
		p.Steps = append(p.Steps, p.Parallel(wf.MergeTransform[Result](mergo.WithAppendSlice),
			p.Series(
				run(V, ".", "go", genargs...),
				run(V, ".", "go", "mod", "tidy"),
			),
			p.Series(
				run(V, dirK8s, "go", genargs...),
				run(V, dirK8s, "go", "mod", "tidy"),
			),
			p.Series(
				run(V, dirTerra, "go", genargs...),
				run(V, dirTerra, "go", "mod", "tidy"),
			),
		))
	}

	if examples {
		p.Steps = append(p.Steps, p.Parallel(wf.MergeTransform[Result](mergo.WithAppendSlice),
			p.Series(
				run(V, dirK8s, "go", "mod", "tidy"),
				run(V, dirK8s, "go", genargs...),
				run(V, dirK8s, "go", "test", "-mod=readonly", vargs, "./..."),
			),
			p.Series(
				run(V, dirTerra, "go", "mod", "tidy"),
				run(V, dirTerra, "go", genargs...),
				run(V, dirTerra, "go", "test", "-mod=readonly", vargs, "./..."),
			),
		))
	}

	if pr {
		p.Steps = append(p.Steps, p.Series(
			run(V, ".", "go", genargs...),
			run(V, ".", "go", "test", vargs, "./..."),
			run(V, ".", "go", "mod", "tidy"),
			// FIXME: causing too many issues right now
			// run(V, dirK8s, "go", genargs...),
			// run(V, dirK8s, "go", "mod", "tidy"),
			// run(V, dirTerra, "go", genargs...),
			// run(V, dirTerra, "go", "mod", "tidy"),
			installRun(V, goFumpt, "-w", "-extra", "."),
			installRun(V, goCILint, vargs, "run", "./..."),
		))
	}

	if scan {
		p.Steps = append(p.Steps, p.Series(
			installRun(V, goVuln, "./..."),
			installRun(V, osvScanner, "."),
		))
	}

	if release {
		p.Steps = append(p.Steps, p.Series(
			run(V, ".", "git", "rev-parse", "--short", "HEAD"),
			wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
				prev := r.Tasks[len(r.Tasks)-1]
				ssha := strings.ReplaceAll(prev.Output, "\n", "")
				d := time.Now().UTC().Format("2006-01-02")
				v := d + "-" + ssha
				r.Tasks = append(r.Tasks, Task{Output: v})
				return r, nil
			}),
			wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
				prev := r.Tasks[len(r.Tasks)-1]
				cmd := exec.Command("git", "tag", "-a", prev.Output, "-s", "-m", "Release "+prev.Output)
				o, err := cmd.CombinedOutput()
				if err != nil {
					return r, err
				}
				r.Tasks = append(r.Tasks, Task{Cmd: cmd.String(), Output: string(o)})
				return r, nil
			}),
		))
	}

	// should be last
	if nodiff {
		p.Steps = append(p.Steps, p.Series(
			run(V, ".", "git", "--no-pager", "diff"),
			wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
				prev := r.Tasks[len(r.Tasks)-1]
				if len(prev.Output) != 0 {
					return r, fmt.Errorf("changes detected: %w", errUnrecoverable)
				}
				return r, nil
			}),
			wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
				if len(r.Tasks) == 0 {
					r.Tasks = append(r.Tasks, Task{Cmd: "print", Output: "no previous tasks, nothing to print."})
					return r, nil
				}
				prev := r.Tasks[len(r.Tasks)-1]
				fmt.Println(prev.Output)
				return r, nil
			}),
		))
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer cancel()

	result := &Result{Tasks: []Task{}}
	resp, err := p.Run(ctx, result)
	if err != nil {
		return err
	}

	logger.Info("pipeline done", "result", resp)
	return nil
}

type CLI struct {
	Dir        string
	Bin        string
	Args       []string
	ShowOutput bool
}

func run(verbose bool, dir, bin string, args ...string) *CLI {
	return &CLI{Dir: dir, Bin: bin, Args: args, ShowOutput: verbose}
}

func (g *CLI) String() string {
	return fmt.Sprintf("CLI{Dir: %s, Command: '%s'}", g.Dir, g.Bin+" "+strings.Join(g.Args, " "))
}

func installRun(verbose bool, bin string, args ...string) wf.Step[Result] {
	cli, ok := binz[bin]
	if !ok {
		// not a tool we use
		return run(verbose, ".", "go", append([]string{"run", bin}, args...)...)
	}
	if _, err := exec.LookPath(cli); err != nil {
		// tool not found => installing
		fmt.Println("install tool", bin)
		instErr := exec.Command("go", "install", bin).Run()
		if instErr != nil {
			return run(verbose, ".", "go", append([]string{"run", bin}, args...)...)
		}
	}
	if _, err := exec.LookPath(cli); err != nil {
		// tool not in the path
		fmt.Println("tool not in the path", bin)
		return run(verbose, ".", "go", append([]string{"run", bin}, args...)...)
	}
	fmt.Println("running local tool", cli)
	return run(verbose, ".", cli, args...)
}

func (g *CLI) Run(ctx context.Context, r *Result) (*Result, error) {
	if g.Bin == "" {
		return r, fmt.Errorf("%T: binary not set: %w", g, errUnrecoverable)
	}
	cmd := exec.CommandContext(ctx, g.Bin, g.Args...)
	var buf strings.Builder
	if g.ShowOutput {
		cmd.Stdout = io.MultiWriter(&buf, os.Stdout)
		cmd.Stderr = io.MultiWriter(&buf, os.Stderr)
	} else {
		cmd.Stdout = &buf
		cmd.Stderr = &buf
	}
	if g.Dir != "" {
		cmd.Dir = g.Dir
	}
	err := cmd.Run()
	r.Tasks = append(r.Tasks, Task{Cmd: cmd.String(), Dir: g.Dir, Output: buf.String()})
	if err != nil {
		err = fmt.Errorf("%s: %s: %w", cmd.String(), err, errUnrecoverable)
	}
	return r, err
}

func LoggerMiddleware[T any](l *slog.Logger) wf.Middleware[T] {
	return func(next wf.Step[T]) wf.Step[T] {
		return wf.MidFunc[T](func(ctx context.Context, res *T) (*T, error) {
			name := wf.Name(next)
			if name != "MidFunc" {
				id, _ := wf.GetStepID(ctx)
				l.Info("start", "Type", name, "id", id, "STEP", next)
			}

			resp, err := next.Run(ctx, res)

			if name != "MidFunc" {
				id, _ := wf.GetStepID(ctx)
				t, errctx := wf.GetStepStartTime(ctx)
				if errors.Is(errctx, wf.ErrMissingFromContext) {
					l.Info("done", "Type", name, "id", id, "Result", fmt.Sprintf("%v", resp))
				} else {
					l.Info("done", "Type", name, "id", id, "duration", time.Since(t),
						"Result", fmt.Sprintf("%v", resp))
				}
			}
			return resp, err
		})
	}
}

type Outputer interface {
	Output() string
}

func ErrorMiddleware[T Outputer](h func(error) bool) wf.Middleware[T] {
	return func(next wf.Step[T]) wf.Step[T] {
		return wf.MidFunc[T](func(ctx context.Context, r *T) (*T, error) {
			resp, err := next.Run(ctx, r)
			if errors.Is(ctx.Err(), context.Canceled) {
				return resp, fmt.Errorf("%s: %v", (*resp).Output(), ctx.Err())
			}
			if h(err) {
				// In case of an error, show what output of the failing step.
				if o := (*resp).Output(); o != "" {
					fmt.Printf("\noutput:\n\n %s\n", o)
				}
				return resp, fmt.Errorf("err: %v", err)
			}
			return resp, err
		})
	}
}

// Badge

//go:embed badge.svg.tpl
var tplFS embed.FS

const (
	brightgreen = "#4c1"
	green       = "#97ca00"
	yellow      = "#dfb317"
	yellowgreen = "#a4a61d"
	orange      = "#fe7d37"
	red         = "#e05d44"
)

func badgetpl(pct float64) (string, error) {
	color := ""
	switch {
	case pct < 50:
		color = orange
	case pct < 60:
		color = yellow
	case pct < 70:
		color = yellowgreen
	case pct < 80:
		color = green
	case pct < 90:
		color = brightgreen
	default:
		color = red
	}
	t := template.Must(template.ParseFS(tplFS, "badge.svg.tpl"))
	data := struct {
		Color      string
		Percentage float64
	}{Percentage: pct, Color: color}

	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeBadge(path string, b []byte) error {
	_, err := os.Stat(filepath.Dir(path))
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(path), 0o755)
	}
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()
	_, err = f.Write(b)
	return err
}

func parseCovPct(r io.Reader) (float64, error) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		t := s.Text()
		if strings.HasPrefix(t, "total:") {
			f := strings.Fields(t)
			fmt.Println("FOUND IT ", f)
			if f[0] == "total:" && len(f) == 3 {
				return strconv.ParseFloat(strings.TrimRight(f[2], "%"), 64)
			}

			break
		}
	}
	if err := s.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "scan:", err)
	}
	return 0.0, fmt.Errorf("no coverage found")
}

func coverPct(w io.Writer, o string) error {
	cmd := exec.Command("go", "tool", "cover", "-func="+o) //nolint:gosec
	slog.Info("exec", slog.String("cmd", cmd.String()))
	defer slog.Info("done", slog.String("cmd", cmd.String()))

	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Stderr.Sync()
		return fmt.Errorf("%q: %s", cmd.String(), err)
	}
	return nil
}
