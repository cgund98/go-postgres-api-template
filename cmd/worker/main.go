package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/cgund98/go-postgres-api-template/internal/config"
	userEvents "github.com/cgund98/go-postgres-api-template/internal/domain/user/events"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/events/handlers"
	awsUtils "github.com/cgund98/go-postgres-api-template/internal/infrastructure/aws"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events/consumer"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events/deserializer"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

var logger = observability.Logger

func main() {

	logger.Info("Starting worker...")

	// Load configuration
	cfg, err := config.LoadSettings()
	if err != nil {
		logger.Error("Failed to load settings", "error", err)
		os.Exit(1)
	}

	// Initialize AWS clients
	awsSession, err := awsUtils.NewSession(cfg.AWS)
	if err != nil {
		logger.Error("Failed to initialize AWS session", "error", err)
		os.Exit(1)
	}
	sqsClient := sqs.New(awsSession)

	// Register event handlers
	userCreatedHandler := handlers.NewUserCreatedHandler()
	userUpdatedHandler := handlers.NewUserUpdatedHandler()
	userDeletedHandler := handlers.NewUserDeletedHandler()

	// Create consumers
	userCreatedConsumer := consumer.NewSQSConsumer[*userEvents.UserCreatedEvent](sqsClient, consumer.SQSConsumerOptions{
		QueueURL:            cfg.Events.QueueURLUserCreated,
		MaxNumberOfMessages: aws.Int64(1),
	})
	userUpdatedConsumer := consumer.NewSQSConsumer[*userEvents.UserUpdatedEvent](sqsClient, consumer.SQSConsumerOptions{
		QueueURL:            cfg.Events.QueueURLUserUpdated,
		MaxNumberOfMessages: aws.Int64(1),
	})
	userDeletedConsumer := consumer.NewSQSConsumer[*userEvents.UserDeletedEvent](sqsClient, consumer.SQSConsumerOptions{
		QueueURL:            cfg.Events.QueueURLUserDeleted,
		MaxNumberOfMessages: aws.Int64(1),
	})

	// Create deserializers
	userCreatedDeserializer := deserializer.NewJSONDeserializer[*userEvents.UserCreatedEvent]()
	userUpdatedDeserializer := deserializer.NewJSONDeserializer[*userEvents.UserUpdatedEvent]()
	userDeletedDeserializer := deserializer.NewJSONDeserializer[*userEvents.UserDeletedEvent]()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consuming messages
	userCreatedConsumer.Start(ctx, userCreatedDeserializer, userCreatedHandler)
	userUpdatedConsumer.Start(ctx, userUpdatedDeserializer, userUpdatedHandler)
	userDeletedConsumer.Start(ctx, userDeletedDeserializer, userDeletedHandler)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...")
}
