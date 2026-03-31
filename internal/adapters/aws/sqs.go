package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSClientInterface defines the interface for SQS operations used by the consumer.
// We define an interface so we can mock the SQS client in tests.
type SQSClientInterface interface {
	ReceiveMessage(ctx context.Context, input *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	DeleteMessageBatch(ctx context.Context, input *sqs.DeleteMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error)
}
