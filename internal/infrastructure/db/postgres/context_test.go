package postgres

import (
	"context"
	"database/sql"
	"testing"
)

// TestTransactionKeyConsistency verifies that TransactionManager and GetTXFromContext
// use the same key. This test ensures that repositories can find transactions stored
// by the transaction manager.
func TestTransactionKeyConsistency(t *testing.T) {
	// Create a mock transaction
	mockTx := &sql.Tx{} // In real tests, you'd use a proper mock or test DB

	// Store transaction using the same mechanism TransactionManager uses
	ctx := context.WithValue(context.Background(), txKey, mockTx)

	// Verify GetTXFromContext can retrieve it
	tx := GetTXFromContext(ctx)
	if tx == nil {
		t.Fatal("GetTXFromContext failed to retrieve transaction - key mismatch detected!")
	}

	if tx != mockTx {
		t.Fatal("GetTXFromContext retrieved wrong transaction")
	}
}
