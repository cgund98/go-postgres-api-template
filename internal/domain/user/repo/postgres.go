package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/cgund98/go-postgres-api-template/internal/domain"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/model"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/db"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/db/postgres"
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
func (r *PostgresRepository) Create(ctx context.Context, u *model.UserCreate) (*model.User, error) {
	tx := postgres.GetTXFromContext(ctx)
	if tx == nil {
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

	err := tx.QueryRowContext(ctx, query,
		newUser.ID,
		newUser.Email,
		newUser.FirstName,
		newUser.LastName,
		newUser.CreatedAt,
		newUser.UpdatedAt,
	).Scan(
		&newUser.ID,
		&newUser.Email,
		&newUser.FirstName,
		&newUser.LastName,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return newUser, nil
}

// GetByID retrieves a user by ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	tx := postgres.GetTXFromContext(ctx)
	if tx == nil {
		return nil, db.ErrNoDBContext
	}
	u := &model.User{}
	query := `
		SELECT id, email, first_name, last_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := tx.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	tx := postgres.GetTXFromContext(ctx)
	if tx == nil {
		return nil, db.ErrNoDBContext
	}
	u := &model.User{}
	query := `
		SELECT id, email, first_name, last_name, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := tx.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.FirstName,
		&u.LastName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

// Update updates an existing user
func (r *PostgresRepository) Update(ctx context.Context, id string, u *model.UserUpdate) (*model.User, error) {
	tx := postgres.GetTXFromContext(ctx)
	if tx == nil {
		return nil, db.ErrNoDBContext
	}
	// Build dynamic update query based on provided fields
	query := `UPDATE users SET updated_at = $1`
	args := []any{time.Now()}
	argIndex := 2

	if u.Email != nil {
		query += fmt.Sprintf(`, email = $%d`, argIndex)
		args = append(args, *u.Email)
		argIndex++
	}
	if u.FirstName != nil {
		query += fmt.Sprintf(`, first_name = $%d`, argIndex)
		args = append(args, *u.FirstName)
		argIndex++
	}
	if u.LastName != nil {
		query += fmt.Sprintf(`, last_name = $%d`, argIndex)
		args = append(args, *u.LastName)
		argIndex++
	}

	query += fmt.Sprintf(` WHERE id = $%d RETURNING id, email, first_name, last_name, created_at, updated_at`, argIndex)
	args = append(args, id)

	updatedUser := &model.User{}
	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&updatedUser.ID,
		&updatedUser.Email,
		&updatedUser.FirstName,
		&updatedUser.LastName,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

// Delete deletes a user by ID
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	tx := postgres.GetTXFromContext(ctx)
	if tx == nil {
		return db.ErrNoDBContext
	}
	query := `DELETE FROM users WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Verify deletion by checking if user still exists
	_, checkErr := r.GetByID(ctx, id)
	if checkErr == nil {
		// User still exists, deletion failed
		return domain.ErrNotFound
	}
	if checkErr != domain.ErrNotFound {
		return checkErr
	}
	// User was deleted successfully
	return nil
}

// List retrieves a list of users with pagination
func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	tx := postgres.GetTXFromContext(ctx)
	if tx == nil {
		return nil, db.ErrNoDBContext
	}
	query := `
		SELECT id, email, first_name, last_name, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := tx.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.FirstName,
			&u.LastName,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Count returns the total number of users
func (r *PostgresRepository) Count(ctx context.Context) (int, error) {
	tx := postgres.GetTXFromContext(ctx)
	if tx == nil {
		return 0, db.ErrNoDBContext
	}
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := tx.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Ensure PostgresRepository implements Repository
var _ Repository = (*PostgresRepository)(nil)
