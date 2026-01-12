package publisher

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events/serializer"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

var logger = observability.Logger

// SNSPublisher implements Publisher using AWS SNS
type SNSPublisher struct {
	topicARN   string
	serializer serializer.Serializer
	snsClient  *sns.SNS
}

// NewSNSPublisher creates a new SNS publisher
func NewSNSPublisher(topicARN string, serializer serializer.Serializer, snsClient *sns.SNS) *SNSPublisher {
	return &SNSPublisher{
		topicARN:   topicARN,
		serializer: serializer,
		snsClient:  snsClient,
	}
}

// Publish publishes an event to SNS
// The event is serialized as JSON directly since all events are JSON structs
func (p *SNSPublisher) Publish(ctx context.Context, event events.Event) error {
	// Call publish batch with a single event
	return p.PublishBatch(ctx, []events.Event{event})
}

// PublishBatch publishes a batch of events to SNS
func (p *SNSPublisher) PublishBatch(_ context.Context, events []events.Event) error {

	batch := make([]*sns.PublishBatchRequestEntry, len(events))

	// Track unique event types in the batch
	eventTypes := map[string]bool{}

	for i, event := range events {
		data, err := p.serializer.Serialize(event)
		if err != nil {
			return fmt.Errorf("failed to serialize event (aggregate_id=%s, event_id=%s, event_type=%s): %w",
				event.AggregateID(), event.EventID(), event.Type(), err)
		}
		eventTypes[event.Type()] = true
		batch[i] = &sns.PublishBatchRequestEntry{
			Id:                     aws.String(event.EventID()),
			Message:                aws.String(string(data)),
			MessageGroupId:         aws.String(event.AggregateID()),
			MessageDeduplicationId: aws.String(event.EventID()),
			MessageAttributes: map[string]*sns.MessageAttributeValue{
				"event_type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(event.Type()),
				},
			},
		}
	}

	// Convert map to list of event types
	eventTypesList := make([]string, 0, len(eventTypes))
	for eventType := range eventTypes {
		eventTypesList = append(eventTypesList, eventType)
	}

	logger.Info("publishing batch of events to SNS", "topic_arn", p.topicARN, "batch_size", len(batch), "event_types", eventTypesList)
	response, err := p.snsClient.PublishBatch(&sns.PublishBatchInput{
		PublishBatchRequestEntries: batch,
		TopicArn:                   aws.String(p.topicARN),
	})
	if err != nil {
		logger.Error("failed to publish batch of events to SNS", "error", err)
		return err
	}
	failureCount := 0
	for _, result := range response.Failed {
		logger.Error("failed to publish event to SNS", "error", *result.Message, "message_id", *result.Id)
		failureCount++
	}
	if failureCount > 0 {
		return fmt.Errorf("failed to publish %d events to SNS", failureCount)
	}
	return nil
}

// Make sure the publisher implements the Publisher interface
var _ Publisher = &SNSPublisher{}
