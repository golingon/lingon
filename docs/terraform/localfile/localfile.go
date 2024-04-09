// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package localfile

//go:generate echo "\n>>>> LOCALFILE: generating hashicorp/local terra provider\n"
//go:generate go run -mod=readonly github.com/golingon/lingon/cmd/terragen -out ./out/local -pkg github.com/golingon/lingon/examples/localfile/out/local -force -provider local=hashicorp/local:2.4.0

import (
	"github.com/golingon/lingon/docs/terraform/localfile/out/local"
	"github.com/golingon/lingon/pkg/terra"
)

// NewLocalFileStack returns a new LocalFileStack which implements the
// terra.Exporter interface
// and can be exported into Terraform configuration
func NewLocalFileStack(filename string) *LocalFileStack {
	return &LocalFileStack{
		Backend: LocalBackend{
			Path: "terraform.tfstate",
		},
		Provider: local.NewProvider(local.ProviderArgs{}),
		File: local.NewFile(
			"file", local.FileArgs{
				Filename: terra.String(filename),
				Content:  terra.String("contents"),
			},
		),
	}
}

type LocalFileStack struct {
	terra.Stack

	Backend  LocalBackend    `validate:"required"`
	Provider *local.Provider `validate:"required"`
	File     *local.File     `validate:"required"`
}

var _ terra.Backend = (*LocalBackend)(nil)

// LocalBackend implements the Terraform local backend type.
// https://developer.hashicorp.com/terraform/language/settings/backends/local
type LocalBackend struct {
	Path string `hcl:"path,attr" validate:"required"`
}

// BackendType defines the type of the backend and implements the terra.Backend
// interface
func (b LocalBackend) BackendType() string {
	return "local"
}
