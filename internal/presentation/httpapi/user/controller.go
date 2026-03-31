package user

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/cgund98/go-postgres-api-template/api/v1/core"
	apiv1user "github.com/cgund98/go-postgres-api-template/api/v1/user"
	"github.com/cgund98/go-postgres-api-template/internal/domain/user"
	"github.com/cgund98/go-postgres-api-template/internal/presentation/httpapi"
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
	huma.Register(api, huma.Operation{
		Path:    "/api/v1/users",
		Method:  http.MethodPost,
		Summary: "Create a new user",
		Tags:    []string{"Users"},
	}, c.CreateUser)

	// Get user
	huma.Register(api, huma.Operation{
		Path:    "/api/v1/users/{id}",
		Method:  http.MethodGet,
		Summary: "Get a user by ID",
		Tags:    []string{"Users"},
	}, c.GetUser)

	// List users
	huma.Register(api, huma.Operation{
		Path:    "/api/v1/users",
		Method:  http.MethodGet,
		Summary: "List users",
		Tags:    []string{"Users"},
	}, c.ListUsers)

	// Update user
	huma.Register(api, huma.Operation{
		Path:    "/api/v1/users/{id}",
		Method:  http.MethodPatch,
		Summary: "Update a user",
		Tags:    []string{"Users"},
	}, c.UpdateUser)

	// Delete user
	huma.Register(api, huma.Operation{
		Path:    "/api/v1/users/{id}",
		Method:  http.MethodDelete,
		Summary: "Delete a user",
		Tags:    []string{"Users"},
	}, c.DeleteUser)
}

// CreateUser handles POST /api/v1/users
func (c *Controller) CreateUser(ctx context.Context, input *apiv1user.CreateUserInput) (*apiv1user.CreateUserOutput, error) {
	u, err := c.service.CreateUser(ctx, input.Body.Email, input.Body.FirstName, input.Body.LastName)
	if err != nil {
		return nil, httpapi.NewHumaError(err)
	}

	return &apiv1user.CreateUserOutput{
		Body: toUserResponse(u),
	}, nil
}

// GetUser handles GET /api/v1/users/{id}
func (c *Controller) GetUser(ctx context.Context, input *apiv1user.GetUserInput) (*apiv1user.GetUserOutput, error) {
	u, err := c.service.GetUser(ctx, input.ID)
	if err != nil {
		return nil, httpapi.NewHumaError(err)
	}

	return &apiv1user.GetUserOutput{
		Body: toUserResponse(u),
	}, nil
}

// ListUsers handles GET /api/v1/users
func (c *Controller) ListUsers(ctx context.Context, input *apiv1user.ListUsersInput) (*apiv1user.ListUsersOutput, error) {
	page := input.Page
	limit := input.Limit

	offset, normalizedLimit := httpapi.NormalizePagination(page, limit)
	users, total, err := c.service.ListUsers(ctx, normalizedLimit, offset)
	if err != nil {
		return nil, httpapi.NewHumaError(err)
	}

	var usersResponse []apiv1user.Response
	for _, u := range users {
		usersResponse = append(usersResponse, toUserResponse(u))
	}

	totalPages := httpapi.CalculateTotalPages(total, normalizedLimit)
	return &apiv1user.ListUsersOutput{
		Body: apiv1user.ListUsersResponse{
			Data: usersResponse,
			Pagination: core.PaginationResponse{
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
	update := &user.UpdateUserCommand{
		Email:     input.Body.Email,
		FirstName: input.Body.FirstName,
		LastName:  input.Body.LastName,
	}

	u, err := c.service.PatchUser(ctx, input.ID, update)
	if err != nil {
		return nil, httpapi.NewHumaError(err)
	}

	return &apiv1user.UpdateUserOutput{
		Body: toUserResponse(u),
	}, nil
}

// DeleteUser handles DELETE /api/v1/users/{id}
func (c *Controller) DeleteUser(ctx context.Context, input *apiv1user.DeleteUserInput) (*apiv1user.DeleteUserOutput, error) {
	err := c.service.DeleteUser(ctx, input.ID)
	if err != nil {
		return nil, httpapi.NewHumaError(err)
	}

	return &apiv1user.DeleteUserOutput{}, nil
}
