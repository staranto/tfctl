// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package aws

import (
	"context"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
)

// options holds optional overrides for AWS config loading.
type options struct {
	profile string
	region  string
	retryer func() awsv2.Retryer
}

// Option customizes how AWS config is loaded.
// Default behavior (no options) inherits the shell environment and shared
// config chain (AWS_PROFILE, ~/.aws/config, ~/.aws/credentials, IMDS, etc.).
type Option func(*options)

// WithProfile sets the shared config profile. Defaults to AWS_PROFILE/env chain.
func WithProfile(profile string) Option {
	return func(o *options) { o.profile = profile }
}

// WithRegion sets the region override. Defaults to env/profile/metadata chain.
func WithRegion(region string) Option {
	return func(o *options) { o.region = region }
}

// Endpoint resolution is service-specific in AWS SDK v2.
// For S3, pass an option to NewS3 that sets Options.EndpointResolverV2.

// WithRetryer injects a custom retryer; if not set, SDK defaults are used.
func WithRetryer(newRetryer func() awsv2.Retryer) Option {
	return func(o *options) { o.retryer = newRetryer }
}

// LoadAWSConfig loads AWS SDK v2 config. By default it inherits the shell's
// AWS setup (AWS_PROFILE, shared config, env, IMDS). Options can override
// profile, region, and retryer without changing callers.
func LoadAWSConfig(ctx context.Context, opts ...Option) (awsv2.Config, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	var loadOpts []func(*config.LoadOptions) error
	if o.profile != "" {
		loadOpts = append(loadOpts, config.WithSharedConfigProfile(o.profile))
	}
	if o.region != "" {
		loadOpts = append(loadOpts, config.WithRegion(o.region))
	}
	if o.retryer != nil {
		loadOpts = append(loadOpts, config.WithRetryer(o.retryer))
	}

	return config.LoadDefaultConfig(ctx, loadOpts...)
}

// NewS3 constructs a v2 S3 client from the provided config. Additional service
// options can be supplied via optFns.
func NewS3(cfg awsv2.Config, optFns ...func(*s3v2.Options)) *s3v2.Client {
	return s3v2.NewFromConfig(cfg, optFns...)
}

// WithS3EndpointResolver allows callers to set the S3 EndpointResolverV2
// in a type-safe way when constructing the client.
func WithS3EndpointResolver(r s3v2.EndpointResolverV2) func(*s3v2.Options) {
	return func(o *s3v2.Options) {
		o.EndpointResolverV2 = r
	}
}
