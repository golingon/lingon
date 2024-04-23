// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package terraform

import (
	"bytes"
	"fmt"
	"log/slog"

	tfjson "github.com/hashicorp/terraform-json"

	aws "github.com/golingon/lingon/docs/terraform/out/aws"
	"github.com/golingon/lingon/docs/terraform/out/aws/aws_subnet"
	"github.com/golingon/lingon/docs/terraform/out/aws/aws_vpc"
	"github.com/golingon/lingon/pkg/terra"
)

func Example_awsProvider() {
	type AWSStack struct {
		terra.Stack
		Provider *aws.Provider `validate:"required"`
	}

	// Initialise a stack with the AWS provider configuration
	_ = AWSStack{
		Provider: &aws.Provider{
			Region: terra.String("eu-north-1"),
		},
	}
}

func Example_awsVPC() {
	type AWSStack struct {
		terra.Stack
		Provider *aws.Provider     `validate:"required"`
		VPC      *aws_vpc.Resource `validate:"required"`
	}

	// Initialise a stack with the AWS provider configuration
	stack := AWSStack{
		Provider: &aws.Provider{
			Region: terra.String("eu-north-1"),
		},
		VPC: &aws_vpc.Resource{
			Name: "vpc",
			Args: aws_vpc.Args{
				CidrBlock:        terra.String("10.0.0.0/16"),
				EnableDnsSupport: terra.Bool(true),
			},
		},
	}
	// Export the stack to Terraform HCL
	var b bytes.Buffer
	if err := terra.Export(&stack, terra.WithExportWriter(&b)); err != nil {
		slog.Error("exporting stack", "err", err)
		return
	}
	fmt.Println(b.String())

	// Output:
	// terraform {
	//   required_providers {
	//     aws = {
	//       source  = "hashicorp/aws"
	//       version = "5.44.0"
	//     }
	//   }
	// }
	//
	// // Provider blocks
	// provider "aws" {
	//   region = "eu-north-1"
	// }
	//
	// // Resource blocks
	// resource "aws_vpc" "vpc" {
	//   cidr_block         = "10.0.0.0/16"
	//   enable_dns_support = true
	// }
}

func Example_awsVPCWithSubnet() {
	type AWSStack struct {
		terra.Stack
		Provider *aws.Provider        `validate:"required"`
		VPC      *aws_vpc.Resource    `validate:"required"`
		Subnet   *aws_subnet.Resource `validate:"required"`
	}

	vpc := aws_vpc.Resource{
		Name: "vpc",
		Args: aws_vpc.Args{
			CidrBlock:        terra.String("10.0.0.0/16"),
			EnableDnsSupport: terra.Bool(true),
		},
	}
	subnet := aws_subnet.Resource{
		Name: "subnet",
		Args: aws_subnet.Args{
			// Reference the VPC's ID, which will translate into a reference
			// in the Terraform configuration
			VpcId: vpc.Attributes().Id(),
		},
	}

	// Initialise a stack with the AWS provider configuration
	stack := AWSStack{
		Provider: &aws.Provider{
			Region: terra.String("eu-north-1"),
		},
		VPC:    &vpc,
		Subnet: &subnet,
	}
	// Export the stack to Terraform HCL
	var b bytes.Buffer
	if err := terra.Export(&stack, terra.WithExportWriter(&b)); err != nil {
		slog.Error("exporting stack", "err", err)
		return
	}
	fmt.Println(b.String())

	// Output:
	// terraform {
	//   required_providers {
	//     aws = {
	//       source  = "hashicorp/aws"
	//       version = "5.44.0"
	//     }
	//   }
	// }
	//
	// // Provider blocks
	// provider "aws" {
	//   region = "eu-north-1"
	// }
	//
	// // Resource blocks
	// resource "aws_vpc" "vpc" {
	//   cidr_block         = "10.0.0.0/16"
	//   enable_dns_support = true
	// }
	//
	// resource "aws_subnet" "subnet" {
	//   vpc_id = aws_vpc.vpc.id
	// }
}

func Example_awsVPCImportState() {
	type AWSStack struct {
		terra.Stack
		Provider *aws.Provider     `validate:"required"`
		VPC      *aws_vpc.Resource `validate:"required"`
	}

	// Initialise a stack with the AWS provider configuration
	stack := AWSStack{
		Provider: &aws.Provider{
			Region: terra.String("eu-north-1"),
		},
		VPC: &aws_vpc.Resource{
			Name: "vpc",
			Args: aws_vpc.Args{
				CidrBlock:        terra.String("10.0.0.0/16"),
				EnableDnsSupport: terra.Bool(true),
			},
		},
	}

	// At this point, you would invoke the Terraform CLI, and at a minimum
	// run the `terraform show` command to get the state back.
	// The state can then be decoded back into our stack.
	// For this test, we will create some dummy state data
	// (don't do this at home!)

	state := tfjson.State{
		Values: &tfjson.StateValues{
			RootModule: &tfjson.StateModule{
				Resources: []*tfjson.StateResource{
					{
						Type: "aws_vpc",
						Name: "vpc",
						AttributeValues: map[string]interface{}{
							"id": "12345",
						},
					},
				},
			},
		},
	}
	ok, err := terra.StackImportState(&stack, &state)
	if err != nil {
		slog.Error("importing stack state", "error", err)
	}
	if !ok {
		// This means the stack includes resources that did not have values
		// in the Terraform state.
		// This is happens if you have not applied your Terraform configuration.
		slog.Info("stack state is not complete")
	}
	// Access the VPC state. If you know the state is complete, you can also use
	// the StateMust() function
	vpcState, ok := stack.VPC.State()
	if !ok {
		slog.Info("vpc does not have state")
		return
	}
	fmt.Println(vpcState.Id)
	// Output: 12345
}
