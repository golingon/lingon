// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by lingon. EDIT AS MUCH AS YOU LIKE.

package nats

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var STS = &appsv1.StatefulSet{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app.kubernetes.io/instance":   "nats",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/name":       "nats",
			"app.kubernetes.io/version":    "2.9.16",
			"helm.sh/chart":                "nats-0.19.13",
		},
		Name:      "nats",
		Namespace: "nats",
	},
	Spec: appsv1.StatefulSetSpec{
		PodManagementPolicy: appsv1.PodManagementPolicyType("Parallel"),
		Replicas:            P(int32(3)),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/instance": "nats",
				"app.kubernetes.io/name":     "nats",
			},
		},
		ServiceName: "nats",
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"checksum/config":      "7f38e14ac96d519ba71b7154f94978c89119ea52eb682d730406fcfa3eeee880",
					"prometheus.io/path":   "/metrics",
					"prometheus.io/port":   "7777",
					"prometheus.io/scrape": "true",
				},
				Labels: map[string]string{
					"app.kubernetes.io/instance": "nats",
					"app.kubernetes.io/name":     "nats",
				},
			},
			Spec: corev1.PodSpec{
				Affinity: &corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "app",
											Operator: metav1.LabelSelectorOperator("In"),
											Values:   []string{"nats"},
										},
									},
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Command: []string{
							"nats-server",
							"--config",
							"/etc/nats-config/nats.conf",
						},
						Env: []corev1.EnvVar{
							{
								Name:      "POD_NAME",
								ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
							}, {
								Name:  "SERVER_NAME",
								Value: "$(POD_NAME)",
							}, {
								Name:      "POD_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
							}, {
								Name:  "CLUSTER_ADVERTISE",
								Value: "$(POD_NAME).nats.$(POD_NAMESPACE).svc.cluster.local",
							},
						},
						Image:           "nats:2.9.16-alpine",
						ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
						Lifecycle: &corev1.Lifecycle{
							PreStop: &corev1.LifecycleHandler{
								Exec: &corev1.ExecAction{
									Command: []string{
										"nats-server",
										"-sl=ldm=/var/run/nats/nats.pid",
									},
								},
							},
						},
						LivenessProbe: &corev1.Probe{
							FailureThreshold:    int32(3),
							InitialDelaySeconds: int32(10),
							PeriodSeconds:       int32(30),
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/",
									Port: intstr.IntOrString{IntVal: int32(8222)},
								},
							},
							SuccessThreshold: int32(1),
							TimeoutSeconds:   int32(5),
						},
						Name: "nats",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: int32(4222),
								Name:          "client",
							}, {
								ContainerPort: int32(6222),
								Name:          "cluster",
							}, {
								ContainerPort: int32(8222),
								Name:          "monitor",
							},
						},
						ReadinessProbe: &corev1.Probe{
							FailureThreshold:    int32(3),
							InitialDelaySeconds: int32(10),
							PeriodSeconds:       int32(10),
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/",
									Port: intstr.IntOrString{IntVal: int32(8222)},
								},
							},
							SuccessThreshold: int32(1),
							TimeoutSeconds:   int32(5),
						},
						Resources: corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceName("cpu"):    resource.MustParse("2"),
								corev1.ResourceName("memory"): resource.MustParse("4Gi"),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceName("cpu"):    resource.MustParse("2"),
								corev1.ResourceName("memory"): resource.MustParse("4Gi"),
							},
						},
						StartupProbe: &corev1.Probe{
							FailureThreshold:    int32(90),
							InitialDelaySeconds: int32(10),
							PeriodSeconds:       int32(10),
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz",
									Port: intstr.IntOrString{IntVal: int32(8222)},
								},
							},
							SuccessThreshold: int32(1),
							TimeoutSeconds:   int32(5),
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/etc/nats-config",
								Name:      "config-volume",
							}, {
								MountPath: "/var/run/nats",
								Name:      "pid",
							},
						},
					}, {
						Command: []string{
							"nats-server-config-reloader",
							"-pid",
							"/var/run/nats/nats.pid",
							"-config",
							"/etc/nats-config/nats.conf",
						},
						Image:           "natsio/nats-server-config-reloader:0.10.1",
						ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
						Name:            "reloader",
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/etc/nats-config",
								Name:      "config-volume",
							}, {
								MountPath: "/var/run/nats",

								Name: "pid",
							},
						},
					}, {
						Args: []string{
							"-connz",
							"-routez",
							"-subz",
							"-varz",
							"-prefix=nats",
							"-use_internal_server_id",
							"http://localhost:8222/",
						},
						Image:           "natsio/prometheus-nats-exporter:0.10.1",
						ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
						Name:            "metrics",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: int32(7777),
								Name:          "metrics",
							},
						},
					},
				},
				DNSPolicy:                     corev1.DNSPolicy("ClusterFirst"),
				ServiceAccountName:            "nats",
				ShareProcessNamespace:         P(true),
				TerminationGracePeriodSeconds: P(int64(60)),
				Volumes: []corev1.Volume{
					{
						Name:         "config-volume",
						VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "nats-config"}}},
					}, {
						Name:         "pid",
						VolumeSource: corev1.VolumeSource{},
					},
				},
			},
		},
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "StatefulSet",
	},
}