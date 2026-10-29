set working-directory := ".."

# Run filesystem/subprocess fixtures in an isolated container; init reaps helpers.
unit:
    docker run --rm --init --network none --mount type=bind,src="{{justfile_directory()}}",dst=/src,readonly -w /src -e GOTOOLCHAIN=local -e GOFLAGS=-buildvcs=false golang:1.24-bookworm go test -race -count=1 -timeout 120s ./...

# Formatting and static analysis, including native-test compilation.
lint:
    #!/usr/bin/env sh
    set -eu
    files=$(gofmt -l pkg examples)
    if [ -n "$files" ]; then
        printf 'Unformatted Go files:\n%s\n' "$files"
        exit 1
    fi
    go vet ./...
    go vet -tags applefm ./...
    if command -v golangci-lint >/dev/null 2>&1; then
        golangci-lint run
    fi

# Native tests only; no fake-process filesystem fixtures on the host.
native:
    go test -tags applefm -run '^TestRealFM' -count=1 -v -timeout 10m ./pkg/fm

# Release validation must not succeed through skipped native tests.
native-required:
    APPLE_FM_GO_REQUIRE_NATIVE=1 just check native

# Compile the SDK and examples on all consumer platforms without CGo.
cross-build:
    #!/usr/bin/env sh
    set -eu
    for target_os in darwin linux windows; do
        for target_arch in amd64 arm64; do
            printf 'Building %s/%s\n' "$target_os" "$target_arch"
            CGO_ENABLED=0 GOOS="$target_os" GOARCH="$target_arch" go build ./...
        done
    done
