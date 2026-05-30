package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cgund98/go-postgres-api-template/internal/adapters/db"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

// TransactionManager implements db.TransactionManager[*DbContext] for PostgreSQL.
// It stores transactions in context using the txKey defined in context.go.
// Repositories must use GetDBFromContext() from the same package to retrieve transactions.
type TransactionManager struct {
	db        *pgxpool.Pool
	txOptions pgx.TxOptions
}

// NewTransactionManager creates a new PostgreSQL transaction manager
func NewTransactionManager(db *pgxpool.Pool, txOptions pgx.TxOptions) *TransactionManager {
	return &TransactionManager{db: db, txOptions: txOptions}
}

// WithTransaction executes a function within a transaction
func (m *TransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	logger := observability.LoggerFromContext(ctx)

	tx, err := m.db.BeginTx(ctx, m.txOptions)
	if err != nil {
		return err
	}

	// Create a new context with the transaction
	dbContext := &DbContext{Tx: tx}
	txCtx := context.WithValue(ctx, txKey, dbContext)

	// Execute the function
	if err := fn(txCtx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			logger.ErrorContext(ctx, "failed to rollback transaction", "error", rollbackErr)
			return rollbackErr
		}
		logger.InfoContext(ctx, "rolled back transaction")
		return err
	}

	// Commit the transaction
	return tx.Commit(ctx)
}

// Ensure TransactionManager implements db.TransactionManager
var _ db.TransactionManager = &TransactionManager{}
