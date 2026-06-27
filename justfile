set shell := ["bash", "-uc"]

# Keep tool caches out of the way and overridable in CI sandboxes.
export GOCACHE := env_var_or_default("GOCACHE", "/tmp/gocache")
export GOMODCACHE := env_var_or_default("GOMODCACHE", "/tmp/gomodcache")
export GOLANGCI_LINT_CACHE := env_var_or_default("GOLANGCI_LINT_CACHE", "/tmp/golangci-lint-cache")

# Default recipe - show available commands
default:
    @just --list

# Build the package
build:
    go build ./...

# Run all tests (includes the runnable doc examples)
test:
    go test ./...

# Run tests with the race detector
test-race:
    go test -race ./...

# Run tests with a coverage profile + HTML report
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
    go test -run='^$' -bench=. -benchmem ./...

# go vet
vet:
    go vet ./...

# Run linters (golangci-lint v2; config in .golangci.yml)
lint:
    golangci-lint run ./...

# Run linters with auto-fix
lint-fix:
    golangci-lint run --fix ./...

# Format the code (gofumpt via golangci-lint formatters)
fmt:
    golangci-lint fmt

# Verify formatting without writing changes (CI gate)
check-formatted:
    golangci-lint fmt --diff

# Ensure go.mod/go.sum are tidy (go.sum may be absent: zero-dependency module)
check-tidy:
    go mod tidy
    git diff --exit-code -- go.mod $(git ls-files go.sum)

# All CI checks
ci: check-formatted vet test lint check-tidy

# Build the Qhull ground-truth oracle tools from the local (gitignored)
# third_party/ source. Requires the Qhull 8.0.2 source under
# third_party/qhull-8.0.2/ — see THIRD_PARTY.md / PLAN.md. Used only to
# regenerate the committed testdata/ fixtures; not needed to build or test.
oracle-build:
    @test -d third_party/qhull-8.0.2/src/libqhull_r || { echo "third_party/qhull-8.0.2 not present; see THIRD_PARTY.md"; exit 1; }
    mkdir -p bin
    for tool in introspect dump_state stepdump; do \
        cc -O2 -I third_party/qhull-8.0.2/src \
            third_party/qhull-8.0.2/$tool.c \
            third_party/qhull-8.0.2/src/libqhull_r/*.c \
            -lm -o bin/$tool; \
    done

# Remove build/coverage artifacts
clean:
    rm -rf bin coverage.out coverage.html

# Auto-fix lint + formatting
fix:
    just lint-fix
    just fmt
