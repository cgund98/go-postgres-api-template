package user

import (
	"context"

	"github.com/cgund98/go-postgres-api-template/internal/domain"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/events"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/model"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/repo"
	"github.com/cgund98/go-postgres-api-template/internal/infrastructure/events/publisher"
)

// TransactionManager defines the interface for managing transactions
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Service handles user business logic
type Service struct {
	repo           repo.Repository
	txManager      TransactionManager
	eventPublisher publisher.Publisher
	// Add other service dependencies here (e.g., invoice service)
}

// NewService creates a new user service
func NewService(
	repo repo.Repository,
	txManager TransactionManager,
	eventPublisher publisher.Publisher,
) *Service {
	return &Service{
		repo:           repo,
		txManager:      txManager,
		eventPublisher: eventPublisher,
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, email, firstName, lastName string) (*model.User, error) {
	var createdUser *model.User

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate request
		if err := s.validateCreateUserRequest(txCtx, email, firstName, lastName); err != nil {
			return err
		}

		// Create user (repository handles ID generation and timestamps)
		createUser := &model.UserCreate{
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
		}

		user, err := s.repo.Create(txCtx, createUser)
		if err != nil {
			return err
		}
		createdUser = user

		// Publish event after successful transaction
		if createdUser != nil {
			event := events.NewUserCreatedEvent(createdUser.ID, createdUser.Email)
			err := s.eventPublisher.Publish(ctx, event)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, userID string) (*model.User, error) {
	var user *model.User

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		u, err := s.repo.GetByID(txCtx, userID)
		if err != nil {
			return err
		}
		user = u
		return nil
	})

	return user, err
}

// PatchUser performs a partial update of a user
func (s *Service) PatchUser(ctx context.Context, userID string, update *model.UserUpdate) (*model.User, error) {
	var updatedUser *model.User
	var changes Changes

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate request and get existing user
		existing, err := s.validatePatchUserRequest(txCtx, userID, update)
		if err != nil {
			return err
		}

		// Generate changes dictionary for event
		changes = GenerateUserChanges(update, existing)

		// Check if there are any fields to update
		if len(changes) == 0 {
			updatedUser = existing
			return nil // No changes, return existing user
		}

		// Perform partial update
		updated, err := s.repo.Update(txCtx, userID, update)
		if err != nil {
			return err
		}
		updatedUser = updated

		// Publish event if there were changes
		if len(changes) > 0 && updatedUser != nil {
			event := events.NewUserUpdatedEvent(updatedUser.ID, changes)
			err := s.eventPublisher.Publish(ctx, event)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

// ListUsers retrieves a list of users with pagination
func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	var users []*model.User
	var total int

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		u, err := s.repo.List(txCtx, limit, offset)
		if err != nil {
			return err
		}
		users = u

		count, err := s.repo.Count(txCtx)
		if err != nil {
			return err
		}
		total = count

		return nil
	})

	return users, total, err
}

// DeleteUser deletes a user by ID
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	var deletedUserID string

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate request and get user
		user, err := s.validateDeleteUserRequest(txCtx, userID)
		if err != nil {
			return err
		}
		deletedUserID = user.ID

		// TODO: Delete associated resources (e.g., invoices) if needed
		// This would call other services' internal methods to avoid nested transactions

		// Publish event
		event := events.NewUserDeletedEvent(deletedUserID)
		err = s.eventPublisher.Publish(ctx, event)
		if err != nil {
			return err
		}

		// Delete the user
		return s.repo.Delete(txCtx, userID)
	})

	if err != nil {
		return err
	}

	return nil
}

// validateCreateUserRequest validates create user request
func (s *Service) validateCreateUserRequest(ctx context.Context, email, firstName, lastName string) error {
	// Validate input
	if err := ValidateCreate(&model.UserCreate{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}); err != nil {
		return err
	}

	// Check email uniqueness
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil && err != domain.ErrNotFound {
		return err
	}
	if existing != nil {
		return domain.ErrAlreadyExists
	}

	return nil
}

// validatePatchUserRequest validates patch user request and returns existing user
func (s *Service) validatePatchUserRequest(ctx context.Context, userID string, update *model.UserUpdate) (*model.User, error) {
	// Validate update input
	if err := ValidateUpdate(update); err != nil {
		return nil, err
	}

	// Get existing user
	existing, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// If email is being updated, check uniqueness
	if update.Email != nil && *update.Email != existing.Email {
		emailUser, err := s.repo.GetByEmail(ctx, *update.Email)
		if err != nil && err != domain.ErrNotFound {
			return nil, err
		}
		if emailUser != nil && emailUser.ID != userID {
			return nil, domain.ErrAlreadyExists
		}
	}

	return existing, nil
}

// validateDeleteUserRequest validates delete user request and returns existing user
func (s *Service) validateDeleteUserRequest(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
