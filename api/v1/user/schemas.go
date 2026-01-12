package user

import "github.com/cgund98/go-postgres-api-template/internal/presentation"

// Input/Output types for Huma operations
// These define the public API contract

// Response represents a user in API responses
type Response struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateUserInput struct {
	Body struct {
		Email     string `json:"email" doc:"User email address" example:"user@example.com" maxLength:"255"`
		FirstName string `json:"first_name" doc:"User first name" example:"John" minLength:"1" maxLength:"100"`
		LastName  string `json:"last_name" doc:"User last name" example:"Doe" minLength:"1" maxLength:"100"`
	}
}

type CreateUserOutput struct {
	Body Response `json:"body"`
}

type GetUserInput struct {
	ID string `path:"id" doc:"User ID" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type GetUserOutput struct {
	Body Response `json:"body"`
}

type ListUsersInput struct {
	Page  int `query:"page" doc:"Page number (1-based)" example:"1" minimum:"1"`
	Limit int `query:"limit" doc:"Number of items per page" example:"10" minimum:"1" maximum:"100"`
}

type ListUsersOutput struct {
	Body ListUsersResponse `json:"body"`
}

type ListUsersResponse struct {
	Data       []Response                      `json:"data"`
	Pagination presentation.PaginationResponse `json:"pagination"`
}

type UpdateUserInput struct {
	ID   string `path:"id" doc:"User ID" example:"123e4567-e89b-12d3-a456-426614174000"`
	Body struct {
		Email     *string `json:"email,omitempty" doc:"User email address" example:"user@example.com" maxLength:"255"`
		FirstName *string `json:"first_name,omitempty" doc:"User first name" example:"John" minLength:"1" maxLength:"100"`
		LastName  *string `json:"last_name,omitempty" doc:"User last name" example:"Doe" minLength:"1" maxLength:"100"`
	}
}

type UpdateUserOutput struct {
	Body Response `json:"body"`
}

type DeleteUserInput struct {
	ID string `path:"id" doc:"User ID" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type DeleteUserOutput struct {
}
