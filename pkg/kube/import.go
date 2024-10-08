// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/golingon/lingon/pkg/internal/api"
	"github.com/golingon/lingon/pkg/kubeutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultAppName   = "lingon"
	maxOutputHead    = 150
	defaultOutputDir = "out"
	headerComment    = "Code generated by lingon. EDIT AS MUCH AS YOU LIKE."
	kindConfigMap    = "ConfigMap"
)

var (
	ErrIncompatibleOptions = errors.New("incompatible options")
	logErrIgnored          = slog.Bool("ignored", true)
)

// jamel is short for Jennifer and YAML.
// It is used to generate Go code from Kubernetes manifests.
// Not meant to be created directly, use newImporter(opts...) instead.
type jamel struct {
	l                 *slog.Logger
	objectsCode       map[string]*jen.Statement    // obj name => obj as Go struct
	kubeAppStructCode map[string]*jen.Statement    // fieldName => fieldType
	objectsMeta       map[string]kubeutil.Metadata // obj name => obj meta
	// nameFieldVar maps object name to object variable name for the kube.App
	// struct
	nameFieldVar map[string]string
	// crdPkgAlias keeps alias for those CRD packages discovered during the
	// conversion to Go
	crdPkgAlias map[string]string
	// crdCurrent holds the current package containing CRD types,
	// only useful for top level var declaration of CRD types.
	crdCurrent string
	// o is all the importOption
	o importOption // options
	// buf is the buffer, also [io.Writer] all the Go generated code is written
	// to
	buf bytes.Buffer
	// useReader specifies to read the manifests from io.Reader or from files
	useReader bool
	// useWriter specifies to write the generated code to o.ManifestReader or to
	// files
	useWriter bool
}

func Import(opts ...ImportOption) error {
	j, err := newImporter(opts...)
	if err != nil {
		return fmt.Errorf("import options: %w", err)
	}
	if j.o.Verbose {
		j.l.Info(
			"import", slog.Group(
				"opt",
				slog.String("app-name", j.o.AppName),
				slog.String("package-name", j.o.OutputPkgName),
				slog.String("output-dir", j.o.OutputDir),
				slog.Bool("reader", j.useReader),
				slog.Any("manifests", j.o.ManifestFiles),
				slog.Bool("writer", j.useWriter),
				slog.Bool("remove-app-name", j.o.RemoveAppName),
				slog.Bool("group-by-kind", j.o.GroupByKind),
				slog.Bool("add-methods", j.o.AddMethods),
				slog.Bool("redact-secrets", j.o.RedactSecrets),
				slog.Bool("ignore-errors", j.o.IgnoreErrors),
				slog.Bool("clean-up", j.o.CleanUp),
			),
		)
	}
	if j.useWriter {
		return j.render()
	}
	if err = j.save(); err != nil {
		return fmt.Errorf("convert to Go: %w", err)
	}
	return nil
}

func newImporter(opts ...ImportOption) (*jamel, error) {
	j := &jamel{
		l:                 Logger(os.Stderr),
		buf:               bytes.Buffer{},
		objectsCode:       make(map[string]*jen.Statement),
		kubeAppStructCode: make(map[string]*jen.Statement),
		objectsMeta:       make(map[string]kubeutil.Metadata),
		crdCurrent:        "",
		crdPkgAlias:       make(map[string]string),
		nameFieldVar:      make(map[string]string),
		o:                 importDefaultOpts,
		useReader:         false,
		useWriter:         false,
	}

	for _, opt := range opts {
		opt(j)
	}

	if err := j.gatekeeperImportOptions(); err != nil {
		return nil, err
	}

	return j, nil
}

// gatekeeperImportOptions returns an error if any options are incompatible
func (j *jamel) gatekeeperImportOptions() error {
	if j.o.AppName == "" {
		j.o.AppName = "lingon"
	}
	if j.o.OutputPkgName == "" {
		j.o.OutputPkgName = strings.ReplaceAll(j.o.AppName, "-", "")
	}

	var err error
	if strings.Contains(j.o.OutputPkgName, "-") {
		err = errors.New("package name cannot contain a dash")
	}
	for _, f := range j.o.ManifestFiles {
		filename := filepath.Base(f)
		e := filepath.Ext(filename)
		if e != ".yaml" && e != ".yml" {
			err = errors.Join(err, fmt.Errorf("not yaml file: %s", f))
			continue
		}

		if !kubeutil.FileExists(f) {
			err = errors.Join(err, fmt.Errorf("file does not exist: %s", f))
			continue
		}
	}
	if j.o.Verbose && j.l == nil {
		err = errors.Join(err, errors.New("verbose option requires a logger"))
	}
	if err != nil {
		return errors.Join(ErrIncompatibleOptions, err)
	}
	return err
}

func (j *jamel) generateGo() error {
	if j.useReader {
		if j.o.Verbose {
			j.l.Info("importing from reader")
		}
		splitYaml, err := kubeutil.ManifestSplit(j.o.ManifestReader)
		if err != nil {
			return err
		}
		if len(splitYaml) == 0 {
			return fmt.Errorf("no manifest found")
		}

		err = j.convertToGo(splitYaml)
		if err != nil {
			return fmt.Errorf("stdin: %w", err)
		}
		return nil
	}

	if j.o.Verbose {
		j.l.Info(
			"importing from manifest",
			slog.Any("files", j.o.ManifestFiles),
		)
	}

	for _, filePath := range j.o.ManifestFiles {
		splitYaml, err := kubeutil.ManifestReadFile(filePath)
		if err != nil {
			if j.o.IgnoreErrors {
				j.l.Error(
					"manifest",
					logErrIgnored,
					slog.String("file", filePath),
					slog.String("error", err.Error()),
				)
				continue
			}
			return err
		}

		if j.o.Verbose {
			j.l.Info(
				"manifest",
				slog.String("file", filePath),
				slog.Int("manifests", len(splitYaml)),
			)
		}
		err = j.convertToGo(splitYaml)
		if err != nil {
			return fmt.Errorf("file %s: %w", filePath, err)
		}
	}

	return nil
}

func (j *jamel) convertToGo(splitYaml []string) error {
	vcpt := 1 // variable name counter to avoid name collisions
	scpt := 1 // struct field name counter to avoid name collisions
	for manifestNumber, y := range splitYaml {
		data := []byte(y)
		if len(data) == 0 {
			continue
		}
		if j.o.Verbose {
			head := min(maxOutputHead, len(data))
			j.l.Info(
				"converting manifest",
				slog.Int("number", manifestNumber+1),
				slog.String("head", string(data[:head])+"..."),
			)
		}
		m, err := kubeutil.ExtractMetadata(data)
		if err != nil {
			if j.o.IgnoreErrors {
				j.l.Error(
					"extract metadata",
					logErrIgnored,
					slog.Int("manifest", manifestNumber+1),
					slog.String("error", err.Error()),
				)
				continue
			}
			return fmt.Errorf(
				"extract metadata of manifest %d: %w",
				manifestNumber+1,
				err,
			)
		}

		// ConfigMap are not cleanup as the comments will be lost.
		if j.o.CleanUp && m.Kind != kindConfigMap {
			data, err = kubeutil.CleanUpYAML(data)
			if err != nil && !j.o.IgnoreErrors {
				return fmt.Errorf("clean up: %w", err)
			}
		}
		//
		// convert kubernetes objects to generated Go code
		//
		jenCode, err := j.yaml2GoJen(data, m)
		if err != nil {
			if j.o.IgnoreErrors {
				j.l.Error(
					"yaml to go",
					logErrIgnored,
					slog.Int("manifest", manifestNumber+1),
					slog.String("error", err.Error()),
				)
				continue
			}
			return err
		}
		if jenCode == nil {
			// List case
			continue
		}

		nameVar := j.o.NameVarFunc(*m)
		if j.o.RemoveAppName {
			nameVar = RemoveAppName(nameVar, j.o.AppName)
		}
		// check for duplicate
		if _, ok := j.objectsCode[nameVar]; ok {
			nameVar += strconv.Itoa(vcpt)
			vcpt++
		}
		j.objectsCode[nameVar] = jenCode

		//
		// kube.App struct
		//
		nameField := j.o.NameFieldFunc(*m)
		if j.o.RemoveAppName {
			nameField = RemoveAppName(nameField, j.o.AppName)
		}
		// check for duplicate
		if _, ok := j.kubeAppStructCode[nameField]; ok {
			nameField += strconv.Itoa(scpt)
			scpt++
		}

		// resolve package path for imports of top-level var declaration
		pkgPath, err := api.PkgPathFromAPIVersion(m.APIVersion)
		if err != nil {
			// try CRD
			if j.crdCurrent != "" {
				pkgPath = j.crdCurrent
				j.crdCurrent = "" // reset
			}
			// fail but don't err
			if pkgPath == "" {
				pkgPath = m.APIVersion
			}
		}

		structFieldType := jen.Qual(pkgPath, m.Kind)
		j.kubeAppStructCode[nameField] = structFieldType

		// nameFieldVar is used to map the name of the object to the variable
		// name
		j.nameFieldVar[nameField] = nameVar
		// objectsMeta maps the name of the variable to the Metadata
		j.objectsMeta[nameVar] = *m

		// hack: reset counters if too high
		if vcpt >= 10 || scpt >= 10 {
			vcpt = 1
			scpt = 1
		}
	}
	return nil
}

func (j *jamel) yaml2GoJen(data []byte, m *kubeutil.Metadata) (
	*jen.Statement,
	error,
) {
	decoded, _, err := j.o.Serializer.Decode(data, nil, nil)
	if err != nil {
		if runtime.IsNotRegisteredError(err) {
			return nil, err
		}
		return nil, fmt.Errorf("decoding manifest for %s: %w", m.GVK(), err)
	}

	var jenCode *jen.Statement
	switch dt := decoded.(type) {
	case *corev1.List:
		l := make([]string, 0, len(dt.Items))
		for _, i := range dt.Items {
			l = append(l, string(i.Raw))
		}
		if err = j.convertToGo(l); err != nil {
			return nil, fmt.Errorf("convert list: %w", err)
		}
	case *corev1.ConfigMap:
		// special case for ConfigMap
		// we want to extract the comments from the YAML
		jenCode, err = j.configMapComment(dt, data)
		if err != nil {
			return nil, fmt.Errorf("configmap comment: %w", err)
		}

	default:
		if decoded == nil {
			return jen.Empty(), nil
		}
		jenCode = j.convertValue(reflect.ValueOf(decoded), false)
	}
	return jenCode, nil
}
