-include .env
export

PROJECT_ROOT := $(shell pwd)
export LOGGER_FOLDER ?= $(PROJECT_ROOT)/out/logs

.PHONY: env up down run migrate seed dev test psql redis-cli clean-db

# Create .env from the example if it does not exist yet.
env:
	@test -f .env || (cp .env.example .env && echo "created .env from .env.example")

# Start PostgreSQL + PostGIS and Redis and wait until they are healthy.
up: env
	docker compose up -d --wait postgres redis

down:
	docker compose down

# Run the API on the host (migrations are applied on startup).
run:
	go run $(PROJECT_ROOT)/cmd/tailverse

migrate:
	go run $(PROJECT_ROOT)/cmd/tailverse -migrate-only

# Load sample walk spots (the API has no endpoint to create them).
seed: migrate
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $(POSTGRES_USER) -d $(POSTGRES_DB) < migrations/seed/walk_spots.sql

# Everything needed for a fresh local start.
dev: up seed run

test:
	go test ./...

psql:
	docker compose exec postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

redis-cli:
	docker compose exec redis redis-cli

# Drop all local data (postgres volume + redis state).
clean-db:
	docker compose down -v
