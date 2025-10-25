.DEFAULT_GOAL := help

BIN := "./bin/chat"
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)
SOURCE := "./cmd/chat-service"
COMPOSE_FILE := "./deploy/local/docker-compose.yaml"
COMPOSE_SENTRY_FILE := "./deploy/local/docker-compose.sentry.yaml"
SWAGGER_FILE := "./deploy/local/docker-compose.swagger-ui.yaml"
CONTAINER_DB_NAME := local-postgres-1
GEN_TYPE := "./cmd/gen-types/"
GEN_PATH := "./internal/types/"
TYPE := FailedJobID JobID
TYPE_LOWER := $(shell echo $(TYPE) | tr '[:upper:]' '[:lower:]')
UI_CLIENT := "./cmd/ui-client/main.go"
SWAGGER_CONFIG := "./api/codegen.yaml"
SWAGGER_YAML := "./api/client.v1.swagger.yaml"
SWAGGER_MANAGER_YAML := "./api/manager.v1.swagger.yaml"
SWAGGER_EVENT_YAML := "./api/client.events.swagger.yaml"
GEN_TYPES_OUT = internal/server-client/v1/pkg/types.gen.go
GEN_SERVER_OUT = internal/server-client/v1/pkg/server.gen.go
GEN_CLIENT_OUT = internal/server-client/v1/pkg/client.gen.go
GEN_SPEC_OUT = internal/server-client/v1/pkg/spec.gen.go

GEN_TYPES_MANAGER_OUT = internal/server-manager/v1/pkg/types.gen.go
GEN_SERVER_MANAGER_OUT = internal/server-manager/v1/pkg/server.gen.go
GEN_MANAGER_OUT = internal/server-manager/v1/pkg/client.gen.go
GEN_SPEC_MANAGER_OUT = internal/server-manager/v1/pkg/spec.gen.go

GEN_TYPES_EVENT_OUT = internal/server-event/v1/pkg/types.gen.go
GEN_SERVER_EVENT_OUT = internal/server-event/v1/pkg/server.gen.go
GEN_EVENT_OUT = internal/server-event/v1/pkg/client.gen.go
GEN_SPEC_EVENT_OUT = internal/server-event/v1/pkg/spec.gen.go

GO_MODULE := github.com/dndev-xx/go-ninja-chat
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./internal/store/*" -not -path "*.gen.go" | tr "\n" " ")

# ANSI color codes for better output
GREEN  := \033[32m
YELLOW := \033[33m
RED    := \033[31m
RESET  := \033[0m

# Phony targets declaration
.PHONY: build run run-client gen_swagger gen_event gen_types test test-fail test-it lint tidy gen gen_ent \
        up up-db up-swagger down db_status db_logs db_stop db_clean sentry_updat format help

fmt:
	@echo "- Format"
	gofumpt -w $(GO_FILES)
	gci write -s standard -s default -s "Prefix($(GO_MODULE))" $(GO_FILES) 2> /dev/null

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

run-manager:
	@echo "$(YELLOW)Running ui client...$(RESET)"
	go run ./cmd/ui-manager/main.go
	@echo "$(GREEN)Execution completed.$(RESET)"

gen_api: gen_client gen_manager

gen_client:
	@echo "$(YELLOW)Generating client API...$(RESET)"
	@oapi-codegen \
		-package pkg \
		-generate types \
		-o $(GEN_TYPES_OUT) \
		$(SWAGGER_YAML)
	@oapi-codegen \
		-package pkg \
		-generate echo-server \
		-o $(GEN_SERVER_OUT) \
		$(SWAGGER_YAML)
	@oapi-codegen \
		-package pkg \
		-generate client \
		-o $(GEN_CLIENT_OUT) \
		$(SWAGGER_YAML)
	@oapi-codegen \
		-package pkg \
		-generate spec \
		-o $(GEN_SPEC_OUT) \
		$(SWAGGER_YAML)
	@echo "$(GREEN)Client API generation completed.$(RESET)"

gen_event:
	@echo "$(YELLOW)Generating client API...$(RESET)"
	@oapi-codegen \
		-package pkg \
		-generate types \
		-o $(GEN_TYPES_EVENT_OUT) \
		$(SWAGGER_EVENT_YAML)
	@oapi-codegen \
		-package pkg \
		-generate echo-server \
		-o $(GEN_SERVER_EVENT_OUT) \
		$(SWAGGER_EVENT_YAML)
	@oapi-codegen \
		-package pkg \
		-generate client \
		-o $(GEN_EVENT_OUT) \
		$(SWAGGER_EVENT_YAML)
	@oapi-codegen \
		-package pkg \
		-generate spec \
		-o $(GEN_SPEC_EVENT_OUT) \
		$(SWAGGER_EVENT_YAML)
	@echo "$(GREEN)Client API generation completed.$(RESET)"

gen_manager:
	@echo "$(YELLOW)Running ui client...$(RESET)"
	oapi-codegen \
      -package pkg \
      -generate types \
      -o $(GEN_TYPES_MANAGER_OUT) \
      $(SWAGGER_MANAGER_YAML)

	oapi-codegen \
      -package pkg \
      -generate echo-server \
      -o $(GEN_SERVER_MANAGER_OUT) \
      $(SWAGGER_MANAGER_YAML)

	oapi-codegen \
      -package pkg \
      -generate client \
      -o $(GEN_MANAGER_OUT) \
      $(SWAGGER_MANAGER_YAML)

	oapi-codegen \
	  -package pkg \
	  -generate spec \
      -o $(GEN_SPEC_MANAGER_OUT) \
      $(SWAGGER_MANAGER_YAML)
	@echo "$(GREEN)Execution completed.$(RESET)"

gen_all_spec: gen_client gen_event gen_manager
	@echo "$(GREEN)All code generation completed.$(RESET)"

gen_types:
	@echo "$(YELLOW)Generating types for $(TYPE)...$(RESET)"
	go run $(GEN_TYPE) types $(TYPE) $(GEN_PATH)types.$(TYPE_LOWER).gen.go
	@echo "$(GREEN)Type generation completed successfully.$(RESET)"

test:
	@echo "$(YELLOW)Running unit tests...$(RESET)"
	go test -v ./...
	@echo "$(GREEN)Tests completed successfully.$(RESET)"

test-fail:
	@echo "$(YELLOW)Running unit tests...$(RESET)"
	go test ./... -v | grep -E "FAIL|--- FAIL:"
	@echo "$(GREEN)Tests completed successfully.$(RESET)"

test-it:
	@echo "$(YELLOW)Running integration tests...$(RESET)"
	TEST_LOG_LEVEL=info \
	TEST_PSQL_ADDRESS=localhost:5433 \
	TEST_PSQL_USER=chat-service \
	TEST_PSQL_PASSWORD=chat-service \
	TEST_PSQL_DEBUG=false \
	TEST_KEYCLOAK_REALM=Bank \
	TEST_KEYCLOAK_CLIENT_ID=integration-testing \
	TEST_KEYCLOAK_CLIENT_SECRET=UMvVbOXsYhdE4IoRHOQlPHJ26l4MBLnU \
	TEST_KEYCLOAK_TEST_USER=integration-testing \
	TEST_KEYCLOAK_TEST_PASSWORD=integration-testing \
	go test -tags integration -count 1 -race ./...
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

gen_ent:
	@echo "$(YELLOW)Generating code...$(RESET)"
	ent generate ./internal/store/schema
	@echo "$(GREEN)Code generation completed successfully.$(RESET)"
# run only linux
up:
	@echo "$(YELLOW)Starting containers from $(COMPOSE_FILE) and $(COMPOSE_SENTRY_FILE)...$(RESET)"
	docker compose -f $(COMPOSE_FILE) up -d
	@echo "$(GREEN)Containers started successfully.$(RESET)"

up-db:
	@echo "$(YELLOW)Starting containers from $(COMPOSE_FILE) ...$(RESET)"
	docker compose -f $(COMPOSE_FILE) up -d
	@echo "$(GREEN)Containers started successfully.$(RESET)"

up-swagger:
	@echo "$(YELLOW)Starting containers from $(SWAGGER_FILE) ...$(RESET)"
	docker compose -f $(SWAGGER_FILE) up -d swagger-ui
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

heap:
	@echo "$(YELLOW)Updating Sentry migrations...$(RESET)"
	go tool pprof -http=:8081 http://localhost:8079/debug/pprof/heap
	@echo "$(GREEN)Sentry migrations updated successfully.$(RESET)"

goroutine:
	@echo "$(YELLOW)Updating Sentry migrations...$(RESET)"
	go tool pprof -http=:8081 http://localhost:8079/debug/pprof/goroutine
	@echo "$(GREEN)Sentry migrations updated successfully.$(RESET)"

help:
	@echo "$(YELLOW)Available commands:$(RESET)"
	@echo "  $(GREEN)build$(RESET): Build project"
	@echo "  $(GREEN)run$(RESET): Build and run project"
	@echo "  $(GREEN)run-client$(RESET): Run ui-client"
	@echo "  $(GREEN)run-manager$(RESET): Run manager-client"
	@echo "  $(GREEN)test$(RESET): Run unit tests"
	@echo "  $(GREEN)gen_event$(RESET): Run gen_event"
	@echo "  $(GREEN)gen_all_spec$(RESET): Run gen_all_spec"
	@echo "  $(GREEN)test-fail$(RESET): Run only failed unit tests"
	@echo "  $(GREEN)test-it$(RESET): Run integration tests"
	@echo "  $(GREEN)lint$(RESET): Lint project using golangci-lint"
	@echo "  $(GREEN)tidy$(RESET): Tidy and vendor dependencies"
	@echo "  $(GREEN)gen$(RESET): Generate code"
	@echo "  $(GREEN)gen_ent$(RESET): Generate code for ent"
	@echo "  $(GREEN)gen_api$(RESET): Generate code for swagger"
	@echo "  $(GREEN)up$(RESET): Start all containers (including Sentry)"
	@echo "  $(GREEN)up-db$(RESET): Start DB containers"
	@echo "  $(GREEN)up-swagger$(RESET): Start SWAGGER containers"
	@echo "  $(GREEN)down$(RESET): Stop all containers (including Sentry)"
	@echo "  $(GREEN)db_status$(RESET): Check status of the database container"
	@echo "  $(GREEN)db_logs$(RESET): Fetch logs for the database container"
	@echo "  $(GREEN)db_stop$(RESET): Stop the database container"
	@echo "  $(GREEN)db_clean$(RESET): Clean up the database container"
	@echo "  $(GREEN)sentry_update$(RESET): Migrate local Sentry"
	@echo "  $(GREEN)heap$(RESET): Start for visual profile heap"
	@echo "  $(GREEN)goroutine$(RESET): Start for visual profile goroutine"
	@echo "  $(GREEN)gen_types$(RESET): Generate types (use TYPE=<name type> to specify type)"
