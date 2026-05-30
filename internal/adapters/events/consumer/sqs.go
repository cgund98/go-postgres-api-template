package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"

	mockaws "github.com/cgund98/go-postgres-api-template/internal/adapters/aws"
	"github.com/cgund98/go-postgres-api-template/internal/domain/events"
	"github.com/cgund98/go-postgres-api-template/internal/domain/events/registry"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

const (
	errorBackoff               = 1 * time.Second
	defaultMaxNumberOfMessages = 10
	defaultVisibilityTimeout   = 30
	defaultWaitTimeSeconds     = 20
)

type SQSConsumerOptions struct {
	QueueURL            string
	MaxNumberOfMessages *int32
	VisibilityTimeout   *int32
	WaitTimeSeconds     *int32
}

// SQSConsumer implements Consumer using AWS SQS
type SQSConsumer struct {
	queueURL            string
	sqsClient           mockaws.SQSClientInterface
	maxNumberOfMessages int32
	visibilityTimeout   int32
	waitTimeSeconds     int32
	logger              *slog.Logger
}

// NewSQSConsumer creates a new SQS consumer
func NewSQSConsumer(rootLogger *slog.Logger, sqsClient mockaws.SQSClientInterface, options SQSConsumerOptions) *SQSConsumer {
	var maxNumberOfMessages int32 = defaultMaxNumberOfMessages
	var visibilityTimeout int32 = defaultVisibilityTimeout
	var waitTimeSeconds int32 = defaultWaitTimeSeconds

	if options.MaxNumberOfMessages != nil {
		maxNumberOfMessages = *options.MaxNumberOfMessages
	}

	if options.VisibilityTimeout != nil {
		visibilityTimeout = *options.VisibilityTimeout
	}

	if options.WaitTimeSeconds != nil {
		waitTimeSeconds = *options.WaitTimeSeconds
	}

	logger := rootLogger.With("queue_url", options.QueueURL)

	return &SQSConsumer{
		queueURL:            options.QueueURL,
		maxNumberOfMessages: maxNumberOfMessages,
		visibilityTimeout:   visibilityTimeout,
		waitTimeSeconds:     waitTimeSeconds,
		sqsClient:           sqsClient,
		logger:              logger,
	}
}

// Ack deletes a message from SQS
func (c *SQSConsumer) Ack(ctx context.Context, messageID string) error {
	ctx, endSpan := observability.StartTraceFromContext(ctx, "Ack")
	defer endSpan()

	_, err := c.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(messageID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete sqs message: %w", err)
	}

	return nil
}

// BatchAck deletes a batch of messages from SQS
func (c *SQSConsumer) BatchAck(ctx context.Context, messageIDs []string) error {
	ctx, endSpan := observability.StartTraceFromContext(ctx, "BatchAck")
	defer endSpan()

	entries := make([]sqstypes.DeleteMessageBatchRequestEntry, len(messageIDs))
	for i, messageID := range messageIDs {
		entries[i] = sqstypes.DeleteMessageBatchRequestEntry{
			Id:            aws.String(messageID),
			ReceiptHandle: aws.String(messageID),
		}
	}

	_, err := c.sqsClient.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
		QueueUrl: aws.String(c.queueURL),
		Entries:  entries,
	})

	if err != nil {
		return fmt.Errorf("failed to delete sqs messages: %w", err)
	}

	return nil
}

// processBatchOfSingleMessages retrieves a batch of sqs messages from SQS
// and processes them one by one
func (c *SQSConsumer) processBatchOfSingleMessages(ctx context.Context, router *events.Router) {
	ctx, endSpan := observability.StartTraceFromContext(ctx, "processBatchOfSingleMessages")
	defer endSpan()

	message, err := c.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: c.maxNumberOfMessages,
		VisibilityTimeout:   c.visibilityTimeout,
		WaitTimeSeconds:     c.waitTimeSeconds,
	})
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to receive sqs messages", "error", err)
		time.Sleep(errorBackoff)
		return
	}

	if len(message.Messages) == 0 {
		return
	}

	for _, message := range message.Messages {
		var envelope registry.Envelope
		if message.Body == nil {
			c.logger.ErrorContext(ctx, "message body is nil")
			return
		}

		err := envelope.Unmarshal([]byte(*message.Body))
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to unmarshal envelope", "error", err)
			return
		}

		msgLogger := c.logger
		if message.MessageId != nil {
			msgLogger = msgLogger.With("message_id", *message.MessageId)
		}
		msgCtx := observability.SetLoggerOnContext(ctx, msgLogger)

		err = router.Route(msgCtx, envelope)
		if err != nil {
			c.logger.ErrorContext(msgCtx, "failed to route event", "error", err)
			return
		}

		err = c.Ack(msgCtx, *message.ReceiptHandle)
		if err != nil {
			c.logger.ErrorContext(msgCtx, "failed to ack sqs message", "error", err)
			return
		}
	}
}

// Start starts consuming messages from SQS. This will begin in a new goroutine and return immediately.
func (c *SQSConsumer) Start(ctx context.Context, router *events.Router) {
	go func() {
		c.logger.InfoContext(ctx, "starting sqs consumer")
		for {
			select {
			case <-ctx.Done():
				c.logger.InfoContext(ctx, "sqs consumer context canceled, stopping")
				return
			default:
				c.processBatchOfSingleMessages(ctx, router)
			}
		}
	}()
}

// Make sure the consumer implements the Consumer interface
var _ events.Consumer = &SQSConsumer{}
