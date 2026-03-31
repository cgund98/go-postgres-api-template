package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"

	domainEvents "github.com/cgund98/go-postgres-api-template/internal/domain/events"
	"github.com/cgund98/go-postgres-api-template/internal/domain/events/registry"
	v1 "github.com/cgund98/go-postgres-api-template/internal/domain/events/registry/users/v1"
)

// mockSQSClient is a mock implementation of SQS client
type mockSQSClient struct {
	deleteMessageFunc           func(context.Context, *sqs.DeleteMessageInput, ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	deleteMessageBatchFunc      func(context.Context, *sqs.DeleteMessageBatchInput, ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error)
	receiveMessageFunc          func(context.Context, *sqs.ReceiveMessageInput, ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	deleteMessageCallCount      int
	deleteMessageBatchCallCount int
	receiveMessageCallCount     int
}

func (m *mockSQSClient) DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	m.deleteMessageCallCount++
	if m.deleteMessageFunc != nil {
		return m.deleteMessageFunc(ctx, input, optFns...)
	}
	return &sqs.DeleteMessageOutput{}, nil
}

func (m *mockSQSClient) DeleteMessageBatch(ctx context.Context, input *sqs.DeleteMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error) {
	m.deleteMessageBatchCallCount++
	if m.deleteMessageBatchFunc != nil {
		return m.deleteMessageBatchFunc(ctx, input, optFns...)
	}
	return &sqs.DeleteMessageBatchOutput{}, nil
}

func (m *mockSQSClient) ReceiveMessage(ctx context.Context, input *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	m.receiveMessageCallCount++
	if m.receiveMessageFunc != nil {
		return m.receiveMessageFunc(ctx, input, optFns...)
	}
	return &sqs.ReceiveMessageOutput{}, nil
}

// mockEnvelopeHandler implements events.Handler for testing
type mockEnvelopeHandler struct {
	handleFunc    func(context.Context, registry.Envelope) error
	callCount     int
	lastEnvelopes []registry.Envelope
}

func (m *mockEnvelopeHandler) Handle(ctx context.Context, envelope registry.Envelope) error {
	m.callCount++
	m.lastEnvelopes = append(m.lastEnvelopes, envelope)
	if m.handleFunc != nil {
		return m.handleFunc(ctx, envelope)
	}
	return nil
}

var _ domainEvents.Handler = &mockEnvelopeHandler{}

func mustMarshalEnvelope(t *testing.T, payload registry.Payload, source string) string {
	t.Helper()
	envelope, err := registry.NewEnvelope(payload, source, "test-correlation-id")
	if err != nil {
		t.Fatalf("failed to create envelope: %v", err)
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("failed to marshal envelope: %v", err)
	}
	return string(data)
}

func TestSQSConsumer_Ack(t *testing.T) {
	tests := []struct {
		name          string
		messageID     string
		mockFunc      func(context.Context, *sqs.DeleteMessageInput, ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
		expectedError bool
	}{
		{
			name:      "successfully acks message",
			messageID: "test-receipt-handle",
			mockFunc: func(_ context.Context, input *sqs.DeleteMessageInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
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
			mockFunc: func(_ context.Context, _ *sqs.DeleteMessageInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
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

			consumer := &SQSConsumer{
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
		mockFunc      func(context.Context, *sqs.DeleteMessageBatchInput, ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error)
		expectedError bool
	}{
		{
			name:       "successfully acks batch of messages",
			messageIDs: []string{"handle-1", "handle-2", "handle-3"},
			mockFunc: func(_ context.Context, input *sqs.DeleteMessageBatchInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error) {
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
			mockFunc: func(_ context.Context, _ *sqs.DeleteMessageBatchInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error) {
				return nil, errors.New("SQS batch delete failed")
			},
			expectedError: true,
		},
		{
			name:       "handles empty batch",
			messageIDs: []string{},
			mockFunc: func(_ context.Context, input *sqs.DeleteMessageBatchInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error) {
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

			consumer := &SQSConsumer{
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
	validEnvelopeBody := func(t *testing.T) string {
		t.Helper()
		return mustMarshalEnvelope(t, &v1.UserCreatedEvent{
			UserID: "user-123",
			Email:  "test@example.com",
		}, "test-service")
	}

	tests := []struct {
		name                 string
		sqsMessages          []sqstypes.Message
		sqsError             error
		handlerError         error
		ackError             error
		expectedHandlerCalls int
		expectedAckCalls     int
	}{
		{
			name: "successfully processes single message",
			sqsMessages: []sqstypes.Message{
				{
					Body:          aws.String("placeholder"),
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
			sqsMessages:          []sqstypes.Message{},
			expectedHandlerCalls: 0,
			expectedAckCalls:     0,
		},
		{
			name: "handles invalid envelope JSON",
			sqsMessages: []sqstypes.Message{
				{
					Body:          aws.String("invalid json"),
					ReceiptHandle: aws.String("receipt-handle-1"),
				},
			},
			expectedHandlerCalls: 0,
			expectedAckCalls:     0,
		},
		{
			name: "handles handler error",
			sqsMessages: []sqstypes.Message{
				{
					Body:          aws.String("placeholder"),
					ReceiptHandle: aws.String("receipt-handle-1"),
				},
			},
			handlerError:         errors.New("handler failed"),
			expectedHandlerCalls: 1,
			expectedAckCalls:     0,
		},
		{
			name: "handles ack error",
			sqsMessages: []sqstypes.Message{
				{
					Body:          aws.String("placeholder"),
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
			// Fill in valid envelope bodies for messages that aren't testing invalid JSON
			for i := range tt.sqsMessages {
				if tt.sqsMessages[i].Body != nil && *tt.sqsMessages[i].Body == "placeholder" {
					body := validEnvelopeBody(t)
					tt.sqsMessages[i].Body = &body
				}
			}

			mockClient := &mockSQSClient{
				receiveMessageFunc: func(_ context.Context, _ *sqs.ReceiveMessageInput, _ ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
					if tt.sqsError != nil {
						return nil, tt.sqsError
					}
					return &sqs.ReceiveMessageOutput{
						Messages: tt.sqsMessages,
					}, nil
				},
				deleteMessageFunc: func(_ context.Context, _ *sqs.DeleteMessageInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
					if tt.ackError != nil {
						return nil, tt.ackError
					}
					return &sqs.DeleteMessageOutput{}, nil
				},
			}

			handler := &mockEnvelopeHandler{
				handleFunc: func(_ context.Context, _ registry.Envelope) error {
					return tt.handlerError
				},
			}

			router := domainEvents.NewRouter()
			router.RegisterHandler(v1.EventTypeUserCreated, handler)

			consumer := &SQSConsumer{
				queueURL:            "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				sqsClient:           mockClient,
				maxNumberOfMessages: 1,
				visibilityTimeout:   30,
				waitTimeSeconds:     20,
				logger:              slog.Default(),
			}

			consumer.processBatchOfSingleMessages(context.Background(), router)

			if handler.callCount != tt.expectedHandlerCalls {
				t.Errorf("expected handler to be called %d times, got %d", tt.expectedHandlerCalls, handler.callCount)
			}

			if mockClient.deleteMessageCallCount != tt.expectedAckCalls {
				t.Errorf("expected Ack to be called %d times, got %d", tt.expectedAckCalls, mockClient.deleteMessageCallCount)
			}
		})
	}
}
