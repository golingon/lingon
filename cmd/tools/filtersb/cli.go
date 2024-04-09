// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/golingon/lingon/pkg/terragen"
)

func main() {
	var (
		outDir             string
		providerStr        string
		includeResources   filterMap = map[string]struct{}{}
		includeDataSources filterMap = map[string]struct{}{}
	)

	tfOutDir := filepath.Join(".lingon", "testdata")
	flag.StringVar(&outDir, "out", "", "directory to generate Go files in")
	flag.Var(&includeResources, "include-resources", "resources to include")
	flag.Var(
		&includeDataSources,
		"include-data-sources",
		"data sources to include",
	)
	flag.StringVar(
		&providerStr,
		"provider",
		"",
		"provider to generate Go files for, e.g. aws=hashicorp/aws:4.49.0",
	)
	flag.Parse()

	if outDir == "" {
		slog.Error("-out flag required")
		os.Exit(1)
	}
	provider, err := terragen.ParseProvider(providerStr)
	if err != nil {
		slog.Error("invalid provider", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	slog.Info(
		"Generating Terraform provider schema",
		slog.String("provider", providerStr),
		slog.String("out", tfOutDir),
	)
	ps, err := terragen.GenerateProviderSchema(
		ctx, provider,
	)
	if err != nil {
		slog.Error("generating provider schema", "err", err)
		os.Exit(1)
	}

	slog.Info(
		"Filtering provider schema",
		slog.Any("resources", includeResources),
		slog.Any("data_sources", includeDataSources),
	)
	for _, schema := range ps.Schemas {
		for rName := range schema.ResourceSchemas {
			if _, ok := includeResources[rName]; !ok {
				delete(schema.ResourceSchemas, rName)
			}
		}
		for dName := range schema.DataSourceSchemas {
			if _, ok := includeDataSources[dName]; !ok {
				delete(schema.DataSourceSchemas, dName)
			}
		}
	}

	outFile := filepath.Join(
		outDir, fmt.Sprintf(
			"%s_%s.json", provider.Name,
			provider.Version,
		),
	)
	f, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0o755)
	if err != nil {
		slog.Error("opening out file", err, slog.String("out", outFile))
	}
	if err := json.NewEncoder(f).Encode(ps); err != nil {
		slog.Error("encoding provider schema", err)
		os.Exit(1)
	}
}

var _ flag.Value = (*providerFlag)(nil)

type providerFlag struct {
	LocalName string
	Provider  terragen.Provider
}

func (p *providerFlag) String() string {
	return fmt.Sprintf(
		"%s:%s=%s",
		p.LocalName,
		p.Provider.Source,
		p.Provider.Version,
	)
}

func (p *providerFlag) Set(value string) error {
	pMap := strings.SplitN(value, "=", 2)
	if len(pMap) == 1 {
		return fmt.Errorf("provider format incorrect: missing `=`")
	}
	localName := pMap[0]
	sourceVersion := strings.SplitN(pMap[1], ":", 2)
	if len(sourceVersion) == 1 {
		return fmt.Errorf(
			"provider format incorrect: missing `:` in `source:version`",
		)
	}
	source := sourceVersion[0]
	version := sourceVersion[1]

	p.LocalName = localName
	p.Provider.Source = source
	p.Provider.Version = version
	return nil
}

var _ flag.Value = (*filterMap)(nil)

type filterMap map[string]struct{}

func (f *filterMap) String() string {
	s := make([]string, len(*f))
	for name := range *f {
		s = append(s, name)
	}
	return strings.Join(s, ", ")
}

func (f *filterMap) Set(value string) error {
	(*f)[value] = struct{}{}
	return nil
}
