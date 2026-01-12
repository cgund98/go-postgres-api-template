package postgres

import (
	"context"
	"database/sql"

	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/db"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

var logger = observability.Logger

// TransactionManager implements db.TransactionManager[*Context] for PostgreSQL.
// It stores transactions in context using the txKey defined in context.go.
// Repositories must use GetDBFromContext() from the same package to retrieve transactions.
type TransactionManager struct {
	db *sql.DB
}

// NewTransactionManager creates a new PostgreSQL transaction manager
func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes a function within a transaction
func (m *TransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Create a new context with the transaction
	// Using context.WithValue to store request-scoped transaction data
	txCtx := context.WithValue(ctx, txKey, tx)

	// Execute the function
	if err := fn(txCtx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Error("failed to rollback transaction", "error", rollbackErr)
			return rollbackErr
		}
		logger.Error("rolled back transaction")
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// Ensure TransactionManager implements db.TransactionManager
var _ db.TransactionManager = &TransactionManager{}
