package publisher

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/uuid"

	"github.com/cgund98/go-postgres-api-template/internal/domain/events"
	"github.com/cgund98/go-postgres-api-template/internal/domain/events/registry"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

// SNSPublisher implements Publisher using AWS SNS
type SNSPublisher struct {
	topicARN  string
	snsClient *sns.Client
}

// NewSNSPublisher creates a new SNS publisher
func NewSNSPublisher(topicARN string, snsClient *sns.Client) *SNSPublisher {
	return &SNSPublisher{
		topicARN:  topicARN,
		snsClient: snsClient,
	}
}

// Publish publishes an event to SNS
func (p *SNSPublisher) Publish(ctx context.Context, args events.PublishArgs) error {
	return p.PublishBatch(ctx, []events.PublishArgs{args})
}

// PublishBatch publishes a batch of events to SNS
func (p *SNSPublisher) PublishBatch(ctx context.Context, args []events.PublishArgs) error {
	logger := observability.LoggerFromContext(ctx)

	batch, eventTypesList, err := buildSNSBatch(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to build SNS batch: %w", err)
	}

	logger.InfoContext(ctx, "publishing batch of events to SNS", "topic_arn", p.topicARN, "batch_size", len(batch), "event_types", eventTypesList)
	response, err := p.snsClient.PublishBatch(ctx, &sns.PublishBatchInput{
		PublishBatchRequestEntries: batch,
		TopicArn:                   aws.String(p.topicARN),
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to publish batch of events to SNS", "error", err)
		return err
	}

	failureCount := 0
	for _, result := range response.Failed {
		logger.ErrorContext(ctx, "failed to publish event to SNS", "error", *result.Message, "message_id", *result.Id)
		failureCount++
	}
	if failureCount > 0 {
		return fmt.Errorf("failed to publish %d events to SNS", failureCount)
	}

	return nil
}

// Make sure the publisher implements the Publisher interface
var _ events.Publisher = &SNSPublisher{}

func buildSNSBatch(ctx context.Context, args []events.PublishArgs) ([]snstypes.PublishBatchRequestEntry, []string, error) {
	batch := make([]snstypes.PublishBatchRequestEntry, len(args))

	parentCorrelationID := events.GetCorrelationID(ctx)

	eventTypes := map[string]bool{}

	for i, arg := range args {
		correlationID := parentCorrelationID
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		envelope, err := registry.NewEnvelope(arg.Payload, arg.Metadata.Source, correlationID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create envelope: %w", err)
		}

		data, err := envelope.Marshal()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to serialize event (aggregate_id=%s, event_id=%s, event_type=%s): %w",
				envelope.AggregateID(), envelope.ID, envelope.CorrelationID(), err)
		}
		eventTypes[envelope.Type] = true
		batch[i] = snstypes.PublishBatchRequestEntry{
			Id:                     aws.String(envelope.ID),
			Message:                aws.String(string(data)),
			MessageGroupId:         aws.String(envelope.AggregateID()),
			MessageDeduplicationId: aws.String(envelope.ID),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"event_type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(envelope.Type),
				},
			},
		}
	}

	eventTypesList := make([]string, 0, len(eventTypes))
	for eventType := range eventTypes {
		eventTypesList = append(eventTypesList, eventType)
	}

	return batch, eventTypesList, nil
}
