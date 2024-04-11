// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"fmt"

	"github.com/golingon/lingon/pkg/terra"
	"github.com/golingon/lingoneks/out/aws/aws_eks_cluster"
	"github.com/golingon/lingoneks/out/aws/aws_iam_openid_connect_provider"
	"github.com/golingon/lingoneks/out/aws/aws_iam_policy_document"
	"github.com/golingon/lingoneks/out/aws/aws_iam_role"
	"github.com/golingon/lingoneks/out/aws/aws_iam_role_policy_attachment"
	"github.com/golingon/lingoneks/out/aws/aws_security_group"
	"github.com/golingon/lingoneks/out/aws/aws_security_group_rule"
	"github.com/golingon/terra_tls/tls_certificate"
)

var (
	arnClusterPolicy = S(
		"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
	)
	arnVPCResourceController = S(
		"arn:aws:iam::aws:policy/AmazonEKSVPCResourceController",
	)
	INGRESS = S("ingress")
	EGRESS  = S("egress")
)

type ClusterOpts struct {
	Name             string
	Version          string
	VPCID            string
	PrivateSubnetIDs [3]string
}

type Cluster struct {
	EKSCluster           *aws_eks_cluster.Resource                `validate:"required"`
	IAMPolicyDocument    *aws_iam_policy_document.DataSource      `validate:"required"`
	IAMRole              *aws_iam_role.Resource                   `validate:"required"`
	IAMRoleClusterPolicy *aws_iam_role_policy_attachment.Resource `validate:"required"`
	IAMRoleVPCController *aws_iam_role_policy_attachment.Resource `validate:"required"`

	// SecurityGroup is the AWS security group for both the EKS control plane
	// and worker nodes
	SecurityGroup   *aws_security_group.Resource      `validate:"required"`
	IngressAllowAll *aws_security_group_rule.Resource `validate:"required"`
	EgressAllowAll  *aws_security_group_rule.Resource `validate:"required"`

	TLSCert         *tls_certificate.DataSource               `validate:"required"`
	IAMOIDCProvider *aws_iam_openid_connect_provider.Resource `validate:"required"`
}

func NewCluster(opts ClusterOpts) *Cluster {
	sg := aws_security_group.New(
		"eks", aws_security_group.Args{
			Name: S("eks-" + opts.Name),
			Description: S(
				fmt.Sprintf(
					"Main security group for EKS cluster %s", opts.Name,
				),
			),
			VpcId: S(opts.VPCID),
			Tags: MergeSTags(
				MergeMaps(TFBaseTags, TFTags("platypus", "platform")),
				KarpenterDiscoveryKey, opts.Name,
			),
		},
	)

	sgAttrs := sg.Attributes()

	ingressAllowAll := aws_security_group_rule.New(
		"eks", aws_security_group_rule.Args{
			SecurityGroupId:       sgAttrs.Id(),
			SourceSecurityGroupId: sgAttrs.Id(),
			Description: S(
				"Allow all for EKS control plane and managed worker nodes",
			),
			Protocol: S("-1"),
			FromPort: N(0),
			ToPort:   N(0),
			Type:     INGRESS,
		},
	)
	egressAllowAll := aws_security_group_rule.New(
		"node_egress_all", aws_security_group_rule.Args{
			SecurityGroupId: sgAttrs.Id(),
			Description:     S("Allow all egress"),
			Protocol:        S("-1"),
			FromPort:        N(0),
			ToPort:          N(0),
			Type:            EGRESS,
			CidrBlocks:      terra.List(S("0.0.0.0/0")),
		},
	)

	iamPolicyDocument := aws_iam_policy_document.Data(
		"eks", aws_iam_policy_document.DataArgs{
			Statement: []aws_iam_policy_document.DataStatement{
				{
					Sid:     S("EKSClusterAssumeRole"),
					Actions: terra.Set(S("sts:AssumeRole")),
					Principals: []aws_iam_policy_document.DataStatementPrincipals{
						{
							Type:        S("Service"),
							Identifiers: terra.Set(S("eks.amazonaws.com")),
						},
					},
				},
			},
		},
	)

	iamRole := aws_iam_role.New(
		"eks", aws_iam_role.Args{
			Name:             S("eks-" + opts.Name),
			AssumeRolePolicy: iamPolicyDocument.Attributes().Json(),
		},
	)

	clusterPolicy := aws_iam_role_policy_attachment.New(
		"cluster_policy", aws_iam_role_policy_attachment.Args{
			PolicyArn: arnClusterPolicy,
			Role:      iamRole.Attributes().Name(),
		},
	)
	vpcController := aws_iam_role_policy_attachment.New(
		"vpc_controller", aws_iam_role_policy_attachment.Args{
			PolicyArn: arnVPCResourceController,
			Role:      iamRole.Attributes().Name(),
		},
	)

	eksCluster := aws_eks_cluster.New(
		"eks", aws_eks_cluster.Args{
			Name:    S(opts.Name),
			RoleArn: iamRole.Attributes().Arn(),
			VpcConfig: &aws_eks_cluster.VpcConfig{
				SecurityGroupIds: terra.Set(sgAttrs.Id()),
				SubnetIds:        terra.SetString(opts.PrivateSubnetIDs[:]...),
			},
			Version: S(opts.Version),
		},
	)
	eksCluster.DependsOn = terra.DependsOn(
		sg,
		iamRole,
		clusterPolicy,
		vpcController,
	)
	// How to add lifecycle to platform_version
	// eksCluster.Lifecycle = &terra.Lifecycle{
	// 	IgnoreChanges: terra.IgnoreChanges(
	// 		eksCluster.Attributes().PlatformVersion(),
	// 	),
	// }

	tlsCert := tls_certificate.Data(
		"eks", tls_certificate.DataArgs{
			Url: eksCluster.Attributes().
				Identity().
				Index(0).
				Oidc().
				Index(0).
				Issuer(),
		},
	)
	iamOIDCProvider := aws_iam_openid_connect_provider.New(
		"eks", aws_iam_openid_connect_provider.Args{
			ClientIdList: terra.Set(terra.String("sts.amazonaws.com")),
			ThumbprintList: terra.CastAsList(
				tlsCert.Attributes().
					Certificates().
					Splat().Sha1Fingerprint(),
			),
			Url: eksCluster.Attributes().
				Identity().
				Index(0).
				Oidc().
				Index(0).
				Issuer(),
		},
	)

	return &Cluster{
		EKSCluster:           eksCluster,
		IAMPolicyDocument:    iamPolicyDocument,
		IAMRole:              iamRole,
		IAMRoleClusterPolicy: clusterPolicy,
		IAMRoleVPCController: vpcController,

		SecurityGroup:   sg,
		IngressAllowAll: ingressAllowAll,
		EgressAllowAll:  egressAllowAll,

		TLSCert:         tlsCert,
		IAMOIDCProvider: iamOIDCProvider,
	}
}
