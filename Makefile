BIN := "./bin/chat"
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)
SOURCE := "./cmd/chat-service"
COMPOSE_FILE := "./deploy/local/docker-compose.yaml"
COMPOSE_SENTRY_FILE := "./deploy/local/docker-compose.sentry.yaml"
CONTAINER_DB_NAME := local-postgres-1

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" $(SOURCE)

run: build
	$(BIN)

test:
	go test ./... -v

lint:
	golangci-lint run

tidy:
	go mod tidy
	go mod vendor

gen:
	go generate ./...

up:
	docker compose -f $(COMPOSE_FILE) up -d

up_sentry:
	docker compose -f $(COMPOSE_SENTRY_FILE) up -d

sentry_update:
	docker compose -f $(COMPOSE_SENTRY_FILE) run --rm sentry upgrade

down_sentry:
	docker compose -f $(COMPOSE_SENTRY_FILE) down

down:
	docker compose -f $(COMPOSE_FILE) down

db_status:
	docker ps -a | grep $(CONTAINER_DB_NAME)

db_logs:
	docker logs $(CONTAINER_DB_NAME)

db_stop:
	docker stop $(CONTAINER_DB_NAME)

db_clean:
	docker rm -f $(CONTAINER_DB_NAME)

help:
	@echo "build: build project"
	@echo "run: build and run project"
	@echo "test: run unit tests"
	@echo "lint: golangci-lint project"
	@echo "tidy: tidy and vendor run"
	@echo "gen: generate code"
	@echo "up: docker compose up"
	@echo "down: docker compsoe down"
	@echo "db_status: docker db_status"
	@echo "db_logs: docker db_logs"
	@echo "db_stop: docker db_stop"
	@echo "db_clean: docker clean db"
	@echo "up_sentry: docker up_sentry"
	@echo "down_sentry: docker down_sentry"
	@echo "sentry_update: migrate local sentry"
