// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

const (
	signOffDependabot = "Signed-off-by: dependabot[bot] <support@github.com>"
)

func Update() {
	fmt.Println("⤴️ update deps")
	iferr(Go("get", "-u", recDir))
	iferr(Go("mod", "tidy"))
	fmt.Println("⤴️ update deps docs")
	docRun(DocKubernetes, "go", "get", "-u", recDir)
	docRun(DocTerraform, "go", "get", "-u", recDir)
	docRun(DocKubernetes, "go", "mod", "tidy")
	docRun(DocTerraform, "go", "mod", "tidy")
	fmt.Println("✅ update deps done")
}

func isDependabot() bool {
	cmd := exec.Command("git", "log", "-1", "--pretty=format:'%b'")
	slog.Info("exec", slog.String("cmd", cmd.String()))
	o, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("git check commit", "err", err, "output", string(o))
		return false
	}
	res := strings.Contains(string(o), signOffDependabot)
	return res
}

func listModifiedFiles() ([]string, error) {
	cmd := exec.Command(
		"git",
		"diff-tree", "--no-commit-id", "--name-only", "HEAD", "-r",
		// "show", "--name-only", `--pretty="" `, "HEAD",
	)
	slog.Info("exec", slog.String("cmd", cmd.String()))
	o, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("listModifiedFiles: %s, %w", string(o), err)
	}
	split := strings.Split(string(o), "\n")
	res := []string{}
	for _, s := range split {
		if s != "" {
			res = append(res, s)
		}
	}
	return res, nil
}
