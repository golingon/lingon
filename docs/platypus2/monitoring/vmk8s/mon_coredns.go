// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package vmk8s

import (
	_ "embed"

	vmo "github.com/VictoriaMetrics/operator/api/operator/v1"
	"github.com/golingon/lingon/pkg/kube"
	ku "github.com/golingon/lingon/pkg/kubeutil"
	"github.com/golingon/lingoneks/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	CDNSPort     = 9153
	CDNSPortName = "metrics"
)

var CDNS = &meta.Metadata{
	Name:      "coredns", // linked to the name of the JobLabel
	Namespace: namespace,
	Instance:  "coredns-" + namespace,
	Component: "monitoring",
	PartOf:    appName,
	Version:   version,
	ManagedBy: "lingon",
}

type MonCoreDNS struct {
	kube.App

	SVC         *corev1.Service
	Scrape      *vmo.VMServiceScrape
	DashboardCM *corev1.ConfigMap
}

func NewMonCoreDNS() *MonCoreDNS {
	return &MonCoreDNS{
		SVC:         CoreDNSSVC,
		Scrape:      CoreDNSScrape,
		DashboardCM: CoreDNSDashboardCM,
	}
}

var CoreDNSSVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels:    CDNS.Labels(),
		Name:      CDNS.Name,
		Namespace: ku.NSKubeSystem,
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: corev1.ClusterIPNone,
		Ports: []corev1.ServicePort{
			{
				Name:       CDNSPortName,
				Port:       int32(CDNSPort),
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(CDNSPort),
			},
		},
		Selector: map[string]string{"k8s-app": "kube-dns"},
	},
	TypeMeta: ku.TypeServiceV1,
}

var CoreDNSScrape = &vmo.VMServiceScrape{
	TypeMeta:   TypeVMServiceScrapevmo,
	ObjectMeta: CDNS.ObjectMeta(),
	Spec: vmo.VMServiceScrapeSpec{
		Endpoints: []vmo.Endpoint{
			{
				BearerTokenFile: PathSA + "/token",
				Port:            CDNSPortName,
			},
		},
		NamespaceSelector: vmo.NamespaceSelector{
			MatchNames: []string{ku.NSKubeSystem}, // kube-system
		},
		// JobLabel: "k8s-app",
		JobLabel: ku.AppLabelName,
		Selector: metav1.LabelSelector{MatchLabels: CDNS.MatchLabels()},
	},
}
