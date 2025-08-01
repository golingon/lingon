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
	KPPort     = 10249
	KPPortName = "http-metrics"
)

var KP = &meta.Metadata{
	Name:      "kube-proxy",
	Namespace: namespace,
	Instance:  "kube-proxy-" + namespace,
	Component: "monitoring",
	PartOf:    appName,
	Version:   version,
	ManagedBy: "lingon",
}

type MonKubeProxy struct {
	kube.App

	KubeProxySVC    *corev1.Service
	KubeProxyScrape *vmo.VMServiceScrape
}

func NewMonKubeProxy() *MonKubeProxy {
	return &MonKubeProxy{
		KubeProxySVC:    KubeProxySVC,
		KubeProxyScrape: KubeProxyScrape,
	}
}

var KubeProxySVC = &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Labels:    KP.Labels(),
		Name:      KP.Name,
		Namespace: ku.NSKubeSystem,
	},
	Spec: corev1.ServiceSpec{
		ClusterIP: corev1.ClusterIPNone,
		Ports: []corev1.ServicePort{
			{
				Name:       KPPortName,
				Port:       int32(KPPort),
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(KPPort),
			},
		},
		Selector: map[string]string{"k8s-app": "kube-proxy"},
		Type:     corev1.ServiceTypeClusterIP,
	},
	TypeMeta: ku.TypeServiceV1,
}

var KubeProxyScrape = &vmo.VMServiceScrape{
	ObjectMeta: KP.ObjectMeta(),
	Spec: vmo.VMServiceScrapeSpec{
		Endpoints: []vmo.Endpoint{
			{
				BearerTokenFile: PathSA + "/token",
				Port:            KPPortName,
				Scheme:          "https",
				TLSConfig:       &vmo.TLSConfig{CAFile: PathSA + "/ca.crt"},
			},
		},
		JobLabel: "k8s-app",
		NamespaceSelector: vmo.NamespaceSelector{
			MatchNames: []string{ku.NSKubeSystem}, // kube-system
		},
		Selector: metav1.LabelSelector{
			MatchLabels: map[string]string{"k8s-app": "kube-proxy"},
		},
	},
	TypeMeta: TypeVMServiceScrapevmo,
}
