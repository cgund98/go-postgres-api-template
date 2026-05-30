package user

import (
	"context"
	"fmt"

	"github.com/cgund98/go-postgres-api-template/internal/domain/events"
	v1 "github.com/cgund98/go-postgres-api-template/internal/domain/events/registry/users/v1"
)

// TransactionManager defines the interface for managing transactions
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Service handles user business logic
type Service struct {
	repo           Repository
	txManager      TransactionManager
	eventPublisher events.Publisher
	// Add other service dependencies here (e.g., invoice service)
}

// NewService creates a new user service
func NewService(
	repo Repository,
	txManager TransactionManager,
	eventPublisher events.Publisher,
) *Service {
	return &Service{
		repo:           repo,
		txManager:      txManager,
		eventPublisher: eventPublisher,
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, email, firstName, lastName string) (User, error) {
	var createdUser User

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.validateCreateUserRequest(txCtx, email, firstName, lastName); err != nil {
			return fmt.Errorf("failed to validate create user request: %w", err)
		}

		user, err := s.repo.Create(txCtx, &CreateUserCommand{
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		createdUser = user

		return nil
	})

	if err != nil {
		return User{}, err
	}

	publishArgs := events.PublishArgs{
		Payload:  &v1.UserCreatedEvent{UserID: createdUser.ID, Email: createdUser.Email},
		Metadata: events.PublishMetadata{Source: "user-service"},
	}
	if err := s.eventPublisher.Publish(ctx, publishArgs); err != nil {
		return User{}, fmt.Errorf("failed to publish user created event: %w", err)
	}

	return createdUser, nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, userID string) (User, error) {
	var user User

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		u, err := s.repo.GetByID(txCtx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user by ID: %w", err)
		}
		user = u
		return nil
	})

	return user, err
}

// PatchUser performs a partial update of a user
func (s *Service) PatchUser(ctx context.Context, userID string, update *UpdateUserCommand) (User, error) {
	var updatedUser User
	var changes Changes

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		existing, err := s.validatePatchUserRequest(txCtx, userID, update)
		if err != nil {
			return fmt.Errorf("failed to validate patch user request: %w", err)
		}

		changes = GenerateUserChanges(update, existing)

		if len(changes) == 0 {
			updatedUser = existing
			return nil
		}

		updated, err := s.repo.Update(txCtx, userID, update)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		updatedUser = updated

		return nil
	})

	if err != nil {
		return User{}, err
	}

	publishArgs := events.PublishArgs{
		Payload:  &v1.UserUpdatedEvent{UserID: updatedUser.ID, Changes: changes},
		Metadata: events.PublishMetadata{Source: "user-service"},
	}
	if err := s.eventPublisher.Publish(ctx, publishArgs); err != nil {
		return User{}, fmt.Errorf("failed to publish user updated event: %w", err)
	}

	return updatedUser, nil
}

// ListUsers retrieves a list of users with pagination
func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]User, int, error) {
	var users []User
	var total int

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		u, err := s.repo.List(txCtx, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}
		users = u

		count, err := s.repo.Count(txCtx)
		if err != nil {
			return fmt.Errorf("failed to count users: %w", err)
		}
		total = count

		return nil
	})

	return users, total, err
}

// DeleteUser deletes a user by ID
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	var deletedUser User
	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.validateDeleteUserRequest(txCtx, userID)
		if err != nil {
			return fmt.Errorf("failed to validate delete user request: %w", err)
		}

		deletedUser = user

		return s.repo.Delete(txCtx, userID)
	})
	if err != nil {
		return err
	}

	publishArgs := events.PublishArgs{
		Payload:  &v1.UserDeletedEvent{UserID: deletedUser.ID},
		Metadata: events.PublishMetadata{Source: "user-service"},
	}
	if err := s.eventPublisher.Publish(ctx, publishArgs); err != nil {
		return fmt.Errorf("failed to publish user deleted event: %w", err)
	}

	return nil
}
