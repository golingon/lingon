// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package vmk8s

import (
	vmo "github.com/VictoriaMetrics/operator/api/operator/v1"
	"github.com/golingon/lingon/pkg/kube"
	ku "github.com/golingon/lingon/pkg/kubeutil"
	"github.com/golingon/lingoneks/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ETPort     = 2379
	ETPortName = "http-metrics"
)

var ET = &meta.Metadata{
	Name:      "kube-etcd", // linked to the name of the JobLabel
	Namespace: namespace,
	Instance:  "kube-etcd-" + namespace,
	Component: "monitoring",
	PartOf:    appName,
	Version:   version,
	ManagedBy: "lingon",
}

type MonETCD struct {
	kube.App

	ETCDRules  *vmo.VMRule
	ETCDSvc    *corev1.Service
	ETCDScrape *vmo.VMServiceScrape
}

func NewMonETCD() *MonETCD {
	return &MonETCD{
		ETCDRules:  ETCDRules,
		ETCDSvc:    ETCDSvc,
		ETCDScrape: ETCDScrape,
	}
}

var ETCDRules = &vmo.VMRule{
	ObjectMeta: ET.ObjectMeta(),
	Spec: vmo.VMRuleSpec{
		Groups: []vmo.RuleGroup{
			{
				Name: "etcd",
				Rules: []vmo.Rule{
					{
						Alert: "etcdInsufficientMembers",
						Annotations: map[string]string{
							"message": `etcd cluster "{{ $labels.job }}": insufficient members ({{ $value }}).`,
						},
						Expr:   `sum(up{job=~".*etcd.*"} == bool 1) by (job) < ((count(up{job=~".*etcd.*"}) by (job) + 1) / 2)`,
						For:    "3m",
						Labels: map[string]string{"severity": "critical"},
					}, {
						Alert: "etcdNoLeader",
						Annotations: map[string]string{
							"message": `etcd cluster "{{ $labels.job }}": member {{ $labels.instance }} has no leader.`,
						},
						Expr:   `etcd_server_has_leader{job=~".*etcd.*"} == 0`,
						For:    "1m",
						Labels: map[string]string{"severity": "critical"},
					}, {
						Alert: "etcdHighNumberOfLeaderChanges",
						Annotations: map[string]string{
							"message": `etcd cluster "{{ $labels.job }}": instance {{ $labels.instance }} has seen {{ $value }} leader changes within the last hour.`,
						},
						Expr:   `rate(etcd_server_leader_changes_seen_total{job=~".*etcd.*"}[15m]) > 3`,
						For:    "15m",
						Labels: map[string]string{"severity": "warning"},
					}, {
						Alert: "etcdHighNumberOfFailedGRPCRequests",
						Annotations: map[string]string{
							"message": `etcd cluster "{{ $labels.job }}": {{ $value }}% of requests for {{ $labels.grpc_method }} failed on etcd instance {{ $labels.instance }}.`,
						},
						Expr: `
100 * sum(rate(grpc_server_handled_total{job=~".*etcd.*", grpc_code!="OK"}[5m])) BY (job, instance, grpc_service, grpc_method)
  /
sum(rate(grpc_server_handled_total{job=~".*etcd.*"}[5m])) BY (job, instance, grpc_service, grpc_method)
  > 1
`,
						For:    "10m",
						Labels: map[string]string{"severity": "warning"},
					}, {
						Alert: "etcdHighNumberOfFailedGRPCRequests",
						Annotations: map[string]string{
							"message": `etcd cluster "{{ $labels.job }}": {{ $value }}% of requests for {{ $labels.grpc_method }} failed on etcd instance {{ $labels.instance }}.`,
						},
						Expr: `
100 * sum(rate(grpc_server_handled_total{job=~".*etcd.*", grpc_code!="OK"}[5m])) BY (job, instance, grpc_service, grpc_method)
  /
sum(rate(grpc_server_handled_total{job=~".*etcd.*"}[5m])) BY (job, instance, grpc_service, grpc_method)
  > 5
`,
						For:    "5m",
						Labels: map[string]string{"severity": "critical"},
					}, {
						Alert: "etcdGRPCRequestsSlow",
						Annotations: map[string]string{
							"message": `etcd cluster "{{ $labels.job }}": gRPC requests to {{ $labels.grpc_method }} are taking {{ $value }}s on etcd instance {{ $labels.instance }}.`,
						},
						Expr: `
histogram_quantile(0.99, 
	sum(rate(grpc_server_handling_seconds_bucket{job=~".*etcd.*", grpc_type="unary"}[5m])) 
by (job, instance, grpc_service, grpc_method, le))
> 0.15
`,
						For:    "10m",
						Labels: map[string]string{"severity": "critical"},
					}, {
						Alert: "etcdMemberCommunicationSlow",
						Annotations: map[string]string{
							"message": `etcd cluster "{{ $labels.job }}": member communication with {{ $labels.To }} is taking {{ $value }}s on etcd instance {{ $labels.instance }}.`,
						},
						Expr:   `histogram_quantile(0.99, rate(etcd_network_peer_round_trip_time_seconds_bucket{job=~".*etcd.*"}[5m])) > 0.15 `,
						For:    "10m",
						Labels: map[string]string{"severity": "warning"},
					}, {
						Alert: "etcdHighNumberOfFailedProposals",
						Annotations: map[string]string{
							"message": `etcd cluster "{{ $labels.job }}": {{ $value }} proposal failures within the last hour on etcd instance {{ $labels.instance }}.`,
						},
						Expr:   `rate(etcd_server_proposals_failed_total{job=~".*etcd.*"}[15m]) > 5`,
						For:    "15m",
						Labels: map[string]string{"severity": "warning"},
					}, {
						Alert: "etcdHighFsyncDurations",
						Annotations: map[string]string{
							"message": `etcd cluster "{{ $labels.job }}": 99th percentile fync durations are {{ $value }}s on etcd instance {{ $labels.instance }}.`,
						},
						Expr:   `histogram_quantile(0.99, rate(etcd_disk_wal_fsync_duration_seconds_bucket{job=~".*etcd.*"}[5m])) > 0.5 `,
						For:    "10m",
						Labels: map[string]string{"severity": "warning"},
					}, {
						Alert: "etcdHighCommitDurations",
						Annotations: map[string]string{
							"message": `etcd cluster "{{ $labels.job }}": 99th percentile commit durations {{ $value }}s on etcd instance {{ $labels.instance }}.`,
						},
						Expr:   `histogram_quantile(0.99, rate(etcd_disk_backend_commit_duration_seconds_bucket{job=~".*etcd.*"}[5m])) > 0.25 `,
						For:    "10m",
						Labels: map[string]string{"severity": "warning"},
					}, {
						Alert: "etcdHighNumberOfFailedHTTPRequests",
						Annotations: map[string]string{
							"message": "{{ $value }}% of requests for {{ $labels.method }} failed on etcd instance {{ $labels.instance }}",
						},
						Expr: `
sum(rate(etcd_http_failed_total{job=~".*etcd.*", code!="404"}[5m])) BY (method) 
/ 
sum(rate(etcd_http_received_total{job=~".*etcd.*"}[5m])) BY (method) 
> 0.01
`,
						For:    "10m",
						Labels: map[string]string{"severity": "warning"},
					}, {
						Alert: "etcdHighNumberOfFailedHTTPRequests",
						Annotations: map[string]string{
							"message": "{{ $value }}% of requests for {{ $labels.method }} failed on etcd instance {{ $labels.instance }}.",
						},
						Expr: `
sum(rate(etcd_http_failed_total{job=~".*etcd.*", code!="404"}[5m])) BY (method) 
/ 
sum(rate(etcd_http_received_total{job=~".*etcd.*"}[5m])) BY (method) 
> 0.05
`,
						For:    "10m",
						Labels: map[string]string{"severity": "critical"},
					}, {
						Alert: "etcdHTTPRequestsSlow",
						Annotations: map[string]string{
							"message": "etcd instance {{ $labels.instance }} HTTP requests to {{ $labels.method }} are slow.",
						},
						Expr:   `histogram_quantile(0.99, rate(etcd_http_successful_duration_seconds_bucket[5m])) > 0.15 `,
						For:    "10m",
						Labels: map[string]string{"severity": "warning"},
					},
				},
			},
		},
	},
	TypeMeta: TypeVMRulevmo,
}

var ETCDSvc = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels:    ET.Labels(),
		Name:      ET.Name,
		Namespace: ku.NSKubeSystem,
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: corev1.ClusterIPNone,
		Ports: []corev1.ServicePort{
			{
				Name:       ETPortName,
				Port:       int32(ETPort),
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(ETPort),
			},
		},
		Selector: map[string]string{"component": "etcd"},
		Type:     corev1.ServiceTypeClusterIP,
	},
	TypeMeta: ku.TypeServiceV1,
}

var ETCDScrape = &vmo.VMServiceScrape{
	ObjectMeta: ET.ObjectMeta(),
	Spec: vmo.VMServiceScrapeSpec{
		Endpoints: []vmo.Endpoint{
			{
				BearerTokenFile: PathSA + "/token",
				Port:            ETPortName,
				Scheme:          "https",
				TLSConfig:       &vmo.TLSConfig{CAFile: PathSA + "/ca.crt"},
			},
		},
		JobLabel: "component",
		NamespaceSelector: vmo.NamespaceSelector{
			MatchNames: []string{ku.NSKubeSystem}, // kube-system
		},
		Selector: metav1.LabelSelector{
			MatchLabels: map[string]string{"component": "etcd"},
		},
	},
	TypeMeta: TypeVMServiceScrapevmo,
}
