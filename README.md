# redrip

Redrip is a command-line tool for interacting with Redash queries.

## Configuration

Redrip uses a configuration file at `~/.redrip/config.conf`. The first time you run the tool, this file will be created automatically with default settings if it doesn't exist.

Configuration file format:

```ini
# Default profile (used when no profile is specified)
[default]
# Redash API URL (required)
redash_url = https://your-redash-url.com/api
# Redash API Key (required)
api_key = YOUR_REDASH_API_KEY
# Directory to save SQL files (optional)
sql_dir = /path/to/save/sql/files

# Example staging profile
[profile stg]
redash_url = https://redash-staging.example.com/api
api_key = STAGING_API_KEY
sql_dir = /path/to/staging/sql/dir

# Example production profile
[profile prd]
redash_url = https://redash-production.example.com/api
api_key = PRODUCTION_API_KEY
sql_dir = /path/to/production/sql/dir
```

Configuration options:

- `redash_url`: The URL of your Redash API (required)
- `api_key`: Your Redash API key (required)
- `sql_dir`: Directory to save SQL files (optional, defaults to current directory if not specified or directory doesn't exist)

Multiple profiles allow you to work with different Redash instances. You can:

1. Use the `--profile` flag to specify a profile: `redrip --profile stg list`
2. Set the `REDRIP_PROFILE` environment variable: `export REDRIP_PROFILE=stg && redrip list`
3. If neither is specified, the `default` profile is used

If you run the tool without setting the required values in the config file, you'll see error messages guiding you to update the configuration.

## Usage

```bash
# List all queries
redrip list

# Get a specific query by ID and save as SQL file
redrip get <query_id>

# Dump all queries as SQL files
redrip dump

# Show current configuration settings
redrip config list

# Compare all local SQL files with Redash queries
redrip diff all

# Compare all local SQL files with Redash queries (JSON output)
redrip diff all --output json

# Compare a specific local SQL file with the corresponding Redash query
redrip diff query <query_id>

# Compare a specific local SQL file with the corresponding Redash query (JSON output)
redrip diff query <query_id> --output json

# Use a specific profile
redrip --profile stg list

# Or, using environment variable
export REDRIP_PROFILE=stg && redrip list
```

### Logging Options

Redrip provides command-line flags to control the verbosity of logging:

```bash
# Enable verbose logging
redrip --verbose list

# Enable debug logging (more detailed)
redrip --debug dump

# Short form flags are also available
redrip -v list
redrip -d get 123
```

### Output Formats

Several commands support different output formats:

- `list`: Supports `--output json` (default) or `--output text`
- `diff`: Supports `--output json` or `--output text` (default)

For JSON output, the diff command returns detailed information including:

- Query ID and name
- Status (MATCH, DIFFERENT, MISSING_IN_REDASH, ERROR)
- Path to local file
- Detailed differences when files don't match
- Summary statistics

## Installation

### Pre-built Binaries

You can download pre-built binaries for your platform from the [Releases](https://github.com/jasonsmithj/redrip/releases) page.

### Using Go

```bash
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

```bash
go test ./...
```

To run tests with race detection and coverage:

```bash
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
