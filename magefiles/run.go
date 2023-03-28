// Copyright (c) Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build mage

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/magefile/mage/mg"
	"github.com/volvo-cars/lingon/magefiles/notice"
	"golang.org/x/exp/slog"
)

const (
	// OSVScanner is the OSV Scanner to find vulnarabilities
	osvScannerRepo    = "github.com/google/osv-scanner/cmd/osv-scanner"
	osvScannerVersion = "@v1"

	// vuln is the GoVulnCheck to find vulnarabilities
	vulnRepo    = "golang.org/x/vuln/cmd/govulncheck"
	vulnVersion = "@latest" // TODO: check version in go.mod

	// syft is the Syft to generate SBOM
	syftRepo    = "github.com/anchore/syft/cmd/syft"
	syftVersion = "@latest" // TODO: check version in go.mod

	// goLicenses is Google's go-licenses to export all licenses
	goLicensesRepo    = "github.com/google/go-licenses"
	goLicensesVersion = "@latest" // TODO: check version in go.mod

	// goCILint is golangci/golangci-lint to lint code
	goCILintRepo    = "github.com/golangci/golangci-lint/cmd/golangci-lint"
	goCILintVersion = "@latest" // TODO: check version in go.mod

	// copyWriteCheck is hashicorp/copywrite to check license headers
	copyWriteRepo    = "github.com/hashicorp/copywrite"
	copyWriteVersion = "@latest" // TODO: check version in go.mod

	// goFumpt is mvdan.cc/gofumpt to format code
	goFumptRepo    = "mvdan.cc/gofumpt"
	goFumptVersion = "@latest"
)

// Run is the namespace for running all checks
type Run mg.Namespace

// AllParallel is used to run all checks in parallel
func (Run) AllParallel() error {
	slog.Info("Running all checks")
	mg.Deps(
		Tidy,
		Run.GoVulnCheck,
		Run.Syft,
		Run.OSVScanner,
		Run.GoFumpt,
		Run.CreateNotice,
		Run.CopyWriteCheck,
		Run.GoCILint,
	)
	return nil
}

// Fix is used to run all fixes sequentially
func (Run) Fix() error {
	slog.Info("Running all fixes")
	mg.SerialDeps(
		Tidy,
		Run.GoFumpt,
		Run.CopyWriteFix,
		Run.GoCILint,
	)
	return nil
}

func (Run) Scan() error {
	slog.Info("Running all scans")
	mg.SerialDeps(
		Run.Syft,
		Run.GoVulnCheck,
		Run.OSVScanner,
		Run.CreateNotice,
		Run.GoLicensesCheck,
		Run.CopyWriteCheck,
	)
	return nil
}

// OSVScanner is the OSV Scanner to find vulnerabilities
func (Run) OSVScanner() error {
	slog.Info("Running OSV Scanner")
	return goRun(
		osvScannerRepo+osvScannerVersion,
		"-r", ".",
	)
}

// GoVulnCheck is the GoVulnCheck to find vulnerabilities
func (Run) GoVulnCheck() error {
	slog.Info("Running govulncheck")
	return goRun(vulnRepo+vulnVersion, "./...")
}

var onlyOnce sync.Once

// Syft is used to generate SBOM
func (Run) Syft() error {
	slog.Info("Running syft - generating SBOM")
	defer slog.Info("DONE syft - generating SBOM")
	var err error
	onlyOnce.Do(
		func() {
			err = exec.Command(
				"go", "run",
				syftRepo+syftVersion,
				"packages",
				"dir:.",
				"-o=spdx-json",
				"--file=bin/sbom.json",
			).Run()
		},
	)
	return err
}

// GoFumptCheck is used to format code
func (Run) GoFumptCheck() error {
	slog.Info("Running gofumpt - formatting code")
	return goRun(goFumptRepo+goFumptVersion, "-l", "-extra", ".")
}

// GoFumpt is used to format code
func (Run) GoFumpt() error {
	slog.Info("Running gofumpt - formatting code")
	return goRun(goFumptRepo+goFumptVersion, "-w", "-extra", ".")
}

// CreateNotice is used to generate a NOTICE file
func (Run) CreateNotice() error {
	slog.Info("Running go-licenses - generating report")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(
		"go", "run",
		goLicensesRepo+goLicensesVersion,
		"report",
		"./...",
		"--template=./magefiles/notice/licenses.tpl",
		"--ignore=github.com/volvo-cars/lingon",
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

	slog.Info("Generating NOTICE")
	noticeFile, err := os.Create("NOTICE")
	if err != nil {
		return fmt.Errorf("creating NOTICE file: %w", err)
	}

	if err := notice.GenerateNotice(noticeFile, &stdout); err != nil {
		return fmt.Errorf("generating NOTICE file: %w", err)
	}
	return nil
}

// GoLicensesCheck is used to export all licenses
func (Run) GoLicensesCheck() error {
	slog.Info("Running go-licenses - exporting licenses")
	return goRun(
		goLicensesRepo+goLicensesVersion,
		"check ./...",
	)
}

// CopyWriteCheck is used to check license headers
func (Run) CopyWriteCheck() error {
	slog.Info("Running copywrite - checking license headers")
	return goRun(
		copyWriteRepo+copyWriteVersion,
		"headers",
		"--dirPath", "./",
		"--config", "./.copywrite.hcl",
		"--plan",
	)
}

// CopyWriteFix is used to check license headers
func (Run) CopyWriteFix() error {
	slog.Info("Running copywrite - fixing license headers")
	return goRun(
		copyWriteRepo+copyWriteVersion,
		"headers",
		"--dirPath", "./",
		"--config", "./.copywrite.hcl",
	)
}

// GoCILint is used to lint code
func (Run) GoCILint() error {
	slog.Info("Running golangci-lint - linting code")
	return goRun(goCILintRepo+goCILintVersion, "-v", "run", "./...")
}

// Tidy runs go mod tidy
func Tidy() error {
	slog.Info("Running go mod tidy")
	return g0("mod", "tidy")
}

func toStringSlice(x ...any) []string {
	var args []string
	for _, arg := range x {
		switch t := arg.(type) {
		case string:
			if t != "" {
				args = append(args, t)
			}
		case []string:
			if t != nil {
				args = append(args, t...)
			}
		default:
			panic("not a string or []string")
		}
	}

	return args
}
