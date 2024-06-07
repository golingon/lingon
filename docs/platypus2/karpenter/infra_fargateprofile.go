// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package karpenter

import (
	"fmt"

	"github.com/golingon/terraproviders/aws/5.45.0/aws_eks_fargate_profile"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_policy_document"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_role"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_role_policy_attachment"

	"github.com/golingon/lingon/pkg/terra"
)

const (
	awsPolicyARNPrefix                  = "arn:aws:iam::aws:policy/"
	awsEKSWorkerNodePolicy              = "AmazonEKSWorkerNodePolicy"
	awsEC2ContainerRegistryReadOnly     = "AmazonEC2ContainerRegistryReadOnly"
	awsEKSFargatePodExecutionRolePolicy = "AmazonEKSFargatePodExecutionRolePolicy"
	awsEKSCNIPolicy                     = "AmazonEKS_CNI_Policy"
	awsSSMManagedInstanceCore           = "AmazonSSMManagedInstanceCore"
)

// FargateProfile is the AWS EKS Fargate profile for the Karpenter pods to
// run on
type FargateProfile struct {
	FargateProfile    *aws_eks_fargate_profile.Resource          `validate:"required"`
	IAMRole           *aws_iam_role.Resource                     `validate:"required"`
	AssumeRole        *aws_iam_policy_document.DataSource        `validate:"required"`
	PolicyAttachments []*aws_iam_role_policy_attachment.Resource `validate:"required,dive,required"`
}

func newFargateProfile(opts InfraOpts) FargateProfile {
	arPolicy := aws_iam_policy_document.Data(
		"fargate", aws_iam_policy_document.DataArgs{
			Statement: []aws_iam_policy_document.DataStatement{
				{
					Effect:  S("Allow"),
					Actions: terra.SetString("sts:AssumeRole"),
					Principals: []aws_iam_policy_document.DataStatementPrincipals{
						{
							Type: S("Service"),
							Identifiers: terra.SetString(
								"eks-fargate-pods.amazonaws.com",
							),
						},
					},
				},
			},
		},
	)

	iamRole := aws_iam_role.New(
		"fargate", aws_iam_role.Args{
			Name: S(opts.Name + "-fargate"),
			Description: S(
				"IAM Role for Fargate profile for Karpenter pods to run",
			),
			AssumeRolePolicy: arPolicy.Attributes().Json(),
		},
	)

	policies := []string{
		awsEKSFargatePodExecutionRolePolicy,
		awsEKSCNIPolicy,
	}

	policyAttachments := make(
		[]*aws_iam_role_policy_attachment.Resource,
		len(policies),
	)
	for i, policy := range policies {
		policyAttachments[i] = aws_iam_role_policy_attachment.New(
			fmt.Sprintf("%s_attach_%s", "fargate", policy),
			aws_iam_role_policy_attachment.Args{
				PolicyArn: S(awsPolicyARNPrefix + policy),
				Role:      iamRole.Attributes().Name(),
			},
		)
	}

	fargateProfile := aws_eks_fargate_profile.New(
		KA.Name, aws_eks_fargate_profile.Args{
			ClusterName:         S(opts.ClusterName),
			FargateProfileName:  S(KA.Name),
			PodExecutionRoleArn: iamRole.Attributes().Arn(),
			SubnetIds:           terra.SetString(opts.PrivateSubnetIDs[:]...),
			Selector: []aws_eks_fargate_profile.Selector{
				{
					Namespace: S(KA.Namespace),
				},
			},
		},
	)
	return FargateProfile{
		FargateProfile:    fargateProfile,
		IAMRole:           iamRole,
		AssumeRole:        arPolicy,
		PolicyAttachments: policyAttachments,
	}
}
