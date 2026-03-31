package user

import (
	"context"
)

// Repository defines the interface for user data access
// The repository extracts the database context from the context.Context internally
type Repository interface {
	// Create creates a new user
	Create(ctx context.Context, u *CreateUserCommand) (User, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (User, error)

	// Update updates an existing user
	Update(ctx context.Context, id string, u *UpdateUserCommand) (User, error)

	// Delete deletes a user by ID
	Delete(ctx context.Context, id string) error

	// List retrieves a list of users with pagination
	List(ctx context.Context, limit, offset int) ([]User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int, error)
}
