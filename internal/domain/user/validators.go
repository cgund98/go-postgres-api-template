package user

import (
	"fmt"
	"strings"

	"github.com/cgund98/go-postgres-api-template/internal/domain"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/model"
)

// ValidateCreate validates UserCreate data
func ValidateCreate(u *model.UserCreate) error {
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
func ValidateUpdate(u *model.UserUpdate) error {
	if u.Email != nil && !isValidEmail(*u.Email) {
		return fmt.Errorf("%w: invalid email format", domain.ErrInvalidInput)
	}
	return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 3 {
		return false
	}
	if !strings.Contains(email, "@") {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	return true
}
