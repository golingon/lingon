// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package karpenter

import (
	"fmt"

	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_policy_document"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_iam_role"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_sqs_queue"
	"github.com/golingon/terraproviders/aws/5.45.0/aws_sqs_queue_policy"

	"github.com/golingon/lingon/pkg/terra"
)

var S = terra.String

type InfraOpts struct {
	Name             string
	ClusterName      string
	ClusterARN       string
	PrivateSubnetIDs [3]string
	OIDCProviderArn  string
	OIDCProviderURL  string
}

// Infra is all the cloud resources for Karpenter to run on a pre-existing EKS
// cluster with an OIDC provider
type Infra struct {
	FargateProfile
	InstanceProfile
	Controller
}

// Controller contains all the resources for the Karpenter controller.
// This includes the IAM role used to manage EC2 nodes (and more) via IRSA
// (IAM Roles for ServiceAccounts) and the SQS (Simple Queue Service) for being
// notified about nodes being terminated (typically spot instances).
type Controller struct {
	NodeTerminationQueue
	IAMRole
}

type NodeTerminationQueue struct {
	SimpleQueue         *aws_sqs_queue.Resource             `validate:"required"`
	QueuePolicy         *aws_sqs_queue_policy.Resource      `validate:"required"`
	QueuePolicyDocument *aws_iam_policy_document.DataSource `validate:"required"`
}

func NewInfra(opts InfraOpts) Infra {
	ip := newInstanceProfile()
	return Infra{
		FargateProfile:  newFargateProfile(opts),
		InstanceProfile: ip,
		Controller:      newController(opts, ip.IAMRole),
	}
}

func newController(opts InfraOpts, ipRole *aws_iam_role.Resource) Controller {
	queue := newNodeTerminationQueue(opts)
	return Controller{
		IAMRole:              newIAMRole(opts, ipRole, queue.SimpleQueue),
		NodeTerminationQueue: queue,
	}
}

type IAMRole struct {
	AssumeRolePolicy *aws_iam_policy_document.DataSource `validate:"required"`
	Role             *aws_iam_role.Resource              `validate:"required"`
	RolePolicy       *aws_iam_policy_document.DataSource `validate:"required"`
}

func newIAMRole(
	opts InfraOpts,
	ipRole *aws_iam_role.Resource,
	queue *aws_sqs_queue.Resource,
) IAMRole {
	assumeRolePolicy := aws_iam_policy_document.Data(
		KA.Name+"_assume_role", aws_iam_policy_document.DataArgs{
			Statement: []aws_iam_policy_document.DataStatement{
				{
					Actions: terra.Set(S("sts:AssumeRoleWithWebIdentity")),
					Principals: []aws_iam_policy_document.DataStatementPrincipals{
						{
							Type:        S("Federated"),
							Identifiers: terra.Set(S(opts.OIDCProviderArn)),
						},
					},
					Condition: []aws_iam_policy_document.DataStatementCondition{
						{
							Test:     S("StringEquals"),
							Variable: S(opts.OIDCProviderURL + ":sub"),
							Values: terra.ListString(
								fmt.Sprintf(
									"system:serviceaccount:%s:%s",
									KA.Namespace,
									KA.Name,
								),
							),
						},
						{
							Test:     S("StringEquals"),
							Variable: S(opts.OIDCProviderURL + ":aud"),
							Values:   terra.ListString("sts.amazonaws.com"),
						},
					},
				},
			},
		},
	)
	policy := aws_iam_policy_document.Data(
		KA.Name, aws_iam_policy_document.DataArgs{
			Statement: []aws_iam_policy_document.DataStatement{
				{
					Actions: terra.SetString(
						"ec2:DescribeImages",
						"ec2:RunInstances",
						"ec2:DescribeSubnets",
						"ec2:DescribeSecurityGroups",
						"ec2:DescribeLaunchTemplates",
						"ec2:DescribeInstances",
						"ec2:DescribeInstanceTypes",
						"ec2:DescribeInstanceTypeOfferings",
						"ec2:DescribeAvailabilityZones",
						"ec2:DeleteLaunchTemplate",
						"ec2:CreateTags",
						"ec2:CreateLaunchTemplate",
						"ec2:CreateFleet",
						"ec2:DescribeSpotPriceHistory",
						"pricing:GetProducts",
						"ssm:GetParameter",
					),
					Effect:    S("Allow"),
					Resources: terra.SetString("*"),
				},
				{
					Actions: terra.SetString(
						"ec2:TerminateInstances",
						"ec2:DeleteLaunchTemplate",
					),
					Effect:    S("Allow"),
					Resources: terra.SetString("*"),
					Condition: []aws_iam_policy_document.DataStatementCondition{
						{
							Test: S("StringEquals"),
							Variable: S(
								"ec2:ResourceTag/karpenter.sh/discovery",
							),
							Values: terra.ListString(opts.ClusterName),
						},
					},
				},
				{
					Actions:   terra.SetString("eks:DescribeCluster"),
					Effect:    S("Allow"),
					Resources: terra.SetString(opts.ClusterARN),
				},

				// The Karpenter IRSA role has to have permission to pass on the
				// InstanceProfile IAM Role
				{
					Actions:   terra.SetString("iam:PassRole"),
					Effect:    S("Allow"),
					Resources: terra.Set(ipRole.Attributes().Arn()),
				},

				// For AWS SQS spot interruption queue
				{
					Actions: terra.SetString(
						"sqs:DeleteMessage",
						"sqs:GetQueueUrl",
						"sqs:GetQueueAttributes",
						"sqs:ReceiveMessage",
					),
					Effect:    S("Allow"),
					Resources: terra.Set(queue.Attributes().Arn()),
				},
			},
		},
	)
	role := aws_iam_role.New(
		KA.Name, aws_iam_role.Args{
			Name: S(opts.Name + "-controller"),
			Description: S(
				"IAM Role for Karpenter Controller (pod) to assume",
			),
			AssumeRolePolicy: assumeRolePolicy.Attributes().Json(),

			InlinePolicy: []aws_iam_role.InlinePolicy{
				{
					Name:   S(KA.Name),
					Policy: policy.Attributes().Json(),
				},
			},
		},
	)
	return IAMRole{
		AssumeRolePolicy: assumeRolePolicy,
		Role:             role,
		RolePolicy:       policy,
	}
}

func newNodeTerminationQueue(opts InfraOpts) NodeTerminationQueue {
	queue := aws_sqs_queue.New(
		KA.Name, aws_sqs_queue.Args{
			Name:                    S(opts.Name),
			MessageRetentionSeconds: terra.Number(300),
		},
	)
	policyDoc := aws_iam_policy_document.Data(
		"node_termination_queue", aws_iam_policy_document.DataArgs{
			Statement: []aws_iam_policy_document.DataStatement{
				{
					Sid:       S("SQSWrite"),
					Resources: terra.Set(queue.Attributes().Arn()),
					Actions:   terra.SetString("sqs:SendMessage"),
					Principals: []aws_iam_policy_document.DataStatementPrincipals{
						{
							Type: S("Service"),
							Identifiers: terra.SetString(
								"events.amazonaws.com",
								"sqs.amazonaws.com",
							),
						},
					},
				},
			},
		},
	)
	queuePolicy := aws_sqs_queue_policy.New(
		KA.Name, aws_sqs_queue_policy.Args{
			QueueUrl: queue.Attributes().Url(),
			Policy:   policyDoc.Attributes().Json(),
		},
	)
	return NodeTerminationQueue{
		SimpleQueue:         queue,
		QueuePolicy:         queuePolicy,
		QueuePolicyDocument: policyDoc,
	}
}
