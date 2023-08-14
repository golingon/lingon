// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by lingon. EDIT AS MUCH AS YOU LIKE.

package policy

import (
	ku "github.com/volvo-cars/lingon/pkg/kubeutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var MetricsSVC = &corev1.Service{
	TypeMeta:   ku.TypeServiceV1,
	ObjectMeta: W.ObjectMetaNameSuffix("metrics"),
	Spec: corev1.ServiceSpec{
		Ports:    []corev1.ServicePort{W.Metrics.Service},
		Selector: W.MatchLabels(),
		Type:     corev1.ServiceTypeClusterIP,
	},
}

var SVC = &corev1.Service{
	TypeMeta:   ku.TypeServiceV1,
	ObjectMeta: W.ObjectMeta(),
	Spec: corev1.ServiceSpec{
		Ports:    []corev1.ServicePort{W.MainPort.Service},
		Selector: W.MatchLabels(),
		Type:     corev1.ServiceTypeClusterIP,
	},
}

var affinity = &corev1.Affinity{
	PodAntiAffinity: &corev1.PodAntiAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
			{
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: W.MatchLabels(),
					},
					TopologyKey: ku.LabelHostname,
				},
				Weight: int32(100),
			},
		},
	},
}

var Deploy = &appsv1.Deployment{
	TypeMeta:   ku.TypeDeploymentV1,
	ObjectMeta: W.ObjectMeta(),
	Spec: appsv1.DeploymentSpec{
		Replicas: P(int32(1)),
		Selector: &metav1.LabelSelector{MatchLabels: W.MatchLabels()},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: W.Labels()},
			Spec: corev1.PodSpec{
				Affinity:           affinity,
				ServiceAccountName: W.SA.Name,
				Containers: []corev1.Container{
					{
						Name:            W.Name,
						Image:           W.Metadata.Img.URL(),
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env: []corev1.EnvVar{
							ku.EnvVarDownAPI(
								"SYSTEM_NAMESPACE",
								"metadata.namespace",
							),
							{
								Name:  "CONFIG_LOGGING_NAME",
								Value: ConfigLoggingCM.Name,
							},
							{
								Name:  "CONFIG_OBSERVABILITY_NAME",
								Value: "sigstore-policy-controller-webhook-observability",
							},
							{
								Name:  "METRICS_DOMAIN",
								Value: "sigstore.dev/policy",
							},
							{
								Name:  "HOME",
								Value: W.HomeDir.VolumeMount.MountPath,
							},
							{
								Name: "WEBHOOK_NAME", Value: SVC.Name,
							},
						},
						LivenessProbe: probe(
							"/healthz",
							int(W.MainPort.Container.ContainerPort),
						),
						ReadinessProbe: probe(
							"/readyz",
							int(W.MainPort.Container.ContainerPort),
						),
						Ports: []corev1.ContainerPort{
							W.MainPort.Container,
							W.Metrics.Container,
						},
						Resources: ku.Resources(
							"100m",
							"128Mi",
							"200m",
							"512Mi",
						),
						SecurityContext: &corev1.SecurityContext{
							Capabilities:           &corev1.Capabilities{Drop: []corev1.Capability{corev1.Capability("ALL")}},
							ReadOnlyRootFilesystem: P(true),
							RunAsUser:              P(int64(1000)),
						},
						VolumeMounts: []corev1.VolumeMount{
							W.HomeDir.VolumeMount,
						},
					},
				},
				TerminationGracePeriodSeconds: P(int64(300)),
				Volumes: []corev1.Volume{
					W.HomeDir.Volume(),
				},
			},
		},
	},
}

func probe(path string, port int) *corev1.Probe {
	return &corev1.Probe{
		FailureThreshold:    int32(6),
		InitialDelaySeconds: int32(20),
		PeriodSeconds:       int32(1),
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				HTTPHeaders: []corev1.HTTPHeader{
					{
						Name:  "k-kubelet-probe",
						Value: "webhook",
					},
				},
				Path:   path,
				Port:   intstr.FromInt(port),
				Scheme: corev1.URISchemeHTTPS,
			},
		},
	}
}