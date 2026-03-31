package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/cgund98/go-postgres-api-template/internal/adapters/db"
	"github.com/cgund98/go-postgres-api-template/internal/adapters/db/postgres"
	"github.com/cgund98/go-postgres-api-template/internal/domain"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/model"
)

// PostgresRepository implements the Repository interface for PostgreSQL.
// It extracts the database context from context.Context internally using
// postgres.GetDBFromContext(), which must match the key used by postgres.TransactionManager.
type PostgresRepository struct {
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository() *PostgresRepository {
	return &PostgresRepository{}
}

// Create creates a new user
func (r *PostgresRepository) Create(ctx context.Context, u *model.CreateUserCommand) (*model.User, error) {
	dbContext := postgres.GetTXFromContext(ctx)
	if dbContext == nil {
		return nil, db.ErrNoDBContext
	}

	now := time.Now()
	newUser := &model.User{
		ID:        uuid.New().String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `
		INSERT INTO users (id, email, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, first_name, last_name, created_at, updated_at
	`

	rows, err := dbContext.Tx.Query(ctx, query,
		newUser.ID,
		newUser.Email,
		newUser.FirstName,
		newUser.LastName,
		newUser.CreatedAt,
		newUser.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	createdUser, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect one row: %w", err)
	}

	return createdUser, nil
}

// GetByID retrieves a user by ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	dbContext := postgres.GetTXFromContext(ctx)
	if dbContext == nil {
		return nil, db.ErrNoDBContext
	}
	query := `
		SELECT id, email, first_name, last_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	rows, err := dbContext.Tx.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model.User])
	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to collect one row: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	dbContext := postgres.GetTXFromContext(ctx)
	if dbContext == nil {
		return nil, db.ErrNoDBContext
	}
	query := `
		SELECT id, email, first_name, last_name, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	rows, err := dbContext.Tx.Query(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model.User])
	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to collect one row: %w", err)
	}
	return user, nil
}

// Update updates an existing user
func (r *PostgresRepository) Update(ctx context.Context, id string, cmd *model.UpdateUserCommand) (*model.User, error) {
	dbContext := postgres.GetTXFromContext(ctx)
	if dbContext == nil {
		return nil, db.ErrNoDBContext
	}
	// Build dynamic update query based on provided fields
	query := `UPDATE users SET updated_at = $1`
	args := []any{time.Now()}
	argIndex := 2

	if cmd.Email != nil {
		query += fmt.Sprintf(`, email = $%d`, argIndex)
		args = append(args, *cmd.Email)
		argIndex++
	}
	if cmd.FirstName != nil {
		query += fmt.Sprintf(`, first_name = $%d`, argIndex)
		args = append(args, *cmd.FirstName)
		argIndex++
	}
	if cmd.LastName != nil {
		query += fmt.Sprintf(`, last_name = $%d`, argIndex)
		args = append(args, *cmd.LastName)
		argIndex++
	}

	query += fmt.Sprintf(` WHERE id = $%d RETURNING id, email, first_name, last_name, created_at, updated_at`, argIndex)
	args = append(args, id)

	rows, err := dbContext.Tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	updatedUser, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect one row: %w", err)
	}

	return updatedUser, nil
}

// Delete deletes a user by ID
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	dbContext := postgres.GetTXFromContext(ctx)
	if dbContext == nil {
		return db.ErrNoDBContext
	}

	query := `DELETE FROM users WHERE id = $1`
	_, err := dbContext.Tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// List retrieves a list of users with pagination
func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	dbContext := postgres.GetTXFromContext(ctx)
	if dbContext == nil {
		return nil, db.ErrNoDBContext
	}
	query := `
		SELECT id, email, first_name, last_name, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := dbContext.Tx.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect rows: %w", err)
	}

	return users, nil
}

// Count returns the total number of users
func (r *PostgresRepository) Count(ctx context.Context) (int, error) {
	dbContext := postgres.GetTXFromContext(ctx)
	if dbContext == nil {
		return 0, db.ErrNoDBContext
	}

	query := `SELECT COUNT(*) FROM users`
	var count int
	err := dbContext.Tx.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Ensure PostgresRepository implements Repository
var _ Repository = (*PostgresRepository)(nil)
