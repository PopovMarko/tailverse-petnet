include .env 
export

export PROJECT_ROOT=${shell pwd}
export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs

env-up:
	@docker compose up -d tailverse-postgres

env-down:
	@docker compose down tailverse-postgres

env-cleanup:
	@read -p "Cleanpu all volumes environment. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down tailverse-postgres tailverse-postgres-port-forwarder && \
		rm -rf out/pgdata && \
		echo "Volumes cleaned up"; \
	else \
		echo "Volumes cleanup cancelled"; \
	fi

log-cleanup:
	@read -p "CLeanup all log files. [y/N]: " ans; \
		if [ "$$ans" = "y" ]; then \
			rm -rf out/logs/* && \
			echo "Log files cleaned up"; \
		else \
			echo "Log files cleanup cancelled"; \
		fi;

port-forward:
	@docker compose up -d tailverse-postgres-port-forwarder

port-forward-close:
	@docker compose down tailverse-postgres-port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Error: No seq parameter. Usage: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm tailverse-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Error: No action provided. Usage make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm tailverse-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@tailverse-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

migrate-up:
	migrate-action action=up

migrate-down:
	migrate-action action=down

run-dev:
	@go mod tidy && \
	go run ./cmd/todo/main.go


ps:
	docker compose ps -a
