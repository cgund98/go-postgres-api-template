package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/cgund98/go-postgres-api-template/internal/domain/user/events"
)

// mockSQSClient is a mock implementation of SQS client
type mockSQSClient struct {
	deleteMessageFunc           func(*sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
	deleteMessageBatchFunc      func(*sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error)
	receiveMessageFunc          func(*sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	deleteMessageCallCount      int
	deleteMessageBatchCallCount int
	receiveMessageCallCount     int
}

func (m *mockSQSClient) DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	m.deleteMessageCallCount++
	if m.deleteMessageFunc != nil {
		return m.deleteMessageFunc(input)
	}
	return &sqs.DeleteMessageOutput{}, nil
}

func (m *mockSQSClient) DeleteMessageBatch(input *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error) {
	m.deleteMessageBatchCallCount++
	if m.deleteMessageBatchFunc != nil {
		return m.deleteMessageBatchFunc(input)
	}
	return &sqs.DeleteMessageBatchOutput{}, nil
}

func (m *mockSQSClient) ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	m.receiveMessageCallCount++
	if m.receiveMessageFunc != nil {
		return m.receiveMessageFunc(input)
	}
	return &sqs.ReceiveMessageOutput{}, nil
}

// mockHandler is a mock implementation of Handler
type mockHandler struct {
	handleFunc func(context.Context, *events.UserCreatedEvent) error
	callCount  int
	lastEvent  *events.UserCreatedEvent
}

func (m *mockHandler) Handle(ctx context.Context, event *events.UserCreatedEvent) error {
	m.callCount++
	m.lastEvent = event
	if m.handleFunc != nil {
		return m.handleFunc(ctx, event)
	}
	return nil
}

// mockBatchHandler is a mock implementation of BatchHandler
type mockBatchHandler struct {
	handleBatchFunc func(context.Context, []*events.UserCreatedEvent) error
	callCount       int
	lastEvents      []*events.UserCreatedEvent
}

func (m *mockBatchHandler) HandleBatch(ctx context.Context, eventList []*events.UserCreatedEvent) error {
	m.callCount++
	m.lastEvents = eventList
	if m.handleBatchFunc != nil {
		return m.handleBatchFunc(ctx, eventList)
	}
	return nil
}

// mockDeserializer is a mock implementation of Deserializer
type mockDeserializer struct {
	deserializeFunc func([]byte) (*events.UserCreatedEvent, error)
	callCount       int
}

func (m *mockDeserializer) Deserialize(data []byte) (*events.UserCreatedEvent, error) {
	m.callCount++
	if m.deserializeFunc != nil {
		return m.deserializeFunc(data)
	}
	// Default: deserialize the JSON
	var event events.UserCreatedEvent
	err := json.Unmarshal(data, &event)
	return &event, err
}

func TestSQSConsumer_Ack(t *testing.T) {
	tests := []struct {
		name          string
		messageID     string
		mockFunc      func(*sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
		expectedError bool
	}{
		{
			name:      "successfully acks message",
			messageID: "test-receipt-handle",
			mockFunc: func(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
				if *input.QueueUrl != "https://sqs.us-east-1.amazonaws.com/123456789/test-queue" {
					t.Errorf("unexpected queue URL: %s", *input.QueueUrl)
				}
				if *input.ReceiptHandle != "test-receipt-handle" {
					t.Errorf("unexpected receipt handle: %s", *input.ReceiptHandle)
				}
				return &sqs.DeleteMessageOutput{}, nil
			},
			expectedError: false,
		},
		{
			name:      "returns error when SQS delete fails",
			messageID: "test-receipt-handle",
			mockFunc: func(_ *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
				return nil, errors.New("SQS delete failed")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSQSClient{
				deleteMessageFunc: tt.mockFunc,
			}

			consumer := &SQSConsumer[*events.UserCreatedEvent]{
				queueURL:  "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				sqsClient: mockClient,
			}

			err := consumer.Ack(context.Background(), tt.messageID)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if mockClient.deleteMessageCallCount != 1 {
					t.Errorf("expected DeleteMessage to be called once, got %d", mockClient.deleteMessageCallCount)
				}
			}
		})
	}
}

func TestSQSConsumer_BatchAck(t *testing.T) {
	tests := []struct {
		name          string
		messageIDs    []string
		mockFunc      func(*sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error)
		expectedError bool
	}{
		{
			name:       "successfully acks batch of messages",
			messageIDs: []string{"handle-1", "handle-2", "handle-3"},
			mockFunc: func(input *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error) {
				if *input.QueueUrl != "https://sqs.us-east-1.amazonaws.com/123456789/test-queue" {
					t.Errorf("unexpected queue URL: %s", *input.QueueUrl)
				}
				if len(input.Entries) != 3 {
					t.Errorf("unexpected number of entries: %d", len(input.Entries))
				}
				for i, entry := range input.Entries {
					expectedID := input.Entries[i].Id
					if entry.Id == nil || *entry.Id != *expectedID {
						t.Errorf("entry %d ID mismatch", i)
					}
					if entry.ReceiptHandle == nil || *entry.ReceiptHandle != *expectedID {
						t.Errorf("entry %d receipt handle mismatch", i)
					}
				}
				return &sqs.DeleteMessageBatchOutput{}, nil
			},
			expectedError: false,
		},
		{
			name:       "returns error when SQS batch delete fails",
			messageIDs: []string{"handle-1", "handle-2"},
			mockFunc: func(_ *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error) {
				return nil, errors.New("SQS batch delete failed")
			},
			expectedError: true,
		},
		{
			name:       "handles empty batch",
			messageIDs: []string{},
			mockFunc: func(input *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error) {
				if len(input.Entries) != 0 {
					t.Errorf("expected empty entries, got %d", len(input.Entries))
				}
				return &sqs.DeleteMessageBatchOutput{}, nil
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSQSClient{
				deleteMessageBatchFunc: tt.mockFunc,
			}

			consumer := &SQSConsumer[*events.UserCreatedEvent]{
				queueURL:  "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				sqsClient: mockClient,
			}

			err := consumer.BatchAck(context.Background(), tt.messageIDs)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(tt.messageIDs) > 0 && mockClient.deleteMessageBatchCallCount != 1 {
					t.Errorf("expected DeleteMessageBatch to be called once, got %d", mockClient.deleteMessageBatchCallCount)
				}
			}
		})
	}
}

func TestSQSConsumer_processBatchOfSingleMessages(t *testing.T) {
	tests := []struct {
		name                 string
		sqsMessages          []*sqs.Message
		sqsError             error
		deserializeError     error
		handlerError         error
		ackError             error
		expectedHandlerCalls int
		expectedAckCalls     int
	}{
		{
			name: "successfully processes single message",
			sqsMessages: []*sqs.Message{
				{
					Body:          aws.String(`{"event_id":"test-id","event_type":"user.created","timestamp":"2023-01-01T00:00:00Z","user_id":"user-123","email":"test@example.com"}`),
					ReceiptHandle: aws.String("receipt-handle-1"),
				},
			},
			expectedHandlerCalls: 1,
			expectedAckCalls:     1,
		},
		{
			name:                 "handles SQS receive error",
			sqsError:             errors.New("SQS receive failed"),
			expectedHandlerCalls: 0,
			expectedAckCalls:     0,
		},
		{
			name:                 "handles empty message batch",
			sqsMessages:          []*sqs.Message{},
			expectedHandlerCalls: 0,
			expectedAckCalls:     0,
		},
		{
			name: "handles deserialization error",
			sqsMessages: []*sqs.Message{
				{
					Body:          aws.String("invalid json"),
					ReceiptHandle: aws.String("receipt-handle-1"),
				},
			},
			deserializeError:     errors.New("deserialization failed"),
			expectedHandlerCalls: 0,
			expectedAckCalls:     0,
		},
		{
			name: "handles handler error",
			sqsMessages: []*sqs.Message{
				{
					Body:          aws.String(`{"event_id":"test-id","event_type":"user.created","timestamp":"2023-01-01T00:00:00Z","user_id":"user-123","email":"test@example.com"}`),
					ReceiptHandle: aws.String("receipt-handle-1"),
				},
			},
			handlerError:         errors.New("handler failed"),
			expectedHandlerCalls: 1,
			expectedAckCalls:     0, // Should not ack if handler fails
		},
		{
			name: "handles ack error",
			sqsMessages: []*sqs.Message{
				{
					Body:          aws.String(`{"event_id":"test-id","event_type":"user.created","timestamp":"2023-01-01T00:00:00Z","user_id":"user-123","email":"test@example.com"}`),
					ReceiptHandle: aws.String("receipt-handle-1"),
				},
			},
			ackError:             errors.New("ack failed"),
			expectedHandlerCalls: 1,
			expectedAckCalls:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSQSClient{
				receiveMessageFunc: func(_ *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
					if tt.sqsError != nil {
						return nil, tt.sqsError
					}
					return &sqs.ReceiveMessageOutput{
						Messages: tt.sqsMessages,
					}, nil
				},
				deleteMessageFunc: func(_ *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
					if tt.ackError != nil {
						return nil, tt.ackError
					}
					return &sqs.DeleteMessageOutput{}, nil
				},
			}

			mockHandler := &mockHandler{
				handleFunc: func(_ context.Context, _ *events.UserCreatedEvent) error {
					return tt.handlerError
				},
			}

			mockDeserializer := &mockDeserializer{
				deserializeFunc: func(data []byte) (*events.UserCreatedEvent, error) {
					if tt.deserializeError != nil {
						return nil, tt.deserializeError
					}
					var event events.UserCreatedEvent
					err := json.Unmarshal(data, &event)
					return &event, err
				},
			}

			consumer := &SQSConsumer[*events.UserCreatedEvent]{
				queueURL:            "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				sqsClient:           mockClient,
				maxNumberOfMessages: 1,
				visibilityTimeout:   30,
				waitTimeSeconds:     20,
				logger:              slog.Default(),
			}

			consumer.processBatchOfSingleMessages(context.Background(), mockDeserializer, mockHandler)

			if mockHandler.callCount != tt.expectedHandlerCalls {
				t.Errorf("expected handler to be called %d times, got %d", tt.expectedHandlerCalls, mockHandler.callCount)
			}

			if mockClient.deleteMessageCallCount != tt.expectedAckCalls {
				t.Errorf("expected Ack to be called %d times, got %d", tt.expectedAckCalls, mockClient.deleteMessageCallCount)
			}
		})
	}
}

func TestSQSConsumer_processBatchOfMessages(t *testing.T) {
	tests := []struct {
		name                 string
		sqsMessages          []*sqs.Message
		sqsError             error
		deserializeError     error
		handlerError         error
		ackError             error
		expectedHandlerCalls int
		expectedAckCalls     int
	}{
		{
			name: "successfully processes batch of messages",
			sqsMessages: []*sqs.Message{
				{
					Body:          aws.String(`{"event_id":"test-id-1","event_type":"user.created","timestamp":"2023-01-01T00:00:00Z","user_id":"user-123","email":"test1@example.com"}`),
					ReceiptHandle: aws.String("receipt-handle-1"),
				},
				{
					Body:          aws.String(`{"event_id":"test-id-2","event_type":"user.created","timestamp":"2023-01-01T00:00:00Z","user_id":"user-456","email":"test2@example.com"}`),
					ReceiptHandle: aws.String("receipt-handle-2"),
				},
			},
			expectedHandlerCalls: 1,
			expectedAckCalls:     2,
		},
		{
			name:                 "handles SQS receive error",
			sqsError:             errors.New("SQS receive failed"),
			expectedHandlerCalls: 0,
			expectedAckCalls:     0,
		},
		{
			name:                 "handles empty message batch",
			sqsMessages:          []*sqs.Message{},
			expectedHandlerCalls: 0,
			expectedAckCalls:     0,
		},
		{
			name: "handles deserialization error",
			sqsMessages: []*sqs.Message{
				{
					Body:          aws.String("invalid json"),
					ReceiptHandle: aws.String("receipt-handle-1"),
				},
			},
			deserializeError:     errors.New("deserialization failed"),
			expectedHandlerCalls: 0,
			expectedAckCalls:     0,
		},
		{
			name: "handles handler error",
			sqsMessages: []*sqs.Message{
				{
					Body:          aws.String(`{"event_id":"test-id","event_type":"user.created","timestamp":"2023-01-01T00:00:00Z","user_id":"user-123","email":"test@example.com"}`),
					ReceiptHandle: aws.String("receipt-handle-1"),
				},
			},
			handlerError:         errors.New("handler failed"),
			expectedHandlerCalls: 1,
			expectedAckCalls:     0, // Should not ack if handler fails
		},
		{
			name: "handles ack error",
			sqsMessages: []*sqs.Message{
				{
					Body:          aws.String(`{"event_id":"test-id","event_type":"user.created","timestamp":"2023-01-01T00:00:00Z","user_id":"user-123","email":"test@example.com"}`),
					ReceiptHandle: aws.String("receipt-handle-1"),
				},
			},
			ackError:             errors.New("ack failed"),
			expectedHandlerCalls: 1,
			expectedAckCalls:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSQSClient{
				receiveMessageFunc: func(_ *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
					if tt.sqsError != nil {
						return nil, tt.sqsError
					}
					return &sqs.ReceiveMessageOutput{
						Messages: tt.sqsMessages,
					}, nil
				},
				deleteMessageFunc: func(_ *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
					if tt.ackError != nil {
						return nil, tt.ackError
					}
					return &sqs.DeleteMessageOutput{}, nil
				},
			}

			mockBatchHandler := &mockBatchHandler{
				handleBatchFunc: func(_ context.Context, _ []*events.UserCreatedEvent) error {
					return tt.handlerError
				},
			}

			mockDeserializer := &mockDeserializer{
				deserializeFunc: func(data []byte) (*events.UserCreatedEvent, error) {
					if tt.deserializeError != nil {
						return nil, tt.deserializeError
					}
					var event events.UserCreatedEvent
					err := json.Unmarshal(data, &event)
					return &event, err
				},
			}

			consumer := &SQSConsumer[*events.UserCreatedEvent]{
				queueURL:            "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				sqsClient:           mockClient,
				maxNumberOfMessages: 10,
				visibilityTimeout:   30,
				waitTimeSeconds:     20,
				logger:              slog.Default(),
			}

			consumer.processBatchOfMessages(context.Background(), mockDeserializer, mockBatchHandler)

			if mockBatchHandler.callCount != tt.expectedHandlerCalls {
				t.Errorf("expected batch handler to be called %d times, got %d", tt.expectedHandlerCalls, mockBatchHandler.callCount)
			}

			if mockClient.deleteMessageCallCount != tt.expectedAckCalls {
				t.Errorf("expected Ack to be called %d times, got %d", tt.expectedAckCalls, mockClient.deleteMessageCallCount)
			}
		})
	}
}
