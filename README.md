# redrip

Redrip is a command-line tool for interacting with Redash queries.

## Configuration

Redrip uses a configuration file at `~/.redrip/config.conf`. The first time you run the tool, this file will be created automatically with default settings if it doesn't exist.

Configuration file format:
```
# Redash API URL (required)
redash_url = https://your-redash-url.com/api
# Redash API Key (required)
api_key = YOUR_REDASH_API_KEY
# Directory to save SQL files (optional)
sql_dir = /path/to/save/sql/files
```

Configuration options:
- `redash_url`: The URL of your Redash API (required)
- `api_key`: Your Redash API key (required)
- `sql_dir`: Directory to save SQL files (optional, defaults to current directory if not specified or directory doesn't exist)

If you run the tool without setting the required values in the config file, you'll see error messages guiding you to update the configuration.

## Usage

```
# List all queries
redrip list

# Get a specific query by ID and save as SQL file
redrip get <query_id>

# Dump all queries as SQL files
redrip dump
```

### Logging Options

Redrip provides command-line flags to control the verbosity of logging:

```
# Enable verbose logging
redrip --verbose list

# Enable debug logging (more detailed)
redrip --debug dump

# Short form flags are also available
redrip -v list
redrip -d get 123
```

## Installation

### Pre-built Binaries

You can download pre-built binaries for your platform from the [Releases](https://github.com/jasonsmithj/redrip/releases) page.

### Using Go

```
go install github.com/jasonsmithj/redrip@latest
```

## Development

### Development Environment Setup

This project uses [asdf](https://asdf-vm.com/) for managing tool versions. The required versions are specified in the `.tool-versions` file.

To set up your development environment:

1. Install asdf if you haven't already:
   ```bash
   # On macOS with Homebrew
   brew install asdf
   
   # Follow instructions to add asdf to your shell
   ```

2. Install the required plugins:
   ```bash
   asdf plugin add golang
   asdf plugin add golangci-lint
   ```

3. Install the tools at the correct versions:
   ```bash
   asdf install
   ```

4. Verify the installation:
   ```bash
   go version # Should show go1.24.2
   ```

### Running Tests

To run the tests:

```
go test ./...
```

To run tests with race detection and coverage:

```
go test -race -coverprofile=coverage.txt -covermode=atomic ./...
```

### Continuous Integration

This project uses GitHub Actions for continuous integration. The CI pipeline:

1. Runs linting with golangci-lint
2. Executes all unit tests with race detection
3. Generates code coverage reports

The GitHub Actions workflow is defined in `.github/workflows/ci.yml` and runs on:
- Push to the main branch
- Any pull request targeting the main branch

### Continuous Delivery

Releases are automated with GitHub Actions:

1. Create and push a new tag (e.g., `git tag v0.1.0 && git push origin v0.1.0`)
2. The release workflow will:
   - Build binaries for multiple platforms (Linux, macOS, Windows)
   - Create a GitHub release with the binaries attached
   - Generate release notes from merged PR titles

The CD workflow is defined in `.github/workflows/release.yml` and runs when a new tag is pushed.

### Local Development Setup

Before submitting code, you can run the same checks locally that would run in CI:

```bash
# Format code
make fmt        # Format without committing changes
make format-fix # Format and automatically commit any changes

# Install the same version of golangci-lint used in CI
make install-lint

# Run linting
make lint

# Run tests with coverage
make test-cover
```

### Testing GitHub Actions Workflows Locally

You can use [act](https://github.com/nektos/act) to test GitHub Actions workflows locally. The Makefile includes several commands to facilitate this:

```bash
# Check workflow syntax
make act-check

# Run specific CI jobs
make act-test    # Run only the test job
make act-build   # Run only the build job

# Run complete CI workflow
make act-ci

# Test Release workflow (dry run)
make act-release
```

To use these commands, you need to have `act` installed:

```bash
# Install with Homebrew (macOS)
brew install act

# Or with Go
go install github.com/nektos/act@latest
```

## License

MIT