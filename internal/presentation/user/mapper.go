package user

import (
	apiv1user "github.com/cgund98/go-postgres-api-template/api/v1/user"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/model"
)

// ToUserResponse converts a domain user to a response DTO
func toUserResponse(u *model.User) *apiv1user.Response {
	return &apiv1user.Response{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToUserResponseList converts a list of domain users to response DTOs
func toUserResponseList(users []*model.User) []apiv1user.Response {
	responses := make([]apiv1user.Response, len(users))
	for i, u := range users {
		responses[i] = *toUserResponse(u)
	}
	return responses
}
