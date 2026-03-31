package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/cgund98/go-postgres-api-template/internal/domain"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/model"
)

const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

var emailRegexp = regexp.MustCompile(emailRegex)

// ValidateCreate validates UserCreate data
func ValidateCreate(u *model.CreateUserCommand) error {
	if u.Email == "" {
		return fmt.Errorf("%w: email is required", domain.ErrInvalidInput)
	}
	if !isValidEmail(u.Email) {
		return fmt.Errorf("%w: invalid email format", domain.ErrInvalidInput)
	}
	if u.FirstName == "" {
		return fmt.Errorf("%w: first_name is required", domain.ErrInvalidInput)
	}
	if u.LastName == "" {
		return fmt.Errorf("%w: last_name is required", domain.ErrInvalidInput)
	}
	return nil
}

// ValidateUpdate validates UserUpdate data
func ValidateUpdate(u *model.UpdateUserCommand) error {
	if u.Email != nil && !isValidEmail(*u.Email) {
		return fmt.Errorf("%w: invalid email format", domain.ErrInvalidInput)
	}
	return nil
}

// isValidEmail performs basic email validation with regex
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 3 {
		return false
	}
	return emailRegexp.MatchString(email)
}

/** Validate User Requests */

// validateCreateUserRequest validates create user request
func (s *Service) validateCreateUserRequest(ctx context.Context, email, firstName, lastName string) error {
	// Validate input
	if err := ValidateCreate(&model.CreateUserCommand{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}); err != nil {
		return err
	}

	// Check email uniqueness
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("failed to get user by email: %w", err)
	}
	if existing != nil {
		return domain.ErrAlreadyExists
	}

	return nil
}

// validatePatchUserRequest validates patch user request and returns existing user
func (s *Service) validatePatchUserRequest(ctx context.Context, userID string, update *model.UpdateUserCommand) (*model.User, error) {
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
		if err != nil {
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
