# Development

## Prerequisites

- Docker & Docker Compose
- Make (optional but recommended)

## Setup

```bash
git clone <repository-url>
cd go-postgres-api-template

# Build and start the dev container + services
make workspace-build
make workspace-up
```

This starts PostgreSQL, LocalStack, and a workspace container with Go, golangci-lint, golang-migrate, and air pre-installed.

```bash
# Create SNS topics and SQS queues in LocalStack
make localstack-setup

# Run database migrations
make migrate
```

## Running Locally

Both the API server and worker use [air](https://github.com/air-verse/air) for live reload. Code changes trigger an automatic rebuild.

```bash
make run-api      # Start the API server on :8080
make run-worker   # Start the SQS event consumer
```

## Makefile Commands

### Code Quality

```bash
make format       # Format code with gofmt
make lint         # Run golangci-lint
make test         # Run all tests
```

### Go Modules

```bash
make mod-download   # Download dependencies
make mod-tidy       # Clean up go.mod/go.sum
make mod-verify     # Verify dependency checksums
```

### Database Migrations

Migrations are managed with [golang-migrate](https://github.com/golang-migrate/migrate) and stored in `resources/db/migrations/`.

```bash
make migrate                        # Run all pending migrations
make migrate-down                   # Rollback the last migration
make migrate-create NAME=add_index  # Create a new migration file pair
make migrate-version                # Show current migration version
```

### Docker Workspace

All dev commands run inside the workspace container:

```bash
make workspace-build    # Build the container image
make workspace-up       # Start the container + dependencies
make workspace-shell    # Open a shell in the running container
```

The container is defined in `resources/docker/workspace.Dockerfile` and includes the full Go toolchain.

### LocalStack

LocalStack emulates AWS SNS and SQS locally:

```bash
make localstack-up      # Start LocalStack
make localstack-setup   # Create topics, queues, and subscriptions
make localstack-logs    # Tail LocalStack logs
make localstack-down    # Stop LocalStack
```

The setup script (`resources/scripts/localstack_setup.sh`) creates:
- An SNS topic (`events-topic`)
- An SQS queue (`user-events`) subscribed to all `user.*` events via prefix filter
- A debug queue subscribed to all events (no filter)

## Configuration

Configuration is loaded from environment variables. `.env.local` is loaded first, then `.env`, then actual environment variables (highest priority).

Key variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | - |
| `AWS_REGION` | AWS region | `us-east-1` |
| `AWS_USE_LOCALSTACK` | Use LocalStack credentials | `false` |
| `AWS_ENDPOINT` | Custom AWS endpoint (for LocalStack) | - |
| `EVENTS_TOPIC_ARN` | SNS topic ARN for publishing events | - |
| `EVENTS_QUEUE_URL_USER` | SQS queue URL for user events | - |
| `SERVER_PORT` | HTTP server port | `8080` |
| `ENVIRONMENT` | Environment name | - |

## API Documentation

With the API server running:

- Swagger UI: http://localhost:8080/docs
- OpenAPI JSON: http://localhost:8080/openapi.json
- OpenAPI YAML: http://localhost:8080/openapi.yaml
