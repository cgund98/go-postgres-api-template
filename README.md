# 🚀 Go PostgreSQL API Template

<div align="center">

**A production-ready Go backend template demonstrating efficient organization and modern design patterns**

[![Go](https://img.shields.io/badge/Go-1.24.3+-00ADD8.svg)](https://go.dev/)
[![Huma](https://img.shields.io/badge/Huma-v2-green.svg)](https://huma.rocks/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-blue.svg)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

*Built with ❤️ using Domain-Driven Design and Event-Driven Architecture*

[Features](#-features) • [Quick Start](#-getting-started) • [Project Structure](#-project-structure) • [Example](#-example-business-case)

</div>

## 👥 Who Is This For?

This template is designed for **backend engineers** building:

- RESTful APIs with clean separation between HTTP handlers, business logic, and data access
- Event-driven microservices with async message processing (SQS/SNS)
- Type-safe codebases with comprehensive error handling and unit tests
- Production-ready systems with proper transaction management and observability
- Unified development environments using Docker containers

## ✨ Features

### 🏗️ Architecture & Design Patterns

- **3-Tier Architecture**: Clear separation between presentation, domain, and infrastructure layers
- **Domain-Driven Design**: Business logic encapsulated in domain services with rich domain models
- **Repository Pattern**: Data access abstraction with generic database context for database-agnostic code
- **Transaction Management**: Application-level transaction management using context.Context
- **Event-Driven Architecture**: Decoupled communication via domain events and message queues
- **Dependency Injection**: Loose coupling through constructor injection, enabling testability
- **Generic Database Context**: Type-safe abstraction over database implementations

### 🛠️ Technology Choices

- **Huma v2**: Modern HTTP framework with automatic OpenAPI/Swagger generation and schema validation
- **Chi**: Lightweight router that Huma mounts to, providing flexibility for non-Huma routes
- **PostgreSQL**: Direct SQL queries using `database/sql` with connection pooling
- **golang-migrate**: Database migration management
- **Viper**: Structured configuration from environment variables and `.env` files
- **Structured Logging**: JSON-formatted logs with `slog` for observability

### 📊 Observability & Operations

- **Structured Logging**: JSON-formatted logs with structured fields for easy parsing
- **Health Checks**: Built-in health check endpoints for monitoring
- **OpenAPI Documentation**: Automatic API documentation at `/docs` and `/openapi.json`

### 👨‍💻 Developer Experience

- **Docker Workspace**: Unified development environment in a containerized workspace
- **Modern Tooling**: golangci-lint for comprehensive linting and code quality
- **Comprehensive Testing**: Unit tests with mocked dependencies for fast, reliable test suites
- **LocalStack Integration**: Local AWS service emulation for SNS/SQS development
- **Makefile Commands**: Convenient commands for common development tasks

## 🏛️ Architecture

This template follows a **3-tier architecture** with clear separation of concerns:

### Presentation Layer (`internal/presentation/`)
- **Controllers**: Route handlers that coordinate between API schemas and domain services
- **Router**: Chi router with Huma API integration for automatic OpenAPI generation
- **Mappers**: Conversion between domain models and API response DTOs

### API Schemas (`api/v1/`)
- **Public Contract**: Input/output types that define the public API contract
- **Versioned**: Organized by API version (`v1`, `v2`, etc.) for backward compatibility
- **Importable**: Can be imported by clients or SDK generators

### Domain Layer (`internal/domain/`)
- **Models**: Domain entities with business logic
- **Services**: Domain-specific business logic with integrated transaction management
- **Repositories**: Data access interfaces using generic database context
- **Events**: Domain events for event-driven communication
- **Handlers**: Event handlers for processing domain events

### Infrastructure Layer (`internal/infrastructure/`)
- **Database**: PostgreSQL connection pooling and transaction management
- **Messaging**: Event publishing (SNS) and consumption (SQS) with generic type support
- **AWS**: AWS SDK integration with LocalStack support
- **Events**: Generic event system with JSON serialization/deserialization

### Observability (`internal/observability/`)
- **Logging**: Structured logging with slog

## 📁 Project Structure

```
go-postgres-api-template/
├── api/
│   └── v1/
│       └── user/
│           └── schemas.go          # Public API contract (input/output types)
│
├── cmd/
│   ├── api/
│   │   └── main.go                 # API server entrypoint
│   └── worker/
│       └── main.go                 # Event consumer entrypoint
│
├── internal/
│   ├── config/
│   │   └── settings.go             # Configuration loading with Viper
│   │
│   ├── domain/
│   │   ├── errors.go               # Domain error definitions
│   │   └── user/                   # User domain
│   │       ├── model/
│   │       │   └── model.go        # Domain models
│   │       ├── events/
│   │       │   ├── schema.go       # Domain events
│   │       │   └── handlers/       # Event handlers
│   │       ├── repo/               # Repository implementations
│   │       ├── service.go          # Domain service
│   │       └── validators.go       # Domain validation
│   │
│   ├── infrastructure/
│   │   ├── db/
│   │   │   ├── context.go          # Generic database context interface
│   │   │   └── postgres/           # PostgreSQL implementation
│   │   │       ├── context.go
│   │   │       ├── pool.go
│   │   │       └── transaction-manager.go
│   │   ├── events/
│   │   │   ├── base.go             # Event interface
│   │   │   ├── consumer/           # SQS consumer
│   │   │   ├── publisher/          # SNS publisher
│   │   │   ├── serializer/         # JSON serializer
│   │   │   └── deserializer/       # JSON deserializer
│   │   └── aws/                    # AWS SDK helpers
│   │
│   ├── presentation/
│   │   ├── router.go               # Chi router with Huma integration
│   │   ├── deps.go                 # Dependency injection
│   │   └── user/
│   │       ├── controller.go       # Route handlers
│   │       └── mapper.go           # Domain to API mapping
│   │
│   └── observability/
│       └── logging.go              # Structured logging setup
│
├── resources/
│   ├── db/migrations/              # Database migrations
│   ├── docker/
│   │   └── workspace.Dockerfile    # Development workspace container
│   └── scripts/
│       ├── migrate.sh              # Migration helper script
│       └── localstack_setup.sh     # LocalStack resource setup
│
├── tests/
│   └── integration/                # Integration tests
│
├── docker-compose.yml              # Local development services
├── Makefile                        # Development commands
├── go.mod
└── README.md
```

## 🚀 Getting Started

### Prerequisites

- Docker & Docker Compose
- Make (optional, but recommended)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go-postgres-api-template
```

2. Start the workspace container:
```bash
make workspace-build
make workspace-up
```

3. Set up LocalStack (for local development):
```bash
# Wait for LocalStack to be healthy, then run setup script
make localstack-setup
```

5. Set up the database:
```bash
# Start PostgreSQL
docker compose up -d postgres

# Run migrations
make migrate
```

### Running Locally

We run both of our applications locally with [air](https://github.com/air-verse/air) for live-reload functionality.

If you make a code change while running the server, it will re-build and restart automatically.

**API Server:**
```bash
make run-api
```

**Worker:**
```bash
make run-worker
```

Both commands run inside the workspace Docker container, ensuring a consistent development environment.

### Development

**Formatting:**
```bash
make format
```

**Linting:**
```bash
make lint
```

**Testing:**
```bash
make test
```

**Go Module Management:**
```bash
make mod-download  # Download dependencies
make mod-tidy      # Clean up dependencies
make mod-verify    # Verify dependencies
```

**Database Migrations:**
```bash
make migrate              # Run migrations (up)
make migrate-down         # Rollback last migration
make migrate-create NAME=my_migration  # Create new migration
make migrate-version      # Show current migration version
```

### 🔧 Transaction Management Pattern

The template uses a **transaction management pattern** that passes database transactions through `context.Context`:

- **`db.TransactionManager`**: Generic interface for transaction management over a database context type
- **`postgres.TransactionManager`**: PostgreSQL implementation that stores `*sql.Tx` transactions in context
- **`postgres.GetTXFromContext()`**: Package-level function that extracts the transaction from context
- **Repository Pattern**: Repositories extract transactions directly from `context.Context` using `postgres.GetTXFromContext()`

This pattern allows:
- **Transaction Management**: Application-level transactions managed via context
- **Type Safety**: Compile-time guarantees for database operations
- **Testability**: Easy to mock transaction managers in tests
- **Implicit Transaction Passing**: Transactions flow through context without explicit parameters

**Example usage:**
```go
// Service uses an abstract TransactionManager interface
type Service struct {
    repo      repo.Repository
    txManager TransactionManager  // Domain interface
}

// Transaction manager handles transactions
err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
    // Repository extracts transaction from context internally
    user, err := s.repo.GetByID(txCtx, userID)
    return err
})
```

**Repository implementation:**
```go
// Repository extracts transaction directly from context
func (r *PostgresRepository) GetByID(ctx context.Context, userID string) (*model.User, error) {
    tx := postgres.GetTXFromContext(ctx)
    if tx == nil {
        return nil, db.ErrNoDBContext
    }
    // Use tx directly for database operations
    err := tx.QueryRow(query, userID).Scan(...)
    // ...
}
```

**Key points:**
- Transactions are stored in context using an unexported key (`txKey`) to prevent collisions
- The same `txKey` must be used by both `TransactionManager.WithTransaction()` and `GetTXFromContext()`
- Repositories are responsible for extracting the transaction from context
- If no transaction is found, repositories return `db.ErrNoDBContext`

## 💼 Example Business Case

This template implements a **User Management System** to demonstrate real-world patterns:

### 👤 User Domain
- **User Registration**: Create users with email uniqueness validation
- **User Updates**: Update user information (e.g., name changes)
- **User Deletion**: Delete users with proper cleanup
- **Event Publishing**: Emits `UserCreatedEvent`, `UserUpdatedEvent`, and `UserDeletedEvent` for downstream processing

### 🔑 Key Patterns Demonstrated
- **Transaction Management**: All operations use application-level transactions via context
- **Event-Driven Processing**: Async workers consume events from SQS queues
- **Business Rule Enforcement**: Prevents invalid states (e.g., duplicate emails)
- **Domain Events**: Decoupled communication between services via events
- **Error Handling**: Proper error wrapping and HTTP status code mapping

## 🗄️ Database Schema

The template includes a User domain. Database migrations are managed using [golang-migrate](https://github.com/golang-migrate/migrate) and located in `resources/db/migrations/`.

To run migrations:
```bash
make migrate
```

The schema includes:

- **Users table**: Stores user information with email uniqueness, timestamps, and UUID primary keys

## 📨 Event System

The template includes a **generic event system** for decoupled communication:

| Component | Purpose |
|-----------|---------|
| **Event Interface** | Base interface for all domain events with metadata (`EventID`, `AggregateID`, `Timestamp`) |
| **Event Publishing** | Events published to SNS topics with JSON serialization |
| **Event Consumption** | Generic SQS consumers with type-safe deserialization |
| **Event Handlers** | Domain-specific consumers process events |

**Example: Publishing a domain event**

```go
// In UserService.CreateUser()
event := events.NewUserCreatedEvent(user.ID, user.Email)
err := s.eventPublisher.Publish(ctx, event)
```

**Example: Consuming events in a worker**

```go
// Worker automatically processes events from SQS
consumer := consumer.NewSQSConsumer[*UserCreatedEvent](sqsClient, options)
consumer.Start(ctx, deserializer, handler)
```

## 🐳 Docker

The project uses a **workspace Docker container** for unified development:

```bash
# Build the workspace container
make workspace-build

# Start the workspace container
make workspace-up

# Run commands inside the workspace
make format
make lint
make run-api
make run-worker

# Open a shell in the workspace
make workspace-shell
```

The workspace container includes:
- Go toolchain
- golangci-lint
- golang-migrate
- All development dependencies

This ensures all developers have the same environment regardless of their local setup.

## ☁️ LocalStack Integration

For local development, the template uses **LocalStack** to emulate AWS services:

```bash
# Start LocalStack
make localstack-up

# Set up SNS topics and SQS queues
make localstack-setup

# View logs
make localstack-logs

# Stop LocalStack
make localstack-down
```

LocalStack provides:
- SNS topic emulation for event publishing
- SQS queue emulation for event consumption
- Automatic resource setup via script

## 🎯 Design Patterns

This template demonstrates the following patterns:

- **3-Tier Architecture**: Clear separation of concerns
- **Repository Pattern**: Data access abstraction with generic database context
- **Unit of Work**: Transaction management via context.Context
- **Domain Events**: Event-driven communication
- **Dependency Injection**: Loose coupling through constructors
- **API Versioning**: Public schemas (`api/v1/`) separate from implementation (`internal/`)

## 📚 API Documentation

Once the API server is running, access:

- **Swagger UI**: `http://localhost:8080/docs`
- **OpenAPI JSON**: `http://localhost:8080/openapi.json`
- **OpenAPI YAML**: `http://localhost:8080/openapi.yaml`
- **Schemas**: `http://localhost:8080/schemas`

All API endpoints are versioned under `/api/v1/`:
- `POST /api/v1/users` - Create a user
- `GET /api/v1/users` - List users with pagination
- `GET /api/v1/users/{id}` - Get a user by ID
- `PATCH /api/v1/users/{id}` - Update a user
- `DELETE /api/v1/users/{id}` - Delete a user

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. ✅ Follow the existing code structure and patterns
2. ✅ Maintain type safety throughout
3. ✅ Write tests for new features with mocked dependencies
4. ✅ Run linting before committing (`make lint`)
5. ✅ Follow the Makefile commands for common tasks
6. ✅ Update documentation for any architectural changes
7. ✅ Keep API schemas in `api/v1/` and implementation in `internal/`

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ for the Go backend community**

⭐ Star this repo if you find it useful!

</div>
