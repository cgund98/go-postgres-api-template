package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Pool manages PostgreSQL database connections
type Pool struct {
	db *sql.DB
}

// NewPool creates a new PostgreSQL connection pool
func NewPool(connectionString string) (*Pool, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{db: db}, nil
}

// DB returns the underlying sql.DB instance
func (p *Pool) DB() *sql.DB {
	return p.db
}

// Close closes the database connection pool
func (p *Pool) Close() error {
	return p.db.Close()
}
