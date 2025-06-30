# MCP Server Makefile Build Configuration

BINARY_NAME := mcp-server-makefile
GO := go
GOFLAGS := -v
LDFLAGS := -s -w

.PHONY: all build clean test lint

# Build the binary
all: build

# Build the MCP server
build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

# Run tests
test:
	$(GO) test -v ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)

# Run linting
lint:
	$(GO) vet ./...
	golangci-lint run

# Install the binary
install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/