package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.opentelemetry.io/otel"

	awsUtils "github.com/cgund98/go-postgres-api-template/internal/adapters/aws"
	"github.com/cgund98/go-postgres-api-template/internal/adapters/events/consumer"
	"github.com/cgund98/go-postgres-api-template/internal/config"
	"github.com/cgund98/go-postgres-api-template/internal/domain/events"
	userEventsV1 "github.com/cgund98/go-postgres-api-template/internal/domain/events/registry/users/v1"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

const serviceName = "postgres-template/worker"

func main() {
	ctx := context.Background()

	logger := observability.NewBootstrapLogger()
	ctx = observability.SetLoggerOnContext(ctx, logger)

	logger.InfoContext(ctx, "Starting worker...")

	// Load configuration
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to load settings", "error", err)
		os.Exit(1)
	}

	// Set up OpenTelemetry.
	_, otelShutdown, err := observability.SetupOTelSDK(ctx, serviceName)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to setup OpenTelemetry", "error", err)
		os.Exit(1)
	}
	logger = observability.NewLogger(serviceName)
	ctx = observability.SetLoggerOnContext(ctx, logger)
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	tracer := otel.Tracer(serviceName)
	ctx = observability.SetTracerOnContext(ctx, tracer)

	// Initialize AWS config
	awsCfg, err := awsUtils.LoadAWSConfig(ctx, cfg)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to load AWS config", "error", err)
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
	userConsumer := consumer.NewSQSConsumer(logger, sqsClient, consumer.SQSConsumerOptions{
		QueueURL:            cfg.EventsQueueURLUser,
		MaxNumberOfMessages: aws.Int32(1),
	})

	// Start consuming messages
	userConsumer.Start(ctx, router)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.InfoContext(ctx, "Shutting down worker...")
}
