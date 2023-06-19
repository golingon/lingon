// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by lingon. EDIT AS MUCH AS YOU LIKE.

package promstack

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

var KubePromtheusStackGrafanaSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/name":       "grafana",
			"app.kubernetes.io/version":    "9.5.3",
			"helm.sh/chart":                "grafana-6.57.1",
		},
		Name:      "kube-promtheus-stack-grafana",
		Namespace: "monitoring",
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{{
			Name:       "http-web",
			Port:       int32(80),
			Protocol:   corev1.Protocol("TCP"),
			TargetPort: intstr.IntOrString{IntVal: int32(3000)},
		}},
		Selector: map[string]string{
			"app.kubernetes.io/instance": "kube-promtheus-stack",
			"app.kubernetes.io/name":     "grafana",
		},
		Type: corev1.ServiceType("ClusterIP"),
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}

var KubePromtheusStackKubeAlertmanagerSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app":                          "kube-prometheus-stack-alertmanager",
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/part-of":    "kube-prometheus-stack",
			"app.kubernetes.io/version":    "46.8.0",
			"chart":                        "kube-prometheus-stack-46.8.0",
			"heritage":                     "Helm",
			"release":                      "kube-promtheus-stack",
			"self-monitor":                 "true",
		},
		Name:      "kube-promtheus-stack-kube-alertmanager",
		Namespace: "monitoring",
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{{
			Name:       "http-web",
			Port:       int32(9093),
			Protocol:   corev1.Protocol("TCP"),
			TargetPort: intstr.IntOrString{IntVal: int32(9093)},
		}},
		Selector: map[string]string{
			"alertmanager":           "kube-promtheus-stack-kube-alertmanager",
			"app.kubernetes.io/name": "alertmanager",
		},
		Type: corev1.ServiceType("ClusterIP"),
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}

var KubePromtheusStackKubeCorednsSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app":                          "kube-prometheus-stack-coredns",
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/part-of":    "kube-prometheus-stack",
			"app.kubernetes.io/version":    "46.8.0",
			"chart":                        "kube-prometheus-stack-46.8.0",
			"heritage":                     "Helm",
			"jobLabel":                     "coredns",
			"release":                      "kube-promtheus-stack",
		},
		Name:      "kube-promtheus-stack-kube-coredns",
		Namespace: "kube-system",
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: "None",
		Ports: []corev1.ServicePort{{
			Name:       "http-metrics",
			Port:       int32(9153),
			Protocol:   corev1.Protocol("TCP"),
			TargetPort: intstr.IntOrString{IntVal: int32(9153)},
		}},
		Selector: map[string]string{"k8s-app": "kube-dns"},
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}

var KubePromtheusStackKubeKubeControllerManagerSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app":                          "kube-prometheus-stack-kube-controller-manager",
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/part-of":    "kube-prometheus-stack",
			"app.kubernetes.io/version":    "46.8.0",
			"chart":                        "kube-prometheus-stack-46.8.0",
			"heritage":                     "Helm",
			"jobLabel":                     "kube-controller-manager",
			"release":                      "kube-promtheus-stack",
		},
		Name:      "kube-promtheus-stack-kube-kube-controller-manager",
		Namespace: "kube-system",
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: "None",
		Ports: []corev1.ServicePort{{
			Name:       "http-metrics",
			Port:       int32(10257),
			Protocol:   corev1.Protocol("TCP"),
			TargetPort: intstr.IntOrString{IntVal: int32(10257)},
		}},
		Selector: map[string]string{"component": "kube-controller-manager"},
		Type:     corev1.ServiceType("ClusterIP"),
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}

var KubePromtheusStackKubeKubeEtcdSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app":                          "kube-prometheus-stack-kube-etcd",
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/part-of":    "kube-prometheus-stack",
			"app.kubernetes.io/version":    "46.8.0",
			"chart":                        "kube-prometheus-stack-46.8.0",
			"heritage":                     "Helm",
			"jobLabel":                     "kube-etcd",
			"release":                      "kube-promtheus-stack",
		},
		Name:      "kube-promtheus-stack-kube-kube-etcd",
		Namespace: "kube-system",
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: "None",
		Ports: []corev1.ServicePort{{
			Name:       "http-metrics",
			Port:       int32(2381),
			Protocol:   corev1.Protocol("TCP"),
			TargetPort: intstr.IntOrString{IntVal: int32(2381)},
		}},
		Selector: map[string]string{"component": "etcd"},
		Type:     corev1.ServiceType("ClusterIP"),
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}

var KubePromtheusStackKubeKubeProxySVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app":                          "kube-prometheus-stack-kube-proxy",
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/part-of":    "kube-prometheus-stack",
			"app.kubernetes.io/version":    "46.8.0",
			"chart":                        "kube-prometheus-stack-46.8.0",
			"heritage":                     "Helm",
			"jobLabel":                     "kube-proxy",
			"release":                      "kube-promtheus-stack",
		},
		Name:      "kube-promtheus-stack-kube-kube-proxy",
		Namespace: "kube-system",
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: "None",
		Ports: []corev1.ServicePort{{
			Name:       "http-metrics",
			Port:       int32(10249),
			Protocol:   corev1.Protocol("TCP"),
			TargetPort: intstr.IntOrString{IntVal: int32(10249)},
		}},
		Selector: map[string]string{"k8s-app": "kube-proxy"},
		Type:     corev1.ServiceType("ClusterIP"),
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}

var KubePromtheusStackKubeKubeSchedulerSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app":                          "kube-prometheus-stack-kube-scheduler",
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/part-of":    "kube-prometheus-stack",
			"app.kubernetes.io/version":    "46.8.0",
			"chart":                        "kube-prometheus-stack-46.8.0",
			"heritage":                     "Helm",
			"jobLabel":                     "kube-scheduler",
			"release":                      "kube-promtheus-stack",
		},
		Name:      "kube-promtheus-stack-kube-kube-scheduler",
		Namespace: "kube-system",
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: "None",
		Ports: []corev1.ServicePort{{
			Name:       "http-metrics",
			Port:       int32(10259),
			Protocol:   corev1.Protocol("TCP"),
			TargetPort: intstr.IntOrString{IntVal: int32(10259)},
		}},
		Selector: map[string]string{"component": "kube-scheduler"},
		Type:     corev1.ServiceType("ClusterIP"),
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}

var KubePromtheusStackKubeOperatorSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app":                          "kube-prometheus-stack-operator",
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/part-of":    "kube-prometheus-stack",
			"app.kubernetes.io/version":    "46.8.0",
			"chart":                        "kube-prometheus-stack-46.8.0",
			"heritage":                     "Helm",
			"release":                      "kube-promtheus-stack",
		},
		Name:      "kube-promtheus-stack-kube-operator",
		Namespace: "monitoring",
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{{
			Name: "https",
			Port: int32(443),
			TargetPort: intstr.IntOrString{
				StrVal: "https",
				Type:   intstr.Type(int64(1)),
			},
		}},
		Selector: map[string]string{
			"app":     "kube-prometheus-stack-operator",
			"release": "kube-promtheus-stack",
		},
		Type: corev1.ServiceType("ClusterIP"),
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}

var KubePromtheusStackKubePrometheusSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app":                          "kube-prometheus-stack-prometheus",
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/part-of":    "kube-prometheus-stack",
			"app.kubernetes.io/version":    "46.8.0",
			"chart":                        "kube-prometheus-stack-46.8.0",
			"heritage":                     "Helm",
			"release":                      "kube-promtheus-stack",
			"self-monitor":                 "true",
		},
		Name:      "kube-promtheus-stack-kube-prometheus",
		Namespace: "monitoring",
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{{
			Name:       "http-web",
			Port:       int32(9090),
			TargetPort: intstr.IntOrString{IntVal: int32(9090)},
		}},
		Selector: map[string]string{
			"app.kubernetes.io/name": "prometheus",
			"prometheus":             "kube-promtheus-stack-kube-prometheus",
		},
		Type: corev1.ServiceType("ClusterIP"),
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}

var KubePromtheusStackKubeStateMetricsSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Annotations: map[string]string{"prometheus.io/scrape": "true"},
		Labels: map[string]string{
			"app.kubernetes.io/component":  "metrics",
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/name":       "kube-state-metrics",
			"app.kubernetes.io/part-of":    "kube-state-metrics",
			"app.kubernetes.io/version":    "2.9.2",
			"helm.sh/chart":                "kube-state-metrics-5.7.0",
			"release":                      "kube-promtheus-stack",
		},
		Name:      "kube-promtheus-stack-kube-state-metrics",
		Namespace: "monitoring",
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{{
			Name:       "http",
			Port:       int32(8080),
			Protocol:   corev1.Protocol("TCP"),
			TargetPort: intstr.IntOrString{IntVal: int32(8080)},
		}},
		Selector: map[string]string{
			"app.kubernetes.io/instance": "kube-promtheus-stack",
			"app.kubernetes.io/name":     "kube-state-metrics",
		},
		Type: corev1.ServiceType("ClusterIP"),
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}

var KubePromtheusStackPrometheusNodeExporterSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app.kubernetes.io/component":  "metrics",
			"app.kubernetes.io/instance":   "kube-promtheus-stack",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/name":       "prometheus-node-exporter",
			"app.kubernetes.io/part-of":    "prometheus-node-exporter",
			"app.kubernetes.io/version":    "1.5.0",
			"helm.sh/chart":                "prometheus-node-exporter-4.17.5",
			"jobLabel":                     "node-exporter",
			"release":                      "kube-promtheus-stack",
		},
		Name:      "kube-promtheus-stack-prometheus-node-exporter",
		Namespace: "monitoring",
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{{
			Name:       "http-metrics",
			Port:       int32(9100),
			Protocol:   corev1.Protocol("TCP"),
			TargetPort: intstr.IntOrString{IntVal: int32(9100)},
		}},
		Selector: map[string]string{
			"app.kubernetes.io/instance": "kube-promtheus-stack",
			"app.kubernetes.io/name":     "prometheus-node-exporter",
		},
		Type: corev1.ServiceType("ClusterIP"),
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
}
