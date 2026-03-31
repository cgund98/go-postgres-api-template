package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	appconfig "github.com/cgund98/go-postgres-api-template/internal/config"
)

// LoadAWSConfig creates an aws.Config using the v2 SDK.
// It respects localstack settings for local development.
func LoadAWSConfig(ctx context.Context, settings *appconfig.Config) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(settings.AwsRegion),
	}

	if settings.AwsUseLocalstack {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}
