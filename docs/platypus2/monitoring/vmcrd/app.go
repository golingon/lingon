// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by lingon. EDIT AS MUCH AS YOU LIKE.

package vmcrd

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/volvo-cars/lingon/pkg/kube"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// validate the struct implements the interface
var _ kube.Exporter = (*Victoriametrics)(nil)

// Victoriametrics contains kubernetes manifests
type Victoriametrics struct {
	kube.App

	VmagentsOperatorVictoriametricsComCRD              *apiextensionsv1.CustomResourceDefinition
	VmalertmanagerconfigsOperatorVictoriametricsComCRD *apiextensionsv1.CustomResourceDefinition
	VmalertmanagersOperatorVictoriametricsComCRD       *apiextensionsv1.CustomResourceDefinition
	VmalertsOperatorVictoriametricsComCRD              *apiextensionsv1.CustomResourceDefinition
	VmauthsOperatorVictoriametricsComCRD               *apiextensionsv1.CustomResourceDefinition
	VmclustersOperatorVictoriametricsComCRD            *apiextensionsv1.CustomResourceDefinition
	VmnodescrapesOperatorVictoriametricsComCRD         *apiextensionsv1.CustomResourceDefinition
	VmpodscrapesOperatorVictoriametricsComCRD          *apiextensionsv1.CustomResourceDefinition
	VmprobesOperatorVictoriametricsComCRD              *apiextensionsv1.CustomResourceDefinition
	VmrulesOperatorVictoriametricsComCRD               *apiextensionsv1.CustomResourceDefinition
	VmservicescrapesOperatorVictoriametricsComCRD      *apiextensionsv1.CustomResourceDefinition
	VmsinglesOperatorVictoriametricsComCRD             *apiextensionsv1.CustomResourceDefinition
	VmstaticscrapesOperatorVictoriametricsComCRD       *apiextensionsv1.CustomResourceDefinition
	VmusersOperatorVictoriametricsComCRD               *apiextensionsv1.CustomResourceDefinition
}

// New creates a new Victoriametrics
func New() *Victoriametrics {
	return &Victoriametrics{
		VmagentsOperatorVictoriametricsComCRD:              VmagentsOperatorVictoriametricsComCRD,
		VmalertmanagerconfigsOperatorVictoriametricsComCRD: VmalertmanagerconfigsOperatorVictoriametricsComCRD,
		VmalertmanagersOperatorVictoriametricsComCRD:       VmalertmanagersOperatorVictoriametricsComCRD,
		VmalertsOperatorVictoriametricsComCRD:              VmalertsOperatorVictoriametricsComCRD,
		VmauthsOperatorVictoriametricsComCRD:               VmauthsOperatorVictoriametricsComCRD,
		VmclustersOperatorVictoriametricsComCRD:            VmclustersOperatorVictoriametricsComCRD,
		VmnodescrapesOperatorVictoriametricsComCRD:         VmnodescrapesOperatorVictoriametricsComCRD,
		VmpodscrapesOperatorVictoriametricsComCRD:          VmpodscrapesOperatorVictoriametricsComCRD,
		VmprobesOperatorVictoriametricsComCRD:              VmprobesOperatorVictoriametricsComCRD,
		VmrulesOperatorVictoriametricsComCRD:               VmrulesOperatorVictoriametricsComCRD,
		VmservicescrapesOperatorVictoriametricsComCRD:      VmservicescrapesOperatorVictoriametricsComCRD,
		VmsinglesOperatorVictoriametricsComCRD:             VmsinglesOperatorVictoriametricsComCRD,
		VmstaticscrapesOperatorVictoriametricsComCRD:       VmstaticscrapesOperatorVictoriametricsComCRD,
		VmusersOperatorVictoriametricsComCRD:               VmusersOperatorVictoriametricsComCRD,
	}
}

// Apply applies the kubernetes objects to the cluster
func (a *Victoriametrics) Apply(ctx context.Context) error {
	return Apply(ctx, a)
}

// Export exports the kubernetes objects to YAML files in the given directory
func (a *Victoriametrics) Export(dir string) error {
	return kube.Export(a, kube.WithExportOutputDirectory(dir))
}

// Apply applies the kubernetes objects contained in Exporter to the cluster
func Apply(ctx context.Context, km kube.Exporter) error {
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	cmd.Env = os.Environ()        // inherit environment in case we need to use kubectl from a container
	stdin, err := cmd.StdinPipe() // pipe to pass data to kubectl
	if err != nil {
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	go func() {
		defer func() {
			err = errors.Join(err, stdin.Close())
		}()
		if errEW := kube.Export(
			km,
			kube.WithExportWriter(stdin),
			kube.WithExportAsSingleFile("stdin"),
		); errEW != nil {
			err = errors.Join(err, errEW)
		}
	}()

	if errS := cmd.Start(); errS != nil {
		return errors.Join(err, errS)
	}

	// waits for the command to exit and waits for any copying
	// to stdin or copying from stdout or stderr to complete
	return errors.Join(err, cmd.Wait())
}

// P converts T to *T, useful for basic types
func P[T any](t T) *T {
	return &t
}
