# Project Instructions

This repository uses a 3-tier Go architecture:

- `internal/presentation/httpapi/` for HTTP routing, validation, and response mapping
- `internal/domain/` for business logic, domain models, and interfaces
- `internal/adapters/` for infrastructure implementations such as Postgres, AWS, and messaging

Both Codex and Cursor should treat this file as the primary project guide. Start by reading:

1. `README.md`
2. `ARCHITECTURE.md`
3. `DEVELOPMENT.md`
4. The closest existing feature module, usually `internal/presentation/httpapi/user/` and `internal/domain/user/`

## Working Rules

- Keep business logic in the domain layer.
- Keep controllers thin: they should validate, call services, and map results.
- Keep repositories focused on data access only.
- Keep API contracts in `api/v1/` and internal implementation in `internal/`.
- Prefer the existing user feature as the reference pattern for new work.
- Do not introduce new frameworks or abstractions unless the existing architecture clearly needs them.
- If you need to touch wiring, update the dependency setup in `internal/presentation/httpapi/deps.go` and the API bootstrap in `cmd/api/main.go`.
- If a change affects runtime behavior, update the relevant tests and documentation.

## Adding a New API Endpoint

Use the user feature as the template and add the new endpoint in this order:

1. Define the public request and response types in `api/v1/<domain>/schemas.go`.
2. Add or extend the domain model, commands, and validators in `internal/domain/<domain>/`.
3. Add or extend the repository interface in `internal/domain/<domain>/repo.go`.
4. Implement the repository in `internal/adapters/<domain>/repo.go` if the feature needs persistence.
5. Add the service method in `internal/domain/<domain>/service.go`.
6. Add the controller method in `internal/presentation/httpapi/<domain>/controller.go`.
7. Register the route in the controller’s `RegisterRoutes` method.
8. Wire the dependency into `internal/presentation/httpapi/deps.go`.
9. Instantiate and register the controller in `cmd/api/main.go`.
10. Add or update tests for the domain, adapter, controller, and any integration path that changed.

## Layering Guidance

- Presentation layer:
  - Handle HTTP path, query, and body input.
  - Translate domain errors into Huma errors with the existing helpers.
  - Map domain models to API response DTOs with small mapper functions.

- Domain layer:
  - Own business rules, validation, and transaction boundaries.
  - Depend on interfaces, not concrete infrastructure types.
  - Return domain errors for expected failure cases such as not found or invalid input.

- Adapters layer:
  - Implement interfaces defined by the domain.
  - Contain Postgres, AWS, or other external-system details.
  - Keep SQL and SDK-specific logic out of domain and controller code.

## API Conventions

- Use Huma operations and typed input/output structs.
- Keep route paths versioned under `/api/v1/...`.
- Use `path`, `query`, and `json` tags consistently with the existing user API.
- Keep list endpoints paginated when the resource can grow.
- Use the existing error translation and pagination helpers rather than duplicating logic.

## Testing and Verification

- Run the smallest useful test set first, then broader tests if the change is cross-cutting.
- Prefer tests near the layer that changed:
  - domain tests for business rules
  - adapter tests for persistence or integration behavior
  - controller tests for request/response mapping
- Format Go files with `gofmt` before finishing.
- If a change is larger, run the relevant `make` target documented in `DEVELOPMENT.md`.

## Event-Driven Changes

- If the feature emits or consumes events, add the event schema under `internal/domain/events/registry/`.
- Register new handlers in the worker bootstrap if the event consumer needs to know about them.
- Keep event payloads versioned and explicit.

## Practical Preference

- When unsure, mirror the user domain implementation first and make the smallest possible change that fits the current code style.
