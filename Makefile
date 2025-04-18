.PHONY: all build fmt lint test clean format act-check act-ci act-release act-test act-build format-fix

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

# Run linting (after formatting)
lint: format
	golangci-lint run --no-config

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
	act -n -W .github/workflows/ci.yml
	@echo "Checking Release workflow syntax..."
	act -n -W .github/workflows/release.yml

# GitHub Actions - Run CI test job locally
act-test:
	@echo "Running CI test job locally..."
	act -j test -W .github/workflows/ci.yml --artifact-server-path /tmp/artifacts

# GitHub Actions - Run CI build job locally
act-build:
	@echo "Running CI build job locally..."
	act -j build -W .github/workflows/ci.yml --artifact-server-path /tmp/artifacts

# GitHub Actions - Run CI workflow locally (all jobs)
act-ci: act-test act-build
	@echo "CI workflow completed."

# GitHub Actions - Run Release workflow locally (dry-run)
act-release:
	@echo "Running Release workflow locally (dry-run)..."
	act -n -j build -W .github/workflows/release.yml 