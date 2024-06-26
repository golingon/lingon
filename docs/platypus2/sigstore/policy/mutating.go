// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by lingon. EDIT AS MUCH AS YOU LIKE.

package policy

import (
	ku "github.com/golingon/lingon/pkg/kubeutil"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultingClusterImagePolicyMWC = &admissionregistrationv1.MutatingWebhookConfiguration{
	TypeMeta:   ku.TypeMutatingWebhookConfigurationV1,
	ObjectMeta: metav1.ObjectMeta{Name: "defaulting.clusterimagepolicy.sigstore.dev"},
	Webhooks: []admissionregistrationv1.MutatingWebhook{
		{
			AdmissionReviewVersions: []string{"v1"},
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				Service: &admissionregistrationv1.ServiceReference{
					Name: SVC.Name, Namespace: SVC.Namespace,
				},
			},
			FailurePolicy: P(admissionregistrationv1.Fail),
			MatchPolicy:   P(admissionregistrationv1.Equivalent),
			Name:          "defaulting.clusterimagepolicy.sigstore.dev",
			SideEffects:   P(admissionregistrationv1.SideEffectClassNone),
		},
	},
}

var MutatingPolicySigstoreDevMWC = &admissionregistrationv1.MutatingWebhookConfiguration{
	TypeMeta:   ku.TypeMutatingWebhookConfigurationV1,
	ObjectMeta: metav1.ObjectMeta{Name: PolicySigstoreDev},
	Webhooks: []admissionregistrationv1.MutatingWebhook{
		{
			AdmissionReviewVersions: []string{"v1"},
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				Service: &admissionregistrationv1.ServiceReference{
					Name: SVC.Name, Namespace: SVC.Namespace,
				},
			},
			FailurePolicy: P(admissionregistrationv1.Fail),
			Name:          PolicySigstoreDev,
			NamespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "policy.sigstore.dev/include",
						Operator: metav1.LabelSelectorOperator("In"),
						Values:   []string{"true"},
					},
				},
			},
			ReinvocationPolicy: P(admissionregistrationv1.IfNeededReinvocationPolicy),
			SideEffects:        P(admissionregistrationv1.SideEffectClassNone),
		},
	},
}
