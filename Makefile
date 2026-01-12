.PHONY: workspace-up workspace-down workspace-build format lint run-api run-worker run-api-watch run-worker-watch build-api build-worker mod-download mod-tidy mod-verify localstack-start localstack-setup localstack-stop localstack-logs migrate migrate-down migrate-create migrate-version

# Docker Compose service name
SERVICE := workspace

# Start the workspace container
workspace-up:
	docker compose up -d $(SERVICE)

# Build the workspace container
workspace-build:
	docker compose build $(SERVICE)

# Stop and remove the workspace container
workspace-down:
	docker compose down


# Open a shell in the workspace container
workspace-shell:
	docker compose exec $(SERVICE) /bin/bash

# Format Go code
format:
	docker compose exec $(SERVICE) go fmt ./...

# Run linter
lint:
	docker compose exec $(SERVICE) golangci-lint run --fix

# Run the API server
run-api:
	docker compose exec $(SERVICE) air -c .air.api.toml

# Run the worker
run-worker:
	docker compose exec $(SERVICE) air -c .air.worker.toml

# Build the API binary
build-api:
	docker compose exec $(SERVICE) go build -o bin/api ./cmd/api

# Build the worker binary
build-worker:
	docker compose exec $(SERVICE) go build -o bin/worker ./cmd/worker

# Build all binaries
build: build-api build-worker

# Run tests
test:
	docker compose exec $(SERVICE) go test ./...

# Get dependencies
deps:
	docker compose exec $(SERVICE) go mod download && docker compose exec $(SERVICE) go mod tidy

# Go mod commands
mod-download:
	docker compose exec $(SERVICE) go mod download

mod-tidy:
	docker compose exec $(SERVICE) go mod tidy

mod-verify:
	docker compose exec $(SERVICE) go mod verify

# Enter the workspace shell
shell:
	docker compose exec $(SERVICE) /bin/bash

# LocalStack commands
localstack-up:
	@echo "Starting LocalStack..."
	docker compose up -d localstack
	@echo "Waiting for LocalStack to be ready..."
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if docker compose exec -T localstack curl -f http://localhost:4566/_localstack/health >/dev/null 2>&1; then \
			echo "LocalStack is ready!"; \
			exit 0; \
		fi; \
		sleep 2; \
		timeout=$$((timeout - 2)); \
	done; \
	echo "Warning: LocalStack may not be fully ready yet"

localstack-setup: localstack-up
	@echo "Setting up LocalStack resources (SNS topics and SQS queues)..."
	@docker compose exec $(SERVICE) bash resources/scripts/localstack_setup.sh

localstack-down:
	@echo "Stopping LocalStack..."
	docker compose stop localstack

localstack-logs:
	docker compose logs -f localstack

# Database migration commands
migrate: ## Run database migrations (up)
	docker compose exec $(SERVICE) bash resources/scripts/migrate.sh up

migrate-down: ## Rollback last migration
	docker compose exec $(SERVICE) bash resources/scripts/migrate.sh down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=my_migration)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=my_migration"; \
		exit 1; \
	fi
	docker compose exec $(SERVICE) bash resources/scripts/migrate.sh create $(NAME)

migrate-version: ## Show current migration version
	docker compose exec $(SERVICE) bash resources/scripts/migrate.sh version
