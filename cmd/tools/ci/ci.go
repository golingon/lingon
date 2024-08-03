// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	// OSVScanner is the OSV Scanner to find vulnerabilities
	osvScannerRepo    = "github.com/google/osv-scanner/cmd/osv-scanner"
	osvScannerVersion = "@v1.8.2"

	// goVuln to find vulnerabilities
	vulnRepo    = "golang.org/x/vuln/cmd/govulncheck"
	vulnVersion = "@latest"
	goVuln      = vulnRepo + vulnVersion

	// goLicenses is Google's go-licenses to export all licenses
	goLicensesRepo    = "github.com/google/go-licenses"
	goLicensesVersion = "@v1.6.0"
	goLicenses        = goLicensesRepo + goLicensesVersion

	// goCILint is for linting code
	goCILintRepo    = "github.com/golangci/golangci-lint/cmd/golangci-lint"
	goCILintVersion = "@v1.59.1"
	goCILint        = goCILintRepo + goCILintVersion

	// goFumpt is mvdan.cc/gofumpt to format code
	goFumptRepo    = "mvdan.cc/gofumpt"
	goFumptVersion = "@v0.6.0"
	goFumpt        = goFumptRepo + goFumptVersion
)

const (
	curDir = "."
	recDir = "./..."
)

var mod = func() string {
	if isUpdatingGoModOnly() {
		slog.Info("updating go.mod")
		return "-mod=mod"
	} else {
		slog.Info("readonly go.mod")
		return "-mod=readonly"
	}
}()

var verbose bool

func main() {
	var cover, lint, doc, examples, fix, nodiff, pr, scan, release, update bool
	flag.BoolVar(&cover, "cover", false, "tests with coverage")
	flag.BoolVar(&lint, "lint", false, "linting and formatting code (gofumpt, golangci-lint)")
	flag.BoolVar(&doc, "doc", false, "generate all docs and readme")
	flag.BoolVar(&examples, "examples", false, "generate and tests all docs examples /!\\ slow without build cache")
	flag.BoolVar(&fix, "fix", false, "same as -lint + generating notice and licenses headers")
	flag.BoolVar(&nodiff, "nodiff", false, "error if git diff is not empty")
	flag.BoolVar(&pr, "pr", false, "run pull request checks: lint + notice + go test + examples /!\\")
	flag.BoolVar(&scan, "scan", false, "scan for vulnerabilities")
	flag.BoolVar(&release, "release", false, "create a new release")
	flag.BoolVar(&update, "update", false, "update dependencies")
	flag.BoolVar(&verbose, "verbose", false, "verbose logging")

	flag.Parse()

	if update {
		Update()
	}
	if cover {
		CoverP()
	}
	if lint {
		Lint()
	}
	if doc {
		DocGen()
	}
	if examples {
		DocExamples()
	}
	if fix {
		Fix()
	}
	if pr {
		PullRequest()
	}
	if scan {
		Scan()
	}
	if release {
		Release()
	}
	if nodiff {
		// should be last
		HasGitDiff()
	}
}

func iferr(err error) {
	if err != nil {
		panic(err)
	}
}

func Release() {
	d := time.Now().UTC().Format("2006-01-02")
	ssha, err := shortSha()
	iferr(err)
	v := d + "-" + ssha
	iferr(TagRelease(v, "Release "+v))
}

func TagRelease(tag, msg string) error {
	cmd := exec.Command("git", "tag", "-a", tag, "-s", "-m", msg)
	slog.Info("exec", slog.String("cmd", cmd.String()))
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	cmdgp := exec.Command("git", "push", "--tags")
	slog.Info("exec", slog.String("cmd", cmdgp.String()))
	_, err = cmdgp.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func shortSha() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	slog.Info("exec", slog.String("cmd", cmd.String()))
	b, err := cmd.CombinedOutput()
	s := strings.ReplaceAll(string(b), "\n", "")
	return s, err
}

func Cover() {
	coverOutput := "cover.out"
	coverMode := "count" // see `go help testflag` for more info
	iferr(
		Go(
			"test",
			recDir,
			"-coverprofile="+coverOutput,
			"-covermode="+coverMode,
		),
	)
	iferr(
		Go(
			"tool",
			"cover",
			"-func="+coverOutput,
			// "-html="+coverOutput,
			// "-o",
			// "cover.html",
		),
	)
	fmt.Println("✅ coverage generated: open cover.html to see results")
}

func Lint() {
	fmt.Println("🧹 code linting")
	iferr(Go("mod", "tidy"))
	iferr(Go("mod", "verify"))
	iferr(GoRun(goFumpt, "-w", "-extra", curDir))
	iferr(GoRun(goCILint, "-v", "run", recDir))
	fmt.Println("✅ code linted")
}

func Fix() {
	Lint()
	fmt.Println("📋 licenses fix")
	iferr(Notice())
	fmt.Println("✅ All fixes applied")
}

func MainBranch() {
	iferr(Go("test", "-v", recDir))
	DocGen()
	fmt.Println("✅ main branch checks passed")
}

func PullRequest() {
	fmt.Println("📝 pull request checks")
	iferr(Go("test", "-v", recDir))
	DocExamples()
	Lint()
	fmt.Println("✅ pull request checks passed")
}

func Scan() {
	iferr(GoRun(goVuln, recDir))
	iferr(OSVScanner())
	fmt.Println("✅ all scans completed")
}

func DocGen() {
	fmt.Println("📝 generating docs")
	argsGen := []string{"go", "generate"}
	if verbose {
		argsGen = append(argsGen, "-v", "-x")
	}
	argsGen = append(argsGen, recDir)

	iferr(Go(argsGen[1:]...))
	docRun(DocKubernetes, argsGen...)
	docRun(DocKubernetes, "go", "mod", "tidy")
	docRun(DocTerraform, argsGen...)
	docRun(DocTerraform, "go", "mod", "tidy")
	fmt.Println("✅ docs generated")
}

func DocExamples() {
	DocGen()
	fmt.Println("📝 testing examples")
	docRun(DocKubernetes, "go", "test", mod, "-v", recDir)
	// no need to recurse the directories
	// as it uses build tags
	docRun(DocTerraform, "go", "test", mod, "-v")
	fmt.Println("✅ docs generated and examples tested")
}

func Go(args ...string) error {
	cmd := exec.Command("go", args...)
	slog.Info("exec go", "cmd", cmd.String())
	defer slog.Info("done exec go", "cmd", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Stderr.Sync()
		_ = os.Stdout.Sync()
		return fmt.Errorf("go: %s", err)
	}
	return nil
}

func GoRun(args ...string) error {
	return Go(append([]string{"run", mod}, args...)...)
}

// HasGitDiff displays the git diff and errors if there is a diff
func HasGitDiff() {
	cmd := exec.Command("git", "--no-pager", "diff")
	slog.Info("exec", slog.String("cmd", cmd.String()))
	b, err := cmd.CombinedOutput()
	iferr(err)
	if len(b) == 0 {
		return
	}
	buf := bytes.NewBuffer(b)
	fmt.Println(buf.String())
	panic("git diff is not empty")
}

type DocFolder int

const (
	DocKubernetes DocFolder = iota + 1
	DocTerraform
)

func (d DocFolder) String() string {
	switch d {
	case DocKubernetes:
		return "./docs/kubernetes"
	case DocTerraform:
		return "./docs/terraform"
	default:
		return fmt.Sprintf("unknown folder: %T", d)
	}
}

// docRun runs a command in the docs directory
func docRun(df DocFolder, args ...string) {
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	slog.Info("docRun", "cmd", cmd.String(), "folder", df)
	defer slog.Info("docRun", "cmd", args[0]+" done", "folder", df)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	switch df {
	case DocKubernetes:
		cmd.Dir = "./docs/kubernetes"
	case DocTerraform:
		cmd.Dir = "./docs/terraform"
	default:
		panic(fmt.Sprintf("unknown folder: %v", df))
	}
	if err := cmd.Run(); err != nil {
		_ = os.Stderr.Sync()
		_ = os.Stdout.Sync()
		panic(err)
	}
}

// OSVScanner is the OSV Scanner to find vulnerabilities
func OSVScanner() error {
	slog.Info("running OSV Scanner")
	defer slog.Info("DONE OSV Scanner")
	// return GoRun(osvScannerRepo+osvScannerVersion, "-r", curDir)
	// not scanning docs/go.mod because of github.com/aws/aws-sdk-go
	// and the osvScanner returns an error when a vulnerability is detected
	return GoRun(osvScannerRepo+osvScannerVersion, curDir)
}

// Notice is used to generate a NOTICE file
func Notice() error {
	slog.Info("running go-licenses - generating report")
	defer slog.Info("DONE go-licenses - generating report")
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(
		"go", "run",
		goLicenses,
		"report",
		recDir,
		"--template=./cmd/tools/ci/licenses.tpl",
		"--ignore=github.com/golingon/lingon",
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	slog.Info("exec", slog.String("cmd", cmd.String()))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"running go-licenses: %w:\n\n%s", err,
			stderr.String(),
		)
	}

	noticeFile, err := os.Create("NOTICE")
	if err != nil {
		return fmt.Errorf("creating NOTICE file: %w", err)
	}

	if err = GenerateNotice(noticeFile, &stdout); err != nil {
		return fmt.Errorf("generating NOTICE file: %w", err)
	}
	return nil
}
