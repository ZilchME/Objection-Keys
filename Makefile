# Makefile for objection-keys

MODULE := objection-keys
BINARY := objection-keys

.PHONY: all build clean test help

## build: Build the binary
build:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(BINARY) ./cmd/objection-keys/

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)

## test: Run tests
test:
	go test -v -race ./...

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
