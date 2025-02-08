BIN := "./bin/chat"
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)
SOURCE := "./cmd/chat-service"

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

help:
	@echo "build: build project"
	@echo "run: build and run project"
	@echo "test: run unit tests"
	@echo "lint: golangci-lint project"
	@echo "tidy: tidy and vendor run"
	@echo "gen: generate code"