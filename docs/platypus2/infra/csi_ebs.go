// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"github.com/golingon/lingon/pkg/terra"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_eks_addon"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_policy_document"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_role"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_role_policy_attachment"
)

type CSI struct {
	CSIDriver *aws_eks_addon.Resource `validate:"required"`
	IAMRole   `validate:"required"`
}

type CSIOpts struct {
	ClusterName     string
	OIDCProviderArn string
	OIDCProviderURL string
}

type IAMRole struct {
	AssumeRolePolicy *aws_iam_policy_document.DataSource      `validate:"required"`
	Role             *aws_iam_role.Resource                   `validate:"required"`
	RolePolicy       *aws_iam_policy_document.DataSource      `validate:"required"`
	PolicyAttach     *aws_iam_role_policy_attachment.Resource `validate:"required"`
}

func NewCSIEBS(opts CSIOpts) *CSI {
	ir := newIAMRole(opts)
	return &CSI{
		CSIDriver: aws_eks_addon.New(
			opts.ClusterName+"-csiebs", aws_eks_addon.Args{
				AddonName: S("aws-ebs-csi-driver"),
				// AddonVersion:             S("v1.19.0-eksbuild.1"),
				AddonVersion:             S("v1.21.0-eksbuild.1"),
				ClusterName:              S(opts.ClusterName),
				ServiceAccountRoleArn:    ir.Role.Attributes().Arn(),
				ResolveConflictsOnCreate: S("OVERWRITE"),
				ResolveConflictsOnUpdate: S("PRESERVE"),
			},
		),
		IAMRole: *ir,
	}
}

func newIAMRole(opts CSIOpts) *IAMRole {
	assumeRolePolicy := aws_iam_policy_document.Data(
		"csi_assume_role", aws_iam_policy_document.DataArgs{
			Statement: []aws_iam_policy_document.DataStatement{
				{
					Actions: terra.Set(S("sts:AssumeRoleWithWebIdentity")),
					Effect:  S("Allow"),

					Condition: []aws_iam_policy_document.DataStatementCondition{
						{
							Test:     S("StringEquals"),
							Variable: S(opts.OIDCProviderURL + ":sub"),
							Values: terra.List(
								S(
									"system:serviceaccount:kube-system:ebs-csi-controller-sa",
								),
							),
						},
						{
							Test:     S("StringEquals"),
							Variable: S(opts.OIDCProviderURL + ":aud"),
							Values:   terra.ListString("sts.amazonaws.com"),
						},
					},
					Principals: []aws_iam_policy_document.DataStatementPrincipals{
						{
							Type:        S("Federated"),
							Identifiers: terra.Set(S(opts.OIDCProviderArn)),
						},
					},
				},
			},
		},
	)

	// small utility function to avoid repeting fields in the policy
	cond := func(action, v, val string) aws_iam_policy_document.DataStatement {
		return aws_iam_policy_document.DataStatement{
			Effect:    S("Allow"),
			Actions:   terra.SetString(action),
			Resources: terra.SetString("*"),
			Condition: []aws_iam_policy_document.DataStatementCondition{
				{
					Test:     S("StringLike"),
					Variable: S(v),
					Values:   terra.ListString(val),
				},
			},
		}
	}

	// converted from
	// https://github.com/kubernetes-sigs/aws-ebs-csi-driver/blob/master/docs/example-iam-policy.json
	//
	policy := aws_iam_policy_document.Data(
		"csiebs", aws_iam_policy_document.DataArgs{
			Statement: []aws_iam_policy_document.DataStatement{
				{
					Effect: S("Allow"),
					Actions: terra.SetString(
						"ec2:CreateSnapshot",
						"ec2:AttachVolume",
						"ec2:DetachVolume",
						"ec2:ModifyVolume",
						"ec2:DescribeAvailabilityZones",
						"ec2:DescribeInstances",
						"ec2:DescribeSnapshots",
						"ec2:DescribeTags",
						"ec2:DescribeVolumes",
						"ec2:DescribeVolumesModifications",
					),
					Resources: terra.SetString("*"),
				},
				{
					Effect:  S("Allow"),
					Actions: terra.SetString("ec2:CreateTags"),
					Resources: terra.SetString(
						"arn:aws:ec2:*:*:volume/*",
						"arn:aws:ec2:*:*:snapshot/*",
					),
					Condition: []aws_iam_policy_document.DataStatementCondition{
						{
							Test:     S("StringEquals"),
							Variable: S("ec2:CreateAction"),
							Values: terra.ListString(
								"CreateVolume",
								"CreateSnapshot",
							),
						},
					},
				},
				{
					Effect:  S("Allow"),
					Actions: terra.SetString("ec2:DeleteTags"),
					Resources: terra.SetString(
						"arn:aws:ec2:*:*:volume/*",
						"arn:aws:ec2:*:*:snapshot/*",
					),
				},
				cond(
					"ec2:CreateVolume",
					"aws:RequestTag/ebs.csi.aws.com/cluster", "true",
				),
				cond(
					"ec2:CreateVolume",
					"aws:RequestTag/CSIVolumeName", "*",
				),
				cond(
					"ec2:DeleteVolume",
					"ec2:ResourceTag/ebs.csi.aws.com/cluster", "true",
				),
				cond(
					"ec2:DeleteVolume",
					"ec2:ResourceTag/CSIVolumeName", "*",
				),
				cond(
					"ec2:DeleteVolume",
					"ec2:ResourceTag/kubernetes.io/created-for/pvc/name", "*",
				),
				cond(
					"ec2:DeleteSnapshot",
					"ec2:ResourceTag/CSIVolumeSnapshotName", "*",
				),
				cond(
					"ec2:DeleteSnapshot",
					"ec2:ResourceTag/ebs.csi.aws.com/cluster", "true",
				),
			},
		},
	)

	csiRole := aws_iam_role.New(
		"csiebs_role", aws_iam_role.Args{
			Name:             S(opts.ClusterName + "-csi"),
			Description:      S("IAM Role for CSI EBS driver"),
			AssumeRolePolicy: assumeRolePolicy.Attributes().Json(),

			InlinePolicy: []aws_iam_role.InlinePolicy{
				{
					Name:   S("csi-ebs-driver"),
					Policy: policy.Attributes().Json(),
				},
			},
		},
	)
	pa := aws_iam_role_policy_attachment.New(
		"csiebs_attach_AmazonEBSCSIDriverPolicy",
		aws_iam_role_policy_attachment.Args{
			PolicyArn: S(
				"arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy",
			),
			Role: csiRole.Attributes().Name(),
		},
	)

	return &IAMRole{
		AssumeRolePolicy: assumeRolePolicy,
		Role:             csiRole,
		RolePolicy:       policy,
		PolicyAttach:     pa,
	}
}
