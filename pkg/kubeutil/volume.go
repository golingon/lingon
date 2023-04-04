// Copyright 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package kubeutil

// adapted from: https://github.com/bwplotka/mimic/blob/prometheus-kubernetes-example/abstractions/kubernetes/volumes/volumes.go
import (
	corev1 "k8s.io/api/core/v1"
)

// VolumeAndMount is a helper struct to create a Volume and a VolumeSource
type VolumeAndMount struct {
	corev1.VolumeMount
	// corev1.Volume has just Name and VolumeSource.
	// A name field is already present in the VolumeMount,
	// so we just add the VolumeSource field here directly.
	VolumeSource corev1.VolumeSource
}

// Volume creates a Volume from the VolumeAndMount
func (vam VolumeAndMount) Volume() corev1.Volume {
	return corev1.Volume{
		Name:         vam.Name,
		VolumeSource: vam.VolumeSource,
	}
}

// VolumesAndMounts is a helper struct to create a list of Volumes and a list of VolumeSource
type VolumesAndMounts []VolumeAndMount

// Volumes creates a list of Volumes from the VolumesAndMounts
func (vams VolumesAndMounts) Volumes() []corev1.Volume {
	volumes := make([]corev1.Volume, 0, len(vams))
	for _, vam := range vams {
		volumes = append(volumes, vam.Volume())
	}
	return volumes
}

// VolumeMounts creates a list of VolumesMount from the VolumesAndMounts
func (vams VolumesAndMounts) VolumeMounts() []corev1.VolumeMount {
	mounts := make([]corev1.VolumeMount, 0, len(vams))
	for _, vam := range vams {
		mounts = append(mounts, vam.VolumeMount)
	}
	return mounts
}
