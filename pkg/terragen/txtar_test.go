// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terragen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tu "github.com/golingon/lingon/pkg/testutil"
	tfjson "github.com/hashicorp/terraform-json"
	"golang.org/x/tools/txtar"
)

type TxtarConfig struct {
	Provider ProviderConfig `json:"provider"`
}

type ProviderConfig struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Version string `json:"version"`
}

func TestTxtar(t *testing.T) {
	txtarFiles, err := filepath.Glob("./testdata/*.txtar")
	tu.AssertNoError(t, err, "globbing testdata/*.txtar")
	for _, txf := range txtarFiles {
		ar, err := txtar.ParseFile(txf)
		tu.AssertNoError(t, err, "parsing txtar file")

		t.Run(
			txf, func(t *testing.T) {
				wd := filepath.Join("testdata", "out", filepath.Base(txf))
				if err := RunTest(wd, ar); err != nil {
					t.Error(err)
					return
				}
				exp, err := os.ReadFile(filepath.Join(wd, "expected.tf"))
				tu.AssertNoError(t, err, "reading expected.tf file")
				act, err := os.ReadFile(filepath.Join(wd, "out", "main.tf"))
				tu.AssertNoError(t, err, "reading out/main.tf file")
				tu.AssertEqual(t, string(act), string(exp))
			},
		)
	}
}

func RunTest(wd string, ar *txtar.Archive) error {
	var cfg TxtarConfig
	if err := json.NewDecoder(bytes.NewReader(ar.Comment)).Decode(&cfg); err != nil {
		return fmt.Errorf("decoding txtar comment: %w", err)
	}

	if err := os.MkdirAll(wd, os.ModePerm); err != nil {
		return fmt.Errorf("creating working directory: %s: %w", wd, err)
	}

	// Write txtar files to directory
	for _, f := range ar.Files {
		if err := os.WriteFile(
			filepath.Join(wd, f.Name),
			f.Data,
			os.ModePerm,
		); err != nil {
			return fmt.Errorf("writing txtar file: %s: %w", f.Name, err)
		}
	}

	// Write go.mod file
	goMod, err := os.ReadFile("../../go.mod")
	if err != nil {
		return fmt.Errorf("reading root go.mod file: %w", err)
	}
	goModStr := strings.Replace(
		string(goMod), "module github.com/golingon/lingon",
		"module test", 1,
	)
	goModStr += "\nreplace github.com/golingon/lingon => ../../../../../\n"
	if err := os.WriteFile(
		filepath.Join(wd, "go.mod"),
		[]byte(goModStr),
		os.ModePerm,
	); err != nil {
		return fmt.Errorf("writing go.mod file: %w", err)
	}

	sch, err := os.Open(filepath.Join(wd, "schema.json"))
	if err != nil {
		return fmt.Errorf("opening schema.json file: %w", err)
	}
	var ps tfjson.ProviderSchemas
	if err := json.NewDecoder(sch).Decode(&ps); err != nil {
		return fmt.Errorf("decoding schema.json file: %w", err)
	}
	genArgs := GenerateGoArgs{
		ProviderName:    cfg.Provider.Name,
		ProviderSource:  cfg.Provider.Source,
		ProviderVersion: cfg.Provider.Version,
		OutDir:          filepath.Join(wd, "out", cfg.Provider.Name),
		PkgPath:         fmt.Sprintf("test/out/%s", cfg.Provider.Name),
		Force:           true,
	}
	if err := GenerateGoCode(genArgs, &ps); err != nil {
		return fmt.Errorf("generating Go code: %w", err)
	}

	{
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = wd
		b, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(b))
			return fmt.Errorf("executing: %s: %w", cmd.String(), err)
		}
	}
	{
		cmd := exec.Command("go", "run", ".")
		cmd.Dir = wd
		b, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(b))
			return fmt.Errorf("executing: %s: %w", cmd.String(), err)
		}
	}
	return nil
}
