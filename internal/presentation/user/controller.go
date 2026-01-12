package user

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	apiv1user "github.com/cgund98/go-postgres-api-template/api/v1/user"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user/model"
	"github.com/cgund98/go-postgres-api-template/internal/presentation"
)

// Controller handles HTTP requests for user operations
type Controller struct {
	service *user.Service
}

// NewUserController creates a new UserController
func NewUserController(service *user.Service) *Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes registers all user routes with the Huma API
func (c *Controller) RegisterRoutes(api huma.API) {
	// Create user
	huma.Post(api, "/api/v1/users", c.CreateUser, huma.OperationTags("Users"))

	// Get user
	huma.Get(api, "/api/v1/users/{id}", c.GetUser, huma.OperationTags("Users"))

	// List users
	huma.Get(api, "/api/v1/users", c.ListUsers, huma.OperationTags("Users"))

	// Update user
	huma.Patch(api, "/api/v1/users/{id}", c.UpdateUser, huma.OperationTags("Users"))

	// Delete user
	huma.Delete(api, "/api/v1/users/{id}", c.DeleteUser, huma.OperationTags("Users"))
}

// CreateUser handles POST /api/v1/users
func (c *Controller) CreateUser(ctx context.Context, input *apiv1user.CreateUserInput) (*apiv1user.CreateUserOutput, error) {
	u, err := c.service.CreateUser(ctx, input.Body.Email, input.Body.FirstName, input.Body.LastName)
	if err != nil {
		return nil, presentation.NewHumaError(err)
	}

	return &apiv1user.CreateUserOutput{
		Body: *toUserResponse(u),
	}, nil
}

// GetUser handles GET /api/v1/users/{id}
func (c *Controller) GetUser(ctx context.Context, input *apiv1user.GetUserInput) (*apiv1user.GetUserOutput, error) {
	u, err := c.service.GetUser(ctx, input.ID)
	if err != nil {
		return nil, presentation.NewHumaError(err)
	}

	return &apiv1user.GetUserOutput{
		Body: *toUserResponse(u),
	}, nil
}

// ListUsers handles GET /api/v1/users
func (c *Controller) ListUsers(ctx context.Context, input *apiv1user.ListUsersInput) (*apiv1user.ListUsersOutput, error) {
	page := input.Page
	limit := input.Limit

	offset, normalizedLimit := presentation.NormalizePagination(page, limit)
	users, total, err := c.service.ListUsers(ctx, normalizedLimit, offset)
	if err != nil {
		return nil, presentation.NewHumaError(err)
	}

	totalPages := presentation.CalculateTotalPages(total, normalizedLimit)
	return &apiv1user.ListUsersOutput{
		Body: apiv1user.ListUsersResponse{
			Data: toUserResponseList(users),
			Pagination: presentation.PaginationResponse{
				Page:       page,
				Limit:      normalizedLimit,
				Total:      total,
				TotalPages: totalPages,
			},
		},
	}, nil
}

// UpdateUser handles PATCH /api/v1/users/{id}
func (c *Controller) UpdateUser(ctx context.Context, input *apiv1user.UpdateUserInput) (*apiv1user.UpdateUserOutput, error) {
	update := &model.UserUpdate{
		Email:     input.Body.Email,
		FirstName: input.Body.FirstName,
		LastName:  input.Body.LastName,
	}

	u, err := c.service.PatchUser(ctx, input.ID, update)
	if err != nil {
		return nil, presentation.NewHumaError(err)
	}

	return &apiv1user.UpdateUserOutput{
		Body: *toUserResponse(u),
	}, nil
}

// DeleteUser handles DELETE /api/v1/users/{id}
func (c *Controller) DeleteUser(ctx context.Context, input *apiv1user.DeleteUserInput) (*apiv1user.DeleteUserOutput, error) {
	err := c.service.DeleteUser(ctx, input.ID)
	if err != nil {
		return nil, presentation.NewHumaError(err)
	}

	return &apiv1user.DeleteUserOutput{}, nil
}
