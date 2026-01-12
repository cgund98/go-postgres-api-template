package repo

import (
	"context"

	"github.com/cgund98/go-postgres-api-template/internal/domain/user/model"
)

// Repository defines the interface for user data access
// The repository extracts the database context from the context.Context internally
type Repository interface {
	// Create creates a new user
	Create(ctx context.Context, u *model.UserCreate) (*model.User, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*model.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*model.User, error)

	// Update updates an existing user
	Update(ctx context.Context, id string, u *model.UserUpdate) (*model.User, error)

	// Delete deletes a user by ID
	Delete(ctx context.Context, id string) error

	// List retrieves a list of users with pagination
	List(ctx context.Context, limit, offset int) ([]*model.User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int, error)
}
