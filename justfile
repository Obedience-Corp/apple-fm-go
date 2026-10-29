mod check '.justfiles/check.just'
mod release '.justfiles/release.just'

[private]
default:
    @just --list --unsorted

# Run fake-process unit tests in Docker.
test:
    just check unit

# Run native smoke tests only; unavailable prerequisites skip.
test-applefm:
    just check native

# Check formatting and vet both default and applefm-tagged packages.
lint:
    just check lint

# Format all Go files.
fmt:
    gofmt -w pkg examples
