.PHONY: all build fmt lint test clean format act-check act-ci act-release act-test act-build format-fix install-lint

# Default to build
all: fmt lint test build

# Build the application
build:
	go build -o redrip ./cmd

# Format code (basic)
fmt:
	go fmt ./...

# Format code with gofmt (explicit)
format:
	gofmt -s -w .
	@if command -v goimports > /dev/null; then \
		echo "Running goimports..."; \
		goimports -w .; \
	fi

# Format code and commit changes if needed
format-fix: format
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Formatting changes detected. Committing..."; \
		git add .; \
		git commit -m "Auto-format code with go fmt"; \
		echo "Changes committed. You may need to push them."; \
	else \
		echo "No formatting changes needed."; \
	fi

# Install golangci-lint v1.54.2 (same as CI)
install-lint:
	@echo "Installing golangci-lint v1.54.2..."
	@if command -v asdf >/dev/null; then \
		echo "Using asdf to install golangci-lint..."; \
		asdf install golangci-lint 1.54.2; \
		asdf local golangci-lint 1.54.2; \
	else \
		echo "asdf not found, installing with Go..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2; \
	fi
	@echo "golangci-lint v1.54.2 installed successfully"

# Run linting (after formatting)
lint: format
	golangci-lint run --timeout=5m ./...

# Run tests
test:
	go test -v ./...

# Test with race detection
test-race:
	go test -race -v ./...

# Test with coverage
test-cover:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# Clean build artifacts
clean:
	rm -f redrip
	rm -f coverage.txt

# GitHub Actions - Check workflow syntax
act-check:
	@echo "Checking CI workflow syntax..."
	act -n -W .github/workflows/ci.yml --container-architecture linux/amd64
	@echo "Checking Release workflow syntax..."
	act -n -W .github/workflows/release.yml --container-architecture linux/amd64

# GitHub Actions - Run CI test job locally
act-test:
	@echo "Running CI test job locally..."
	act -j test -W .github/workflows/ci.yml --artifact-server-path /tmp/artifacts --container-architecture linux/amd64

# GitHub Actions - Run CI build job locally
act-build:
	@echo "Running CI build job locally..."
	act -j build -W .github/workflows/ci.yml --artifact-server-path /tmp/artifacts --container-architecture linux/amd64

# GitHub Actions - Run CI workflow locally (all jobs)
act-ci: act-test act-build
	@echo "CI workflow completed."

# GitHub Actions - Run Release workflow locally (dry-run)
act-release:
	@echo "Running Release workflow locally (dry-run)..."
	act -n -j build -W .github/workflows/release.yml --container-architecture linux/amd64 