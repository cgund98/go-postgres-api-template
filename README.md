# Go PostgreSQL API Template

A Go backend template with a 3-tier architecture, event-driven messaging, and a containerized development environment.

[![Go](https://img.shields.io/badge/Go-1.25.0+-00ADD8.svg)](https://go.dev/)
[![Huma](https://img.shields.io/badge/Huma-v2-green.svg)](https://huma.rocks/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-blue.svg)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Tech Stack

| Category | Technology |
|----------|-----------|
| HTTP | [Huma v2](https://huma.rocks/) on [Chi](https://github.com/go-chi/chi) |
| Database | PostgreSQL via [pgx v5](https://github.com/jackc/pgx) |
| Messaging | AWS SNS/SQS via [AWS SDK v2](https://github.com/aws/aws-sdk-go-v2) |
| Config | [envconfig](https://github.com/kelseyhightower/envconfig) + [godotenv](https://github.com/joho/godotenv) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Logging | `log/slog` (structured JSON) |

## Quick Start

```bash
git clone <repository-url>
cd go-postgres-api-template

make workspace-build    # Build the dev container
make workspace-up       # Start all services (Postgres, LocalStack, workspace)
make localstack-setup   # Create SNS topics and SQS queues
make migrate            # Run database migrations
make run-api            # Start the API server (with live reload)
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for the full development guide.

## What's Included

The template ships with a working **User** domain to demonstrate real patterns:

- CRUD API with validation and pagination (`POST/GET/PATCH/DELETE /api/v1/users`)
- Transaction management through `context.Context`
- Event publishing on create/update/delete (`user.created.v1`, `user.updated.v1`, `user.deleted.v1`)
- SQS worker that consumes events via an envelope-based router
- Automatic OpenAPI docs at `/docs`

## Project Structure

```
go-postgres-api-template/
├── api/v1/                         # Public API schemas (input/output types)
├── cmd/
│   ├── api/                        # API server entrypoint
│   └── worker/                     # Event consumer entrypoint
├── internal/
│   ├── config/                     # Configuration loading
│   ├── domain/                     # Business logic, models, interfaces
│   │   ├── events/                 # Event system (router, publisher, envelope registry)
│   │   └── user/                   # User domain (service, repo interface, handlers)
│   ├── adapters/                   # Infrastructure implementations
│   │   ├── aws/                    # AWS SDK v2 config + SQS interface
│   │   ├── db/                     # Transaction manager + Postgres implementation
│   │   ├── events/                 # SNS publisher + SQS consumer
│   │   └── user/                   # Postgres user repository
│   ├── presentation/httpapi/       # HTTP controllers, routing, middleware
│   └── observability/              # Logging setup
├── resources/
│   ├── db/migrations/              # SQL migration files
│   ├── docker/                     # Dockerfiles
│   └── scripts/                    # Setup scripts (migrations, LocalStack)
└── tests/integration/              # Integration tests
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for a detailed breakdown of each layer.

## API Endpoints

Once running, API docs are available at `http://localhost:8080/docs`.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/users` | Create a user |
| `GET` | `/api/v1/users` | List users (paginated) |
| `GET` | `/api/v1/users/{id}` | Get a user by ID |
| `PATCH` | `/api/v1/users/{id}` | Update a user |
| `DELETE` | `/api/v1/users/{id}` | Delete a user |

## Contributing

1. Follow the existing code structure
2. Write tests for new features
3. Run `make lint` before committing
4. Keep API schemas in `api/v1/` and implementation in `internal/`
5. Update documentation for architectural changes

## License

MIT - see [LICENSE](LICENSE) for details.
