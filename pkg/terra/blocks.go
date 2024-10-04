// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terra

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Resource represents a Terraform Resource.
// The generated Go structs from a Terraform provider resource will implement
// this interface and be used when exported to Terraform Configuration
type Resource interface {
	// Type returns the resource type, e.g. aws_iam_role
	Type() string
	// LocalName returns the unique name of the resource as it will be stored in
	// the state
	LocalName() string
	// Configuration returns the arguments for the resource
	Configuration() interface{}
	// Dependencies returns the list of resources that this resource depends_on
	Dependencies() Dependencies
	// LifecycleManagement returns the lifecycle configuration for this resource
	LifecycleManagement() *Lifecycle
	// ImportState takes the given attributes value map (from a Terraform state)
	// and imports it
	// into this resource
	ImportState(attributes io.Reader) error
}

// DataSource represents a Terraform DataSource.
// The generated Go structs from a Terraform provider data resource will
// implement this interface.
type DataSource interface {
	DataSource() string
	LocalName() string
	Configuration() interface{}
}

// Provider represents a Terraform Provider.
// The generated Go structs from a Terraform provider configuration will
// implement this interface.
type Provider interface {
	LocalName() string
	Source() string
	Version() string
	Configuration() interface{}
}

// Backend represents a Terraform Backend.
// Users will define their backends to implement this interface.
type Backend interface {
	BackendType() string
}

// StackObjects contains all the blocks that are extracted from a user-defined
// stack.
type StackObjects struct {
	Backend     Backend
	Providers   []Provider
	Resources   []Resource
	DataSources []DataSource
}

const (
	tagLingon = "lingon"
)

var (
	ErrNoBackendBlock        = errors.New("stack must have a backend block")
	ErrMultipleBackendBlocks = errors.New(
		"stack cannot have multiple backend blocks",
	)
	ErrNoProviderBlock  = errors.New("stack must have a provider block")
	ErrNotExportedField = errors.New(
		"stack has non-exported (private) field",
	)
	ErrUnknownPublicField = errors.New("unknown public field")
)

// ObjectsFromStack takes a terra stack and returns all the terra objects
// (resources, data sources, providers, and backend) that are defined in the
// stack.
func ObjectsFromStack(stack Exporter) (*StackObjects, error) {
	if err := validator.New().Struct(stack); err != nil {
		return nil, fmt.Errorf("stack validation failed: %w", err)
	}

	rv := reflect.ValueOf(stack)
	sb := StackObjects{}
	if err := parseStackStructFields(rv, &sb); err != nil {
		return nil, err
	}

	return &sb, nil
}

// parseStackStructFields takes a struct reflect.Value and appends any Terraform
// objects that it finds to the provider stackObjects.
func parseStackStructFields(rv reflect.Value, sb *StackObjects) error {
	// Skip nil pointers.
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return nil
	}
	// Resolve pointers.
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	for i := 0; i < rv.NumField(); i++ {
		sf := rv.Type().Field(i)
		fv := rv.Field(i)
		// Handle embedded (i.e. anonymous) structs.
		if sf.Anonymous {
			if !sf.IsExported() {
				return fmt.Errorf(
					"%w, PkgPath: %s, Name: %s, Type: %s",
					ErrNotExportedField,
					sf.PkgPath,
					sf.Name,
					sf.Type,
				)
			}
			// Proceed and parse the embedded struct.
			if err := parseStackStructFields(fv, sb); err != nil {
				return err
			}
			continue
		}

		// Check if field has a terriyaki struct tag.
		if tkiTag, ok := sf.Tag.Lookup(tagLingon); ok {
			// Ignore fields with the "-" value.
			if tkiTag == "-" {
				continue
			}
		}
		if !sf.IsExported() {
			return fmt.Errorf(
				"%w, PkgPath: %s, Name: %s, Type: %s",
				ErrNotExportedField,
				sf.PkgPath,
				sf.Name,
				sf.Type,
			)
		}

		tkiObjects := make([]interface{}, 0)
		// Handle slices and arrays.
		switch fv.Kind() {
		case reflect.Array, reflect.Slice:
			// If it's an array or slice, iterate over the elements and process
			// them individually.
			for j := 0; j < fv.Len(); j++ {
				tkiObjects = append(tkiObjects, fv.Index(j).Interface())
			}
		default:
			tkiObjects = append(tkiObjects, fv.Interface())
		}
		for _, obj := range tkiObjects {
			switch v := obj.(type) {
			case Resource:
				sb.Resources = append(sb.Resources, v)
			case DataSource:
				sb.DataSources = append(sb.DataSources, v)
			case Provider:
				sb.Providers = append(sb.Providers, v)
			case Backend:
				if sb.Backend != nil {
					return ErrMultipleBackendBlocks
				}
				sb.Backend = v
			case Exporter:
				// Recursively parse the struct.
				if err := parseStackStructFields(reflect.ValueOf(v), sb); err != nil {
					return err
				}
			default:
				// Not sure if this should be an error, but rather be explicit
				// at this point.
				return fmt.Errorf(
					"%w: %s, type: %s",
					ErrUnknownPublicField,
					sf.Name,
					sf.Type,
				)
			}
		}
	}
	return nil
}
