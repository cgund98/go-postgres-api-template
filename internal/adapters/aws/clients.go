package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cgund98/go-postgres-api-template/internal/config"
)

// NewClients creates new AWS SDK clients
func NewSession(settings *config.Config) (*session.Session, error) {
	cfg := &aws.Config{
		Region: aws.String(settings.AwsRegion),
	}

	if settings.AwsEndpoint != "" {
		cfg.Endpoint = aws.String(settings.AwsEndpoint)
	}

	if settings.AwsUseLocalstack {
		cfg.Credentials = credentials.NewStaticCredentials("test", "test", "")
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	return sess, nil
}
