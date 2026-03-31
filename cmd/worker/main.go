package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	awsUtils "github.com/cgund98/go-postgres-api-template/internal/adapters/aws"
	"github.com/cgund98/go-postgres-api-template/internal/adapters/events/consumer"
	"github.com/cgund98/go-postgres-api-template/internal/config"
	"github.com/cgund98/go-postgres-api-template/internal/domain/events"
	userEventsV1 "github.com/cgund98/go-postgres-api-template/internal/domain/events/registry/users/v1"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

var logger = observability.Logger

func main() {
	ctx := context.Background()

	logger.Info("Starting worker...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load settings", "error", err)
		os.Exit(1)
	}

	// Initialize AWS config
	awsCfg, err := awsUtils.LoadAWSConfig(ctx, cfg)
	if err != nil {
		logger.Error("Failed to load AWS config", "error", err)
		os.Exit(1)
	}

	// Initialize SQS client
	sqsClient := sqs.NewFromConfig(awsCfg, func(o *sqs.Options) {
		if cfg.AwsEndpoint != "" {
			o.BaseEndpoint = aws.String(cfg.AwsEndpoint)
		}
	})

	// Register event handlers
	userCreatedHandler := user.NewCreateUserHandler()
	userUpdatedHandler := user.NewUpdateUserHandler()
	userDeletedHandler := user.NewDeleteUserHandler()

	router := events.NewRouter()
	router.RegisterHandler(userEventsV1.EventTypeUserCreated, userCreatedHandler)
	router.RegisterHandler(userEventsV1.EventTypeUserUpdated, userUpdatedHandler)
	router.RegisterHandler(userEventsV1.EventTypeUserDeleted, userDeletedHandler)

	// Create consumer
	userConsumer := consumer.NewSQSConsumer(sqsClient, consumer.SQSConsumerOptions{
		QueueURL:            cfg.EventsQueueURLUser,
		MaxNumberOfMessages: aws.Int32(1),
	})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start consuming messages
	userConsumer.Start(ctx, router)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...")
}
