package user

import "github.com/cgund98/go-postgres-api-template/internal/domain/user/model"

// Changes represents field changes for an update event
type Changes map[string]any

// GenerateUserChanges generates a changes map from UserUpdate and existing User
// Only includes fields that are being updated (non-nil in UserUpdate)
func GenerateUserChanges(update *model.UserUpdate, existing *model.User) Changes {
	changes := make(Changes)

	if update.Email != nil && *update.Email != existing.Email {
		changes["email"] = map[string]any{
			"old": existing.Email,
			"new": *update.Email,
		}
	}
	if update.FirstName != nil && *update.FirstName != existing.FirstName {
		changes["first_name"] = map[string]any{
			"old": existing.FirstName,
			"new": *update.FirstName,
		}
	}
	if update.LastName != nil && *update.LastName != existing.LastName {
		changes["last_name"] = map[string]any{
			"old": existing.LastName,
			"new": *update.LastName,
		}
	}

	return changes
}
