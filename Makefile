BIN := "./bin/chat"
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)
SOURCE := "./cmd/chat-service"
COMPOSE_FILE := "./deploy/local/docker-compose.yaml"
COMPOSE_SENTRY_FILE := "./deploy/local/docker-compose.sentry.yaml"
CONTAINER_DB_NAME := local-postgres-1
GEN_TYPE := "./cmd/gen-types/"
GEN_PATH := "./internal/types/"
TYPE_LOWER := $(shell echo $(TYPE) | tr '[:upper:]' '[:lower:]')
UI_CLIENT := "./cmd/ui-client/main.go"

# ANSI color codes for better output
GREEN  := \033[32m
YELLOW := \033[33m
RED    := \033[31m
RESET  := \033[0m

build:
	@echo "$(YELLOW)Building project...$(RESET)"
	go build -o $(BIN) -ldflags "$(LDFLAGS)" $(SOURCE)
	@echo "$(GREEN)Build completed successfully.$(RESET)"

run: build
	@echo "$(YELLOW)Running project...$(RESET)"
	$(BIN)
	@echo "$(GREEN)Execution completed.$(RESET)"

run-client:
	@echo "$(YELLOW)Running ui client...$(RESET)"
	go run $(UI_CLIENT)
	@echo "$(GREEN)Execution completed.$(RESET)"

gen_types:
	@echo "$(YELLOW)Generating types for $(TYPE)...$(RESET)"
	go run $(GEN_TYPE) types $(TYPE) $(GEN_PATH)types.$(TYPE_LOWER).gen.go
	@echo "$(GREEN)Type generation completed successfully.$(RESET)"

test:
	@echo "$(YELLOW)Running unit tests...$(RESET)"
	go test ./... -v
	@echo "$(GREEN)Tests completed successfully.$(RESET)"

test-fail:
	@echo "$(YELLOW)Running unit tests...$(RESET)"
	go test ./... -v | grep -E "FAIL|--- FAIL:"
	@echo "$(GREEN)Tests completed successfully.$(RESET)"

test-it:
	@echo "$(YELLOW)Running integration tests...$(RESET)"
	go test -tags integration ./...
	@echo "$(GREEN)Tests completed successfully.$(RESET)"

lint:
	@echo "$(YELLOW)Running linter...$(RESET)"
	golangci-lint run
	@echo "$(GREEN)Linting completed successfully.$(RESET)"

tidy:
	@echo "$(YELLOW)Tidying and vendoring dependencies...$(RESET)"
	go mod tidy
	go mod vendor
	@echo "$(GREEN)Dependencies tidied and vendored successfully.$(RESET)"

gen:
	@echo "$(YELLOW)Generating code...$(RESET)"
	go generate ./...
	@echo "$(GREEN)Code generation completed successfully.$(RESET)"
# run only linux
up:
	@echo "$(YELLOW)Starting containers from $(COMPOSE_FILE) and $(COMPOSE_SENTRY_FILE)...$(RESET)"
	docker compose -f $(COMPOSE_FILE) -f $(COMPOSE_SENTRY_FILE) up -d
	@echo "$(GREEN)Containers started successfully.$(RESET)"

up-db:
	@echo "$(YELLOW)Starting containers from $(COMPOSE_FILE) ...$(RESET)"
	docker compose -f $(COMPOSE_FILE) up -d
	@echo "$(GREEN)Containers started successfully.$(RESET)"

down:
	@echo "$(YELLOW)Stopping containers from $(COMPOSE_FILE) and $(COMPOSE_SENTRY_FILE)...$(RESET)"
	docker compose -f $(COMPOSE_FILE) -f $(COMPOSE_SENTRY_FILE) down
	@echo "$(GREEN)Containers stopped successfully.$(RESET)"

db_status:
	@echo "$(YELLOW)Checking status of database container $(CONTAINER_DB_NAME)...$(RESET)"
	docker ps -a | grep $(CONTAINER_DB_NAME)
	@echo "$(GREEN)Database status checked.$(RESET)"

db_logs:
	@echo "$(YELLOW)Fetching logs for database container $(CONTAINER_DB_NAME)...$(RESET)"
	docker logs $(CONTAINER_DB_NAME)
	@echo "$(GREEN)Logs fetched successfully.$(RESET)"

db_stop:
	@echo "$(YELLOW)Stopping database container $(CONTAINER_DB_NAME)...$(RESET)"
	docker stop $(CONTAINER_DB_NAME)
	@echo "$(GREEN)Database container stopped successfully.$(RESET)"

db_clean:
	@echo "$(YELLOW)Cleaning up database container $(CONTAINER_DB_NAME)...$(RESET)"
	docker rm -f $(CONTAINER_DB_NAME)
	@echo "$(GREEN)Database container cleaned up successfully.$(RESET)"

sentry_update:
	@echo "$(YELLOW)Updating Sentry migrations...$(RESET)"
	docker compose -f $(COMPOSE_SENTRY_FILE) run --rm sentry upgrade
	@echo "$(GREEN)Sentry migrations updated successfully.$(RESET)"

help:
	@echo "$(YELLOW)Available commands:$(RESET)"
	@echo "  $(GREEN)build$(RESET): Build project"
	@echo "  $(GREEN)run$(RESET): Build and run project"
	@echo "  $(GREEN)run-client$(RESET): Run ui-client"
	@echo "  $(GREEN)test$(RESET): Run unit tests"
	@echo "  $(GREEN)test-fail$(RESET): Run only failed unit tests"
	@echo "  $(GREEN)test-it$(RESET): Run integration tests"
	@echo "  $(GREEN)lint$(RESET): Lint project using golangci-lint"
	@echo "  $(GREEN)tidy$(RESET): Tidy and vendor dependencies"
	@echo "  $(GREEN)gen$(RESET): Generate code"
	@echo "  $(GREEN)up$(RESET): Start all containers (including Sentry)"
	@echo "  $(GREEN)up-db$(RESET): Start DB containers"
	@echo "  $(GREEN)down$(RESET): Stop all containers (including Sentry)"
	@echo "  $(GREEN)db_status$(RESET): Check status of the database container"
	@echo "  $(GREEN)db_logs$(RESET): Fetch logs for the database container"
	@echo "  $(GREEN)db_stop$(RESET): Stop the database container"
	@echo "  $(GREEN)db_clean$(RESET): Clean up the database container"
	@echo "  $(GREEN)sentry_update$(RESET): Migrate local Sentry"
	@echo "  $(GREEN)gen_types$(RESET): Generate types (use TYPE=<name type> to specify type)"