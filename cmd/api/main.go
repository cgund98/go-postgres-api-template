package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"

	awsUtils "github.com/cgund98/go-postgres-api-template/internal/adapters/aws"
	"github.com/cgund98/go-postgres-api-template/internal/adapters/events/publisher"
	"github.com/cgund98/go-postgres-api-template/internal/config"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
	httprouter "github.com/cgund98/go-postgres-api-template/internal/presentation/httpapi"
	httpuser "github.com/cgund98/go-postgres-api-template/internal/presentation/httpapi/user"
)

const (
	serviceName = "postgres-template/http-api"
)

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger := observability.NewBootstrapLogger()
	ctx = observability.SetLoggerOnContext(ctx, logger)

	// Load configuration
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to load settings", "error", err)
		os.Exit(1)
	}

	// Set up OpenTelemetry.
	otelProviders, otelShutdown, err := observability.SetupOTelSDK(ctx, serviceName)
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

	// Initialize database
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Initialize AWS config
	awsCfg, err := awsUtils.LoadAWSConfig(ctx, cfg)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to load AWS config", "error", err)
		os.Exit(1)
	}

	// Initialize SNS client
	snsClient := sns.NewFromConfig(awsCfg, func(o *sns.Options) {
		if cfg.AwsEndpoint != "" {
			o.BaseEndpoint = aws.String(cfg.AwsEndpoint)
		}
	})

	// Initialize event publisher
	eventPub := publisher.NewSNSPublisher(cfg.EventsTopicArn, snsClient)
	logger.InfoContext(ctx, "event publisher initialized", "topic_arn", cfg.EventsTopicArn)

	// Initialize dependencies
	deps := httprouter.NewDependencies(dbPool, eventPub)

	// Setup router with Chi and Huma
	router := httprouter.NewRouter(serviceName, logger, tracer, otelProviders.MeterProvider)

	// Register API v1 routes
	userController := httpuser.NewUserController(deps.UserService)
	userController.RegisterRoutes(router.HumaAPI())

	// Health check endpoint (using Chi router directly)
	router.ChiRouter().Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.ErrorContext(ctx, "Failed to write health check response", "error", err)
		}
	})

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Start server in a goroutine
	srvErr := make(chan error, 1)
	go func() {
		logger.InfoContext(ctx, "Server starting on port", "port", cfg.ServerPort)
		srvErr <- server.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		logger.ErrorContext(ctx, "Server failed to start", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = server.Shutdown(shutdownCtx)
	if err != nil {
		logger.ErrorContext(ctx, "Server failed to shutdown", "error", err)
		os.Exit(1)
	}

	logger.InfoContext(ctx, "Server shut down successfully")
}
