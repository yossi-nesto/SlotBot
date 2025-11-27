# Binary name
BINARY_NAME=slotbot
BUILD_DIR=bin

.PHONY: all build run test clean help

all: build

## Build the application
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

## Run the application
run:
	@echo "Running..."
	@go run ./cmd/server

## Run tests
test:
	@echo "Testing..."
	@go test -v ./...

## Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

## Show help
help:
	@echo "Available targets:"
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  %-15s %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
