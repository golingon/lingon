// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rogpeppe/go-internal/txtar"
)

func Folder2Txtar(folder string) (*txtar.Archive, error) {
	files, err := listFiles(folder)
	if err != nil {
		return nil, err
	}
	var ar txtar.Archive
	ar.Files = make([]txtar.File, 0, len(files))
	for i, f := range files {
		ar.Files = append(ar.Files, txtar.File{Name: f})
		ar.Files[i].Data, err = os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}
	}
	return &ar, nil
}

func listFiles(root string) ([]string, error) {
	var files []string
	fi, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.New("root is not a directory")
	}
	err = filepath.Walk(
		root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("walk %q %q, %w", path, info.Name(), err)
			}

			if !info.IsDir() {
				files = append(files, path)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("walk: %w", err)
	}
	return files, nil
}
