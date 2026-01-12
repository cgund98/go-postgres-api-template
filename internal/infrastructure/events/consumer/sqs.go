package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	mockaws "github.com/cgund98/go-postgres-api-template/internal/infrastructure/aws"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events/deserializer"
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
	MaxNumberOfMessages *int64
	VisibilityTimeout   *int64
	WaitTimeSeconds     *int64
}

// SQSConsumer implements Consumer using AWS SQS
type SQSConsumer[T events.Event] struct {
	queueURL            string
	sqsClient           mockaws.SQSClientInterface
	maxNumberOfMessages int64
	visibilityTimeout   int64
	waitTimeSeconds     int64
	logger              *slog.Logger
}

// NewSQSConsumer creates a new SQS consumer
func NewSQSConsumer[T events.Event](sqsClient mockaws.SQSClientInterface, options SQSConsumerOptions) *SQSConsumer[T] {
	var maxNumberOfMessages = int64(defaultMaxNumberOfMessages)
	var visibilityTimeout = int64(defaultVisibilityTimeout)
	var waitTimeSeconds = int64(defaultWaitTimeSeconds)

	if options.MaxNumberOfMessages != nil {
		maxNumberOfMessages = *options.MaxNumberOfMessages
	}

	if options.VisibilityTimeout != nil {
		visibilityTimeout = *options.VisibilityTimeout
	}

	if options.WaitTimeSeconds != nil {
		waitTimeSeconds = *options.WaitTimeSeconds
	}

	logger := observability.Logger.With("queueURL", options.QueueURL)

	return &SQSConsumer[T]{
		queueURL:            options.QueueURL,
		maxNumberOfMessages: maxNumberOfMessages,
		visibilityTimeout:   visibilityTimeout,
		waitTimeSeconds:     waitTimeSeconds,
		sqsClient:           sqsClient,
		logger:              logger,
	}
}

// Ack deletes a message from SQS
func (c *SQSConsumer[T]) Ack(_ context.Context, messageID string) error {

	_, err := c.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(messageID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete sqs message: %w", err)
	}

	return nil
}

// BatchAck deletes a batch of messages from SQS
func (c *SQSConsumer[T]) BatchAck(_ context.Context, messageIDs []string) error {
	entries := make([]*sqs.DeleteMessageBatchRequestEntry, len(messageIDs))
	for i, messageID := range messageIDs {
		entries[i] = &sqs.DeleteMessageBatchRequestEntry{
			Id:            aws.String(messageID),
			ReceiptHandle: aws.String(messageID),
		}
	}

	_, err := c.sqsClient.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
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
func (c *SQSConsumer[T]) processBatchOfSingleMessages(ctx context.Context, deserializer deserializer.Deserializer[T], handler events.Handler[T]) {
	message, err := c.sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: aws.Int64(c.maxNumberOfMessages),
		VisibilityTimeout:   aws.Int64(c.visibilityTimeout),
		WaitTimeSeconds:     aws.Int64(c.waitTimeSeconds),
	})
	if err != nil {
		c.logger.Error("failed to receive sqs messages", "error", err)
		time.Sleep(errorBackoff)
		return
	}

	if len(message.Messages) == 0 {
		return
	}

	for _, message := range message.Messages {
		event, err := deserializer.Deserialize([]byte(*message.Body))
		if err != nil {
			c.logger.Error("failed to deserialize event", "error", err)
			return
		}

		err = handler.Handle(ctx, event)
		if err != nil {
			c.logger.Error("failed to handle event", "error", err)
			return
		}

		err = c.Ack(ctx, *message.ReceiptHandle)
		if err != nil {
			c.logger.Error("failed to ack sqs message", "error", err)
			return
		}
	}

}

// Start starts consuming messages from SQS. This will begin in a new goroutine and return immediately.
func (c *SQSConsumer[T]) Start(ctx context.Context, deserializer deserializer.Deserializer[T], handler events.Handler[T]) {
	go func() {
		c.logger.Info("starting sqs consumer")
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("sqs consumer context canceled, stopping")
				return
			default:
				c.processBatchOfSingleMessages(ctx, deserializer, handler)
			}
		}
	}()
}

func (c *SQSConsumer[T]) processBatchOfMessages(ctx context.Context, deserializer deserializer.Deserializer[T], handler events.BatchHandler[T]) {
	message, err := c.sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: aws.Int64(c.maxNumberOfMessages),
		VisibilityTimeout:   aws.Int64(c.visibilityTimeout),
		WaitTimeSeconds:     aws.Int64(c.waitTimeSeconds),
	})

	if err != nil {
		c.logger.Error("failed to receive sqs messages", "error", err)
		time.Sleep(errorBackoff)
		return
	}

	if len(message.Messages) == 0 {
		return
	}

	// Deserialize the messages into events
	events := make([]T, 0, len(message.Messages))
	for _, message := range message.Messages {
		event, err := deserializer.Deserialize([]byte(*message.Body))
		if err != nil {
			c.logger.Error("failed to deserialize event", "error", err)
			return
		}
		events = append(events, event)
	}

	// Handle the events
	err = handler.HandleBatch(ctx, events)
	if err != nil {
		c.logger.Error("failed to handle events", "error", err)
		return
	}

	// Ack the messages
	for _, message := range message.Messages {
		err = c.Ack(ctx, *message.ReceiptHandle)
		if err != nil {
			c.logger.Error("failed to ack sqs message", "error", err)
			return
		}
	}

}

// StartBatch starts consuming messages from SQS in batch mode. This will begin in a new goroutine and return immediately.
func (c *SQSConsumer[T]) StartBatch(ctx context.Context, deserializer deserializer.Deserializer[T], handler events.BatchHandler[T]) {
	go func() {
		c.logger.Info("starting sqs consumer in batch mode", "queue_url", c.queueURL)
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("sqs consumer context canceled, stopping")
				return
			default:
				c.processBatchOfMessages(ctx, deserializer, handler)
			}
		}
	}()
}

// Make sure the consumer implements the Consumer interface
var _ Consumer[events.Event] = &SQSConsumer[events.Event]{}
