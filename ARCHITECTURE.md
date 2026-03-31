# Architecture

The codebase follows a 3-tier architecture: **presentation**, **domain**, and **adapters** (infrastructure). Dependencies point inward — the domain layer has no knowledge of HTTP, databases, or AWS.

## Presentation Layer

`internal/presentation/httpapi/`

Handles HTTP concerns: routing, request/response mapping, error translation.

- **Router** (`router.go`): Chi router with Huma mounted on top. Huma provides automatic OpenAPI generation and request validation from struct tags.
- **Controllers** (`user/controller.go`): Thin handlers that call domain services and map results to API responses. No business logic lives here.
- **Mappers** (`user/mapper.go`): Convert between domain models and API response DTOs.
- **Error Mapping** (`errors.go`): Translates domain errors (`ErrNotFound`, `ErrInvalidInput`) into HTTP status codes.

API schemas live in `api/v1/` — they define the public contract (request/response types) and are separate from internal types so they can be versioned independently.

## Domain Layer

`internal/domain/`

Contains all business logic. No imports from `adapters/` or `presentation/`.

### Models and Services

Each domain (e.g., `user/`) contains:

- **Model** (`model.go`): Domain entities as plain structs. Service methods return values `(User, error)`, not pointers — errors communicate absence, not nil.
- **Repository interface** (`repo.go`): Defines data access operations. The interface lives next to the model it serves.
- **Service** (`service.go`): Orchestrates business logic within transactions. Depends on the repository interface and a `TransactionManager` interface.
- **Validators** (`validators.go`): Input validation and business rule enforcement (e.g., email uniqueness).
- **Handlers** (`handler.go`): Event handlers that process incoming envelopes for this domain.

### Transaction Management

Transactions flow through `context.Context`:

```go
err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
    user, err := s.repo.GetByID(txCtx, userID)
    // ... more repo calls using txCtx share the same transaction
    return err
})
```

`TransactionManager` is a domain interface. The Postgres implementation (`adapters/db/postgres/`) stores a `pgx.Tx` in the context. Repositories extract it via `postgres.GetTXFromContext(ctx)`.

This means:
- Services don't know about pgx or SQL
- Multiple repo calls within a `WithTransaction` block share the same transaction
- Easy to mock in tests

### Event System

Events use a [CloudEvents](https://cloudevents.io/)-inspired envelope format:

```
Envelope { id, type, source, time, data, attributes }
    └── data: JSON-encoded Payload (e.g., UserCreatedEvent)
```

**Publishing**: Services create a `PublishArgs` with a typed payload. The SNS publisher wraps it in an envelope and sends it to SNS with `event_type` as a message attribute for subscription filtering.

```go
publishArgs := events.PublishArgs{
    Payload:  &v1.UserCreatedEvent{UserID: user.ID, Email: user.Email},
    Metadata: events.PublishMetadata{Source: "user-service"},
}
s.eventPublisher.Publish(ctx, publishArgs)
```

**Consuming**: The SQS consumer receives raw messages, unmarshals envelopes, and routes to the correct handler by event type:

```go
router := events.NewRouter()
router.RegisterHandler(v1.EventTypeUserCreated, user.NewCreateUserHandler())

consumer := consumer.NewSQSConsumer(sqsClient, options)
consumer.Start(ctx, router)
```

**Event schemas** are versioned under `domain/events/registry/users/v1/`. Each event implements the `Payload` interface (`EventType()`, `AggregateID()`).

## Adapters Layer

`internal/adapters/`

Implements the interfaces defined in the domain layer.

### Database (`adapters/db/`)

- **Transaction manager interface** (`transaction-manager.go`): Abstract interface.
- **Postgres implementation** (`postgres/`): Uses `pgxpool` for connection pooling. The transaction manager begins a `pgx.Tx`, stores it in context, and commits/rolls back when the callback returns.

### User Repository (`adapters/user/`)

Implements `user.Repository` using pgx. Uses `pgx.CollectOneRow` and `pgx.RowToStructByName` for struct scanning — no manual `Scan()` calls.

### Messaging (`adapters/events/`)

- **SNS Publisher** (`publisher/sns.go`): Serializes envelopes to JSON and publishes to SNS using `PublishBatch`. Sets `event_type` as a message attribute so SNS subscriptions can filter.
- **SQS Consumer** (`consumer/sqs.go`): Long-polls SQS, unmarshals envelopes, routes to handlers, and acks on success. Runs in a goroutine with graceful shutdown via context cancellation.

### AWS (`adapters/aws/`)

- **Config loader** (`clients.go`): Creates an `aws.Config` using the v2 SDK's `config.LoadDefaultConfig`. Supports static credentials for LocalStack.
- **SQS interface** (`sqs.go`): Defines `SQSClientInterface` so the real `sqs.Client` can be swapped for a mock in tests.
