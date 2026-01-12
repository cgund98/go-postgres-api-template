package postgres

import (
	"context"
	"database/sql"
)

// txKey is an unexported type used as a key for storing transactions in context.
// Using a pointer to this type is more efficient and idiomatic.
//
// IMPORTANT: This key MUST be shared between TransactionManager.WithTransaction()
// and GetDBFromContext() to ensure repositories can find transactions. Since both
// are in the same package, they automatically share this variable.
//
// If you create a new database implementation (e.g., MySQL), it should use its own
// unique key in its own package to avoid conflicts.
//
// See TestTransactionKeyConsistency for a test that verifies this coupling.
var txKey = &struct{ name string }{"postgres_tx"}

// Context implements the db.DB interface for PostgreSQL
// It wraps a sql.Tx transaction
type Context struct {
	Tx *sql.Tx // The transaction
}

// GetTXFromContext extracts the DB context from a transaction context.
// Returns a Context wrapping the transaction if in a transaction, otherwise returns nil.
// This is a package-level function that repositories can use.
func GetTXFromContext(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		// We're in a transaction - return a wrapper that adapts Tx to Context
		return tx
	}
	return nil
}
