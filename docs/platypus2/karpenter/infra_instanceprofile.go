// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package karpenter

import (
	"fmt"

	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_instance_profile"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_policy_document"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_role"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_role_policy_attachment"

	"github.com/golingon/lingon/pkg/terra"
)

// InstanceProfile is the AWS EC2 Instance Profile for the nodes provisioned by
// Karpenter to use.
type InstanceProfile struct {
	InstanceProfile   *aws_iam_instance_profile.Resource         `validate:"required"`
	IAMRole           *aws_iam_role.Resource                     `validate:"required"`
	AssumeRole        *aws_iam_policy_document.DataSource        `validate:"required"`
	PolicyAttachments []*aws_iam_role_policy_attachment.Resource `validate:"required,dive,required"`
}

func newInstanceProfile() InstanceProfile {
	arPolicy := aws_iam_policy_document.Data(
		"eks_node", aws_iam_policy_document.DataArgs{
			Statement: []aws_iam_policy_document.DataStatement{
				{
					Sid:     S("EKSNodeAssumeRole"),
					Effect:  S("Allow"),
					Actions: terra.SetString("sts:AssumeRole"),
					Principals: []aws_iam_policy_document.DataStatementPrincipals{
						{
							Type: S("Service"),
							Identifiers: terra.SetString(
								"ec2.amazonaws.com",
							),
						},
					},
				},
			},
		},
	)

	iamRole := aws_iam_role.New(
		"eks_node", aws_iam_role.Args{
			Name: S("platypus-karpenter-node"),
			Description: S(
				"IAM Role for Karpenter's InstanceProfile to use when launching nodes",
			),
			AssumeRolePolicy: arPolicy.Attributes().Json(),
		},
	)

	policies := []string{
		awsEKSWorkerNodePolicy,
		awsEC2ContainerRegistryReadOnly,
		awsEKSCNIPolicy,
		awsSSMManagedInstanceCore,
	}

	policyAttachments := make(
		[]*aws_iam_role_policy_attachment.Resource,
		len(policies),
	)
	for i, policy := range policies {
		policyAttachments[i] = aws_iam_role_policy_attachment.New(
			fmt.Sprintf("eks_node_attach_%s", policy),
			aws_iam_role_policy_attachment.Args{
				PolicyArn: S(awsPolicyARNPrefix + policy),
				Role:      iamRole.Attributes().Name(),
			},
		)
	}

	instanceProfile := aws_iam_instance_profile.New(
		KA.Name, aws_iam_instance_profile.Args{
			Name: S("platypus-karpenter-instance-profile"),
			Role: iamRole.Attributes().Name(),
		},
	)

	return InstanceProfile{
		InstanceProfile:   instanceProfile,
		IAMRole:           iamRole,
		AssumeRole:        arPolicy,
		PolicyAttachments: policyAttachments,
	}
}
