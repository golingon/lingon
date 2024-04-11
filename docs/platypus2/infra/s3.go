// Copyright (c) 2023 Volvo Car Corporation
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"github.com/golingon/lingon/pkg/terra"
	"github.com/golingon/lingoneks/out/aws/aws_s3_bucket"
	"github.com/golingon/lingoneks/out/aws/aws_s3_bucket_public_access_block"
	"github.com/golingon/lingoneks/out/aws/aws_s3_bucket_server_side_encryption_configuration"
	"github.com/golingon/lingoneks/out/aws/aws_s3_bucket_versioning"
)

type Bucket struct {
	S3 *aws_s3_bucket.Resource `validate:"required"`
	// ACL          *aws.S3BucketAcl
	// `validate:"required"`
	Versioning   *aws_s3_bucket_versioning.Resource                           `validate:"required"`
	PublicAccess *aws_s3_bucket_public_access_block.Resource                  `validate:"required"`
	SSE          *aws_s3_bucket_server_side_encryption_configuration.Resource `validate:"required"`
}

func NewBucket(bucketName string) *Bucket {
	b := aws_s3_bucket.New(
		"s3", aws_s3_bucket.Args{
			Bucket: S(bucketName),
			Tags:   Stags("Name", "Lingon Experiment"),
		},
	)

	bucketID := b.Attributes().Id()

	// When bucket owner enforced is applied
	// use bucket policies to control access.
	// Otherwise, we get error: The bucket does not allow ACLs
	//
	// acl := aws.NewS3BucketAcl(
	// 	"s3", aws.S3BucketAclArgs{
	// 		Bucket: bucketID,
	// 		Acl:    S("private"),
	// 	},
	// )

	vv := aws_s3_bucket_versioning.New(
		"s3", aws_s3_bucket_versioning.Args{
			Bucket: bucketID,
			VersioningConfiguration: &aws_s3_bucket_versioning.VersioningConfiguration{
				Status: S("Enabled"),
			},
		},
	)

	pab := aws_s3_bucket_public_access_block.New(
		"s3", aws_s3_bucket_public_access_block.Args{
			Bucket:                bucketID,
			BlockPublicAcls:       terra.Bool(true),
			BlockPublicPolicy:     terra.Bool(true),
			IgnorePublicAcls:      terra.Bool(true),
			RestrictPublicBuckets: terra.Bool(true),
		},
	)

	enc := aws_s3_bucket_server_side_encryption_configuration.New(
		"s3", aws_s3_bucket_server_side_encryption_configuration.Args{
			Bucket: bucketID,
			Rule:   RuleEncryptKMS(),
		},
	)

	return &Bucket{
		S3: b,
		// ACL:          acl,
		Versioning:   vv,
		PublicAccess: pab,
		SSE:          enc,
	}
}

func RuleEncryptKMS() []aws_s3_bucket_server_side_encryption_configuration.Rule {
	return []aws_s3_bucket_server_side_encryption_configuration.Rule{
		{
			ApplyServerSideEncryptionByDefault: &aws_s3_bucket_server_side_encryption_configuration.RuleApplyServerSideEncryptionByDefault{
				SseAlgorithm: S("aws:kms"),
			},
		},
	}
}
