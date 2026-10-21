# apple-fm-go recipes

[private]
default:
    @just --list --unsorted

# Run the unit tests (no fm binary required)
test:
    go test ./...

# Run the integration test against a real fm on macOS 27
test-applefm:
    go test -tags applefm ./...

# Lint with golangci-lint when installed, otherwise go vet
lint:
    #!/usr/bin/env sh
    if command -v golangci-lint >/dev/null 2>&1; then
        golangci-lint run
    else
        go vet ./...
    fi

# Format all Go files
fmt:
    gofmt -w .
