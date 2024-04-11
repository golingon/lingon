// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"fmt"

	"github.com/golingon/lingon/pkg/terra"
	"github.com/golingon/lingoneks/out/aws/aws_eip"
	"github.com/golingon/lingoneks/out/aws/aws_internet_gateway"
	"github.com/golingon/lingoneks/out/aws/aws_nat_gateway"
	"github.com/golingon/lingoneks/out/aws/aws_route"
	"github.com/golingon/lingoneks/out/aws/aws_route_table"
	"github.com/golingon/lingoneks/out/aws/aws_route_table_association"
	"github.com/golingon/lingoneks/out/aws/aws_subnet"
	"github.com/golingon/lingoneks/out/aws/aws_vpc"
)

type Opts struct {
	Name               string
	AZs                [3]string
	CIDR               string
	PublicSubnetCIDRs  [3]string
	PrivateSubnetCIDRs [3]string
	KarpenterDiscovery string
	CommonTags         map[string]string
}

type AWSVPC struct {
	VPC *aws_vpc.Resource `validate:"required"`

	PublicSubnets  [3]*aws_subnet.Resource                  `validate:"required,dive,required"`
	PublicRT       *aws_route_table.Resource                `validate:"required"`
	PublicRoute    *aws_route.Resource                      `validate:"required"`
	PublicRTAssocs [3]*aws_route_table_association.Resource `validate:"required,dive,required"`

	PrivateSubnets  [3]*aws_subnet.Resource                  `validate:"required,dive,required"`
	PrivateRTs      [3]*aws_route_table.Resource             `validate:"required,dive,required"`
	PrivateRoutes   [3]*aws_route.Resource                   `validate:"required,dive,required"`
	PrivateRTAssocs [3]*aws_route_table_association.Resource `validate:"required,dive,required"`

	InternetGateway *aws_internet_gateway.Resource `validate:"required"`
	EIPNat          [3]*aws_eip.Resource           `validate:"required,dive,required"`
	NatGateways     [3]*aws_nat_gateway.Resource   `validate:"required,dive,required"`
}

func NewAWSVPC(opts Opts) *AWSVPC {
	name := opts.Name

	tags := func(
		name string,
		tags ...string,
	) terra.MapValue[terra.StringValue] {
		ss := []string{TagName, name}
		ss = append(ss, tags...)
		return MergeSTags(opts.CommonTags, ss...)
	}

	vpc := aws_vpc.New(
		name, aws_vpc.Args{
			CidrBlock: S(opts.CIDR),
			// Tags:             ttags(map[string]string{TagName: opts.Name}),
			InstanceTenancy:    S("default"),
			EnableDnsSupport:   B(true), // must be enabled for EFS
			EnableDnsHostnames: B(true), // must be enabled for EFS
			Tags:               tags(opts.Name),
		},
	)

	igw := aws_internet_gateway.New(
		name, aws_internet_gateway.Args{
			VpcId: vpc.Attributes().Id(),
			Tags:  tags(name + "-igw"),
		},
	)

	eipNats := [3]*aws_eip.Resource{}
	for i := 0; i < 3; i++ {
		eipNats[i] = aws_eip.New(
			fmt.Sprintf("nats_%d", i), aws_eip.Args{
				// Vpc:    B(true), // deprecated
				Domain: S("vpc"),
				Tags:   tags("nat-" + opts.AZs[i]),
			},
		)
	}

	publicSubnets := [3]*aws_subnet.Resource{}
	for i := 0; i < 3; i++ {
		publicSubnets[i] = aws_subnet.New(
			fmt.Sprintf("public_%d", i), aws_subnet.Args{
				VpcId:               vpc.Attributes().Id(),
				AvailabilityZone:    S(opts.AZs[i]),
				CidrBlock:           S(opts.PublicSubnetCIDRs[i]),
				MapPublicIpOnLaunch: terra.Bool(true),
				Tags:                tags(name + "-public"),
			},
		)
	}

	publicRT := aws_route_table.New(
		"public", aws_route_table.Args{
			VpcId: vpc.Attributes().Id(),
			Tags:  tags(name + "-public"),
		},
	)
	publicRoute := aws_route.New(
		"public", aws_route.Args{
			DestinationCidrBlock: Anywhere,
			RouteTableId:         publicRT.Attributes().Id(),
			GatewayId:            igw.Attributes().Id(),
		},
	)

	pubRTAssocs := [3]*aws_route_table_association.Resource{}
	for i := 0; i < 3; i++ {
		pubRTAssocs[i] = aws_route_table_association.New(
			fmt.Sprintf("public_%d", i), aws_route_table_association.Args{
				SubnetId:     publicSubnets[i].Attributes().Id(),
				RouteTableId: publicRT.Attributes().Id(),
			},
		)
	}

	natGateways := [3]*aws_nat_gateway.Resource{}
	for i := 0; i < 3; i++ {
		ng := aws_nat_gateway.New(
			fmt.Sprintf("nat_gateway_%d", i), aws_nat_gateway.Args{
				SubnetId:     publicSubnets[i].Attributes().Id(),
				AllocationId: eipNats[i].Attributes().Id(),
				Tags:         tags(fmt.Sprintf("ng-%d", i)),
			},
		)
		ng.DependsOn = terra.DependsOn(igw)
		natGateways[i] = ng
	}

	privateSubnets := [3]*aws_subnet.Resource{}
	for i := 0; i < 3; i++ {
		privateSubnets[i] = aws_subnet.New(
			fmt.Sprintf("private_%d", i), aws_subnet.Args{
				VpcId:            vpc.Attributes().Id(),
				AvailabilityZone: S(opts.AZs[i]),
				CidrBlock:        S(opts.PrivateSubnetCIDRs[i]),
				Tags: MergeSTags(
					opts.CommonTags,
					TagName, name+"-private",
					KarpenterDiscoveryKey, opts.KarpenterDiscovery,
				),
			},
		)
	}

	privateRTs := [3]*aws_route_table.Resource{}
	for i := 0; i < 3; i++ {
		privateRTs[i] = aws_route_table.New(
			fmt.Sprintf("private_%d", i), aws_route_table.Args{
				VpcId: vpc.Attributes().Id(),
				Tags:  tags(fmt.Sprintf("platypus-private-%d", i)),
			},
		)
	}
	privateRoutes := [3]*aws_route.Resource{}
	for i := 0; i < 3; i++ {
		privateRoutes[i] = aws_route.New(
			fmt.Sprintf("private_%d", i), aws_route.Args{
				RouteTableId:         privateRTs[i].Attributes().Id(),
				DestinationCidrBlock: Anywhere,
				NatGatewayId:         natGateways[i].Attributes().Id(),
			},
		)
	}

	privateRTAssocs := [3]*aws_route_table_association.Resource{}
	for i := 0; i < 3; i++ {
		privateRTAssocs[i] = aws_route_table_association.New(
			fmt.Sprintf("private_%d", i), aws_route_table_association.Args{
				SubnetId:     privateSubnets[i].Attributes().Id(),
				RouteTableId: privateRTs[i].Attributes().Id(),
			},
		)
	}
	return &AWSVPC{
		VPC:             vpc,
		InternetGateway: igw,
		EIPNat:          eipNats,
		PublicSubnets:   publicSubnets,
		PublicRT:        publicRT,
		PublicRoute:     publicRoute,
		PublicRTAssocs:  pubRTAssocs,

		NatGateways:     natGateways,
		PrivateSubnets:  privateSubnets,
		PrivateRTs:      privateRTs,
		PrivateRoutes:   privateRoutes,
		PrivateRTAssocs: privateRTAssocs,
	}
}
