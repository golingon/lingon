// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terragen

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/golingon/lingon/pkg/internal/terrajen"
	"golang.org/x/tools/txtar"

	tfjson "github.com/hashicorp/terraform-json"
)

var (
	ErrPackageLocationNotEmpty = errors.New(
		"providers pkg location is not empty",
	)
	ErrProviderSchemaNotFound = errors.New("provider schema not found")
)

type GenerateGoArgs struct {
	ProviderName    string
	ProviderSource  string
	ProviderVersion string
	// OutDir is the filesystem location where the generated files will be
	// created.
	OutDir string
	// PkgPath is the Go pkg path to the generated files location, specified by
	// OutDir. E.g. if OutDir is in a module called "my-module" in a directory
	// called "gen",
	// then the PkgPath should be "my-module/gen".
	PkgPath string
	// Force enables overriding any existing generated files per-provider.
	Force bool
	// Clean enables cleaning the generated files location before generating the
	// new files.
	Clean bool
}

// GenerateGoCode generates Go code for creating Terraform objects for the given
// providers and their schemas.
func GenerateGoCode(
	args GenerateGoArgs,
	providerSchema *tfjson.ProviderSchema,
) error {
	if args.OutDir == "" {
		return errors.New("outDir is empty")
	}
	if args.PkgPath == "" {
		return errors.New("pkgPath is empty")
	}

	slog.Info(
		"Generating Go wrapper",
		slog.String("provider", args.ProviderName),
		slog.String("source", args.ProviderSource),
		slog.String("version", args.ProviderVersion),
	)
	providerGenerator := terrajen.ProviderGenerator{
		GoProviderPkgPath:        args.PkgPath,
		GeneratedPackageLocation: args.OutDir,
		ProviderName:             args.ProviderName,
		ProviderSource:           args.ProviderSource,
		ProviderVersion:          args.ProviderVersion,
	}

	arch, err := generateProviderTxtar(providerGenerator, providerSchema)
	if err != nil {
		return err
	}
	if err := createDirIfNotEmpty(args.OutDir, args.Force, args.Clean); err != nil {
		return fmt.Errorf(
			"creating providers pkg directory %q: %w",
			args.OutDir,
			err,
		)
	}
	// Write the txtar archive to the filesystem.
	if err := writeTxtarArchive(arch); err != nil {
		return fmt.Errorf("writing txtar archive: %w", err)
	}

	return nil
}

func writeTxtarArchive(ar *txtar.Archive) error {
	for _, file := range ar.Files {
		dir := filepath.Dir(file.Name)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("creating directory %q: %w", dir, err)
		}
		if err := os.WriteFile(file.Name, file.Data, 0o644); err != nil {
			return fmt.Errorf("writing file %q: %w", file.Name, err)
		}
	}
	return nil
}

func generateProviderTxtar(
	provider terrajen.ProviderGenerator,
	schema *tfjson.ProviderSchema,
) (*txtar.Archive, error) {
	ar := txtar.Archive{}

	//
	// Generate Provider
	//
	providerSchema := provider.SchemaProvider(schema.ConfigSchema.Block)
	providerFile := terrajen.ProviderFile(providerSchema)
	providerBuf := bytes.Buffer{}
	if err := providerFile.Render(&providerBuf); err != nil {
		terrajen.JenDebug(err)
		return nil, fmt.Errorf("rendering provider file: %w", err)
	}
	ar.Files = append(ar.Files, txtar.File{
		Name: providerSchema.FilePath,
		Data: providerBuf.Bytes(),
	})

	subPkgFile, ok := terrajen.SubPkgFile(providerSchema)
	if ok {
		subPkgBuf := bytes.Buffer{}
		if err := subPkgFile.Render(&subPkgBuf); err != nil {
			terrajen.JenDebug(err)
			return nil, fmt.Errorf("rendering sub package file: %w", err)
		}
		ar.Files = append(ar.Files, txtar.File{
			Name: providerSchema.SubPkgPath,
			Data: subPkgBuf.Bytes(),
		})
	}
	//
	// Generate Resources
	//
	for name, resource := range schema.ResourceSchemas {
		resourceSchema := provider.SchemaResource(name, resource.Block)
		rsf := terrajen.ResourceFile(resourceSchema)
		resourceBuf := bytes.Buffer{}
		if err := rsf.Render(&resourceBuf); err != nil {
			terrajen.JenDebug(err)
			return nil, fmt.Errorf("rendering resource file: %w", err)
		}
		ar.Files = append(ar.Files, txtar.File{
			Name: resourceSchema.FilePath,
			Data: resourceBuf.Bytes(),
		})

		rsSubPkgFile, ok := terrajen.SubPkgFile(resourceSchema)
		if !ok {
			continue
		}
		rsSubPkgBuf := bytes.Buffer{}
		if err := rsSubPkgFile.Render(&rsSubPkgBuf); err != nil {
			terrajen.JenDebug(err)
			return nil, fmt.Errorf("rendering sub package file: %w", err)
		}
		ar.Files = append(ar.Files, txtar.File{
			Name: resourceSchema.SubPkgPath,
			Data: rsSubPkgBuf.Bytes(),
		})
	}

	//
	// Generate Data blocks
	//
	for name, data := range schema.DataSourceSchemas {
		dataSchema := provider.SchemaData(name, data.Block)
		df := terrajen.DataSourceFile(dataSchema)
		dataBuf := bytes.Buffer{}
		if err := df.Render(&dataBuf); err != nil {
			terrajen.JenDebug(err)
			return nil, fmt.Errorf("rendering data file: %w", err)
		}
		ar.Files = append(ar.Files, txtar.File{
			Name: dataSchema.FilePath,
			Data: dataBuf.Bytes(),
		})

		dataSubPkgFile, ok := terrajen.SubPkgFile(dataSchema)
		if !ok {
			continue
		}
		dataSubPkgBuf := bytes.Buffer{}
		if err := dataSubPkgFile.Render(&dataSubPkgBuf); err != nil {
			terrajen.JenDebug(err)
			return nil, fmt.Errorf("rendering sub package file: %w", err)
		}
		ar.Files = append(ar.Files, txtar.File{
			Name: dataSchema.SubPkgPath,
			Data: dataSubPkgBuf.Bytes(),
		})
	}
	return &ar, nil
}

func createDirIfNotEmpty(path string, force, clean bool) error {
	f, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		// Create the directory
		if err := os.MkdirAll(
			path,
			os.ModePerm,
		); err != nil {
			return err
		}
		return nil
	}

	_, readErr := f.Readdirnames(1)
	if readErr != nil {
		if readErr == io.EOF {
			return nil
		}
		return err
	}
	// The directory is not empty. If force or clean flags are not provided, we
	// have a problem.
	if !force && !clean {
		return ErrPackageLocationNotEmpty
	}
	if !clean {
		return nil
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("cleaning directory: %w", err)
	}
	// Create the directory again now that it's gone
	return createDirIfNotEmpty(path, false, false)
}

// ParseProvider takes a provider as a string and returns a Provider object.
// An error is returned if the string could not be parsed.
// Example provider: aws=hashicorp/aws:4.60.0
func ParseProvider(s string) (Provider, error) {
	pMap := strings.SplitN(s, "=", 2)
	if len(pMap) == 1 {
		return Provider{}, fmt.Errorf("provider format incorrect: missing `=`")
	}
	p := Provider{
		Name: pMap[0],
	}
	sourceVersion := strings.SplitN(pMap[1], ":", 2)
	if len(sourceVersion) == 1 {
		return Provider{}, fmt.Errorf(
			"provider format incorrect: missing `:` in `source:version`",
		)
	}
	p.Source = sourceVersion[0]
	p.Version = sourceVersion[1]
	// Add the provider to the map
	return p, nil
}
