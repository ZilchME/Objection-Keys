# Makefile for objection-keys

MODULE := objection-keys
BINARY := objection-keys
APP := Objection Keys.app

.PHONY: all build build-app clean test help

## build: Build the binary
build:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(BINARY) ./cmd/objection-keys/

## build-app: Build the macOS menu bar app bundle
build-app: build
	mkdir -p "$(APP)/Contents/MacOS" "$(APP)/Contents/Resources"
	cp "$(BINARY)" "$(APP)/Contents/MacOS/$(BINARY)"
	cp build/macos/Info.plist "$(APP)/Contents/Info.plist"
	rm -rf "$(APP)/Contents/Resources/sounds"
	cp -R sounds "$(APP)/Contents/Resources/sounds"

## clean: Remove build artifacts
clean:
	rm -rf $(BINARY) "$(APP)"

## test: Run tests
test:
	go test -v -race ./...

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
