package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/cgund98/go-postgres-api-template/internal/config"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/aws"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/db/postgres"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events/publisher"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events/serializer"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
	"github.com/cgund98/go-postgres-api-template/internal/presentation"
	presentationuser "github.com/cgund98/go-postgres-api-template/internal/presentation/user"
)

var logger = observability.Logger

func main() {
	// Load configuration
	cfg, err := config.LoadSettings()
	if err != nil {
		logger.Error("Failed to load settings", "error", err)
		os.Exit(1)
	}

	// Initialize database
	dbPool, err := postgres.NewPool(cfg.Database.URL)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Initialize AWS clients
	awsSession, err := aws.NewSession(cfg.AWS)
	if err != nil {
		logger.Error("Failed to initialize AWS session", "error", err)
		os.Exit(1)
	}
	snsClient := sns.New(awsSession)

	// Initialize event publisher
	serializer := serializer.NewJSONSerializer()
	eventPub := publisher.NewSNSPublisher(cfg.Events.TopicARN, serializer, snsClient)
	logger.Info("event publisher initialized", "topic_arn", cfg.Events.TopicARN)

	// Initialize dependencies
	deps := presentation.NewDependencies(dbPool, eventPub)

	// Setup router with Chi and Huma
	router := presentation.NewRouter()

	// Register API v1 routes
	userController := presentationuser.NewUserController(deps.UserService)
	userController.RegisterRoutes(router.HumaAPI())

	// Health check endpoint (using Chi router directly)
	router.ChiRouter().Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			logger.Error("Failed to write health check response", "error", err)
		}
	})

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting on port", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited")
}
