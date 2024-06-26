// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by lingon. EDIT AS MUCH AS YOU LIKE.

package policy

import (
	ku "github.com/golingon/lingon/pkg/kubeutil"
	rbacv1 "k8s.io/api/rbac/v1"
)

var CRB = ku.BindClusterRole(
	CR.Name,
	W.SA,
	CR,
	W.Labels(),
)

var CR = &rbacv1.ClusterRole{
	TypeMeta:   ku.TypeClusterRoleV1,
	ObjectMeta: W.ObjectMetaNoNS(),
	Rules: []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"events"},
			Verbs:     []string{"create", "patch"},
		}, {
			APIGroups: []string{"admissionregistration.k8s.io"},
			Resources: []string{
				"validatingwebhookconfigurations",
				"mutatingwebhookconfigurations",
			},
			Verbs: []string{"list", "watch"},
		}, {
			APIGroups: []string{"admissionregistration.k8s.io"},
			ResourceNames: []string{
				// "policy.sigstore.dev",
				ValidatingPolicySigstoreDevVWC.Name,
				// "defaulting.clusterimagepolicy.sigstore.dev",
				DefaultingClusterImagePolicyMWC.Name,
				// "validating.clusterimagepolicy.sigstore.dev",
				ValidatingClusterImagePolicyVWC.Name,
			},
			Resources: []string{
				"validatingwebhookconfigurations",
				"mutatingwebhookconfigurations",
			},
			Verbs: []string{"get", "update", "delete"},
		}, {
			APIGroups:     []string{""},
			ResourceNames: []string{W.Namespace},
			Resources:     []string{"namespaces"},
			Verbs:         []string{"get", "list"},
		}, {
			APIGroups:     []string{""},
			ResourceNames: []string{W.Namespace},
			Resources:     []string{"namespaces/finalizers"},
			Verbs:         []string{"update"},
		}, {
			APIGroups: []string{""},
			Resources: []string{"serviceaccounts", "secrets"},
			Verbs:     []string{"get"},
		}, {
			APIGroups: []string{"policy.sigstore.dev"},
			Resources: []string{
				"clusterimagepolicies",
				"clusterimagepolicies/status",
			},
			Verbs: []string{"get", "list", "update", "watch", "patch"},
		}, {
			APIGroups: []string{"apiextensions.k8s.io"},
			Resources: []string{"customresourcedefinitions"},
			Verbs:     []string{"get", "list", "watch", "update"},
		}, {
			APIGroups:     []string{"apiextensions.k8s.io"},
			ResourceNames: []string{"trustroots.policy.sigstore.dev"},
			Resources:     []string{"customresourcedefinitions"},
			Verbs:         []string{"get", "update", "list"},
		}, {
			APIGroups: []string{"policy.sigstore.dev"},
			Resources: []string{"trustroots", "trustroots/status"},
			Verbs:     []string{"get", "list", "update", "watch", "patch"},
		},
	},
}

var RB = ku.BindRole(
	Role.Name,
	W.SA,
	Role,
	W.Labels(),
)

var Role = &rbacv1.Role{
	TypeMeta:   ku.TypeRoleV1,
	ObjectMeta: W.ObjectMeta(),
	Rules: []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"configmaps", "secrets"},
			Verbs:     []string{"get", "list", "update", "watch"},
		}, {
			APIGroups: []string{"coordination.k8s.io"},
			Resources: []string{"leases"},
			Verbs: []string{
				"get",
				"list",
				"create",
				"update",
				"delete",
				"patch",
				"watch",
			},
		}, {
			APIGroups:     []string{""},
			ResourceNames: []string{ConfigImagePoliciesCM.Name},
			Resources:     []string{"configmaps"},
			Verbs: []string{
				"get",
				"list",
				"create",
				"update",
				"patch",
				"watch",
			},
		}, {
			APIGroups: []string{""},
			// ResourceNames: []string{"config-sigstore-keys"},
			ResourceNames: []string{ConfigSigstoreKeysCM.Name},
			Resources:     []string{"configmaps"},
			Verbs: []string{
				"get",
				"list",
				"create",
				"update",
				"patch",
				"watch",
			},
		}, {
			APIGroups: []string{"policy.sigstore.dev"},
			Resources: []string{"trustroots"},
			Verbs:     []string{"get", "list"},
		},
	},
}
