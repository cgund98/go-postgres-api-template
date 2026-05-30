package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cgund98/go-postgres-api-template/internal/domain/events"
	"github.com/cgund98/go-postgres-api-template/internal/domain/events/registry"
	userEventsV1 "github.com/cgund98/go-postgres-api-template/internal/domain/events/registry/users/v1"
	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

/** -------------------------------- CreateUserHandler -------------------------------- */

type CreateUserHandler struct {
}

func NewCreateUserHandler() *CreateUserHandler {
	return &CreateUserHandler{}
}

func (h *CreateUserHandler) Handle(ctx context.Context, envelope registry.Envelope) error {
	var userCreatedEvent userEventsV1.UserCreatedEvent
	if err := json.Unmarshal(envelope.Data, &userCreatedEvent); err != nil {
		return fmt.Errorf("failed to unmarshal user created event: %w", err)
	}

	logger := observability.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "Handling user created event", "user_id", userCreatedEvent.UserID, "email", userCreatedEvent.Email)

	return nil
}

var _ events.Handler = &CreateUserHandler{}

/** -------------------------------- UpdateUserHandler -------------------------------- */

type UpdateUserHandler struct {
}

func NewUpdateUserHandler() *UpdateUserHandler {
	return &UpdateUserHandler{}
}

func (h *UpdateUserHandler) Handle(ctx context.Context, envelope registry.Envelope) error {
	var userUpdatedEvent userEventsV1.UserUpdatedEvent
	if err := json.Unmarshal(envelope.Data, &userUpdatedEvent); err != nil {
		return fmt.Errorf("failed to unmarshal user updated event: %w", err)
	}

	logger := observability.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "Handling user updated event", "user_id", userUpdatedEvent.UserID, "changes", userUpdatedEvent.Changes)

	return nil
}

var _ events.Handler = &UpdateUserHandler{}

/** -------------------------------- DeleteUserHandler -------------------------------- */

type DeleteUserHandler struct {
}

func NewDeleteUserHandler() *DeleteUserHandler {
	return &DeleteUserHandler{}
}

func (h *DeleteUserHandler) Handle(ctx context.Context, envelope registry.Envelope) error {
	var userDeletedEvent userEventsV1.UserDeletedEvent
	if err := json.Unmarshal(envelope.Data, &userDeletedEvent); err != nil {
		return fmt.Errorf("failed to unmarshal user deleted event: %w", err)
	}

	logger := observability.LoggerFromContext(ctx)
	logger.InfoContext(ctx, "Handling user deleted event", "user_id", userDeletedEvent.UserID)

	return nil
}

var _ events.Handler = &DeleteUserHandler{}
