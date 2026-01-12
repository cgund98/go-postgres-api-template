package presentation

import (
	"github.com/cgund98/go-postgres-api-template/internal/domain/user"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/repo"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/db/postgres"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events/publisher"
)

// Dependencies holds all dependencies for the presentation layer
type Dependencies struct {
	UserService *user.Service
	EventPub    publisher.Publisher
	// Add other dependencies as needed
}

// NewDependencies creates new dependencies
func NewDependencies(dbPool *postgres.Pool, eventPub publisher.Publisher) *Dependencies {
	// Create PostgreSQL transaction manager (takes sql.DB)
	txManager := postgres.NewTransactionManager(dbPool.DB())

	// Create repository (it extracts DB from context internally)
	userRepo := repo.NewPostgresRepository()

	// Create service
	userService := user.NewService(userRepo, txManager, eventPub)

	return &Dependencies{
		UserService: userService,
		EventPub:    eventPub,
	}
}
