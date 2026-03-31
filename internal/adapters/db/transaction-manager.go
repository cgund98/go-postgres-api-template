package db

import "context"

// TransactionManager manages database transactions
// It's generic over the database context type
type TransactionManager interface {
	// WithTransaction executes a function within a transaction
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
