package httpapi

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cgund98/go-postgres-api-template/internal/adapters/db/postgres"
	"github.com/cgund98/go-postgres-api-template/internal/adapters/events/publisher"
	useradapters "github.com/cgund98/go-postgres-api-template/internal/adapters/user"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user"
)

// Dependencies holds all dependencies for the presentation layer
type Dependencies struct {
	UserService *user.Service
	EventPub    publisher.Publisher
	// Add other dependencies as needed
}

// NewDependencies creates new dependencies
func NewDependencies(dbPool *pgxpool.Pool, eventPub publisher.Publisher) *Dependencies {
	// Create PostgreSQL transaction manager (takes sql.DB)
	txManager := postgres.NewTransactionManager(dbPool, pgx.TxOptions{})

	// Create repository (it extracts DB from context internally)
	userRepo := useradapters.NewPostgresRepository()

	// Create service
	userService := user.NewService(userRepo, txManager, eventPub)

	return &Dependencies{
		UserService: userService,
		EventPub:    eventPub,
	}
}
