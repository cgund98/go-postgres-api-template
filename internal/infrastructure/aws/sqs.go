package aws

import "github.com/aws/aws-sdk-go/service/sqs"

// SQSClientInterface defines the interface for SQS operations used by the consumer
// We define an interface so we can mock the SQS client in tests
type SQSClientInterface interface {
	ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
	DeleteMessageBatch(input *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error)
}
