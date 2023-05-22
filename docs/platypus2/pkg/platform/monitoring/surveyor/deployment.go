// Code generated by lingon. EDIT AS MUCH AS YOU LIKE.

package surveyor

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var Deploy = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{
			"app.kubernetes.io/instance":   "surveyor",
			"app.kubernetes.io/managed-by": "Helm",
			"app.kubernetes.io/name":       "surveyor",
			"app.kubernetes.io/version":    "0.5.0",
			"helm.sh/chart":                "surveyor-0.16.2",
		},
		Name: "surveyor",
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: P(int32(1)),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/instance": "surveyor",
				"app.kubernetes.io/name":     "surveyor",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance": "surveyor",
					"app.kubernetes.io/name":     "surveyor",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Args: []string{
							"-p",
							"7777",
							"-s=nats://nats:4222",
							"--timeout=3s",
							"-c=1",
						},
						Image:           "natsio/nats-surveyor:0.5.0",
						ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz",
									Port: intstr.IntOrString{
										StrVal: "http",
										Type:   intstr.Type(int64(1)),
									},
								},
							},
						},
						Name: "surveyor",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: int32(7777),
								Name:          "http",
								Protocol:      corev1.Protocol("TCP"),
							},
						},
					},
				},
				ServiceAccountName: "surveyor",
			},
		},
	},
	TypeMeta: metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	},
}
