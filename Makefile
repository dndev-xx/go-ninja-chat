BIN := "./bin/chat"
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)
SOURCE := "./cmd/chat-service"

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" $(SOURCE)

run: build
	$(BIN)

lint:
	golangci-lint run

help:
	@echo "build: build project"
	@echo "run: build and run project"
	@echo "lint: golangci-lint project"