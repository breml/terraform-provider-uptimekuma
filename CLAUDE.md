# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider for Uptime Kuma, built using the Terraform Plugin Framework (not the older
Plugin SDK). The provider enables managing Uptime Kuma resources via Terraform infrastructure-as-code.

**Key Features**:

- 85+ resource types (monitors, notifications, status pages, maintenance windows, etc.)
- 81+ corresponding data sources for querying existing resources
- Built-in retry logic and connection pooling for reliability
- Comprehensive testing with Docker-based acceptance tests
- Strict code quality standards with 80+ linters enabled

## Essential Commands

- **Install tools**: `task install` - builds all development tools to `bin/` directory
- **Install git hooks**: `task install-githooks` - sets up lefthook pre-commit and pre-push hooks
- **Build**: `task build` or `go build -v ./...`
- **Format**: `task fmt` - runs gofumpt and newline-after-block formatters
- **Lint**: `task lint` - runs markdown linting and golangci-lint with strict configuration
- **Unit tests**: `task test` - runs with 120s timeout, parallel=10, shuffled execution
- **Single test**: `go test -v -timeout=120s ./internal/provider -run TestName`
- **Acceptance tests**: `task testacc` - requires `TF_ACC=1`, runs with 480s timeout
- **Generate docs**: `task generate-docs` - generates Terraform provider documentation
- **Clean**: `task clean` - removes coverage files and build artifacts

## Quick Start

### Running Tests Locally

1. Install dependencies: `task install`
2. Install git hooks: `task install-githooks`
3. Run unit tests: `task test`
4. Run acceptance tests: `TF_ACC=1 task testacc` (requires Docker)

### Making Changes

1. Write code following patterns in [internal/provider/CLAUDE.md](internal/provider/CLAUDE.md)
2. Format: `task fmt` (auto-fixes formatting issues)
3. Lint: `task lint` (auto-fixes many linting issues)
4. Test: `task test` (ensure tests pass)
5. Commit (pre-commit hook runs automatically)
6. Push (pre-push hook runs tests automatically)

## Architecture Overview

### Project Structure

```text
terraform-provider-uptimekuma/
├── main.go                      # Provider entry point
├── CLAUDE.md                    # 📖 Documentation navigation hub
├── CODE_STYLE.md                # 📖 Code style & linting guide
├── internal/
│   ├── client/                  # Client abstraction with pooling
│   │   └── CLAUDE.md           # 📖 Client documentation
│   └── provider/                # Provider implementation (339 files)
│       └── CLAUDE.md           # 📖 Provider documentation
├── docs/                        # Generated Terraform documentation
├── examples/                    # Terraform examples
└── tools/                       # Documentation generation
```

### Key Components

- **[main.go](main.go)**: Standard Terraform provider entrypoint
- **[CODE_STYLE.md](CODE_STYLE.md)**: Code quality standards and linting guide
- **[internal/client](internal/client/)**: Client creation with retry logic and connection pooling
- **[internal/provider](internal/provider/)**: All resources, data sources, and provider configuration
- **[docs/](docs/)**: Generated Terraform provider documentation

### Resource Categories

The provider manages 85+ resource types across multiple categories:

**Monitors** (18 types): HTTP, ping, DNS, TCP, databases (PostgreSQL, MySQL, MongoDB, Redis, SQL Server),
MQTT, Docker, real browser, SNMP, push, gRPC, Steam, and monitor groups.

**Notifications** (51 types): Webhook, Slack, Teams, Discord, email (SMTP), push services (Pushover, Telegram,
Signal, etc.), SMS services, enterprise alerting (PagerDuty, Opsgenie, Splunk), regional platforms (Feishu,
DingTalk), and many more.

**Status Pages**: Public status pages with monitor groups and incidents.

**Maintenance**: Scheduled maintenance windows linked to monitors and status pages.

**Infrastructure**: Tags, proxies, and Docker host integration.

**Data Sources**: Each resource has a corresponding data source (81 total) for querying existing resources by ID or name.

### Design Patterns

- **Base models**: Common fields (MonitorBaseModel, NotificationBaseModel) embedded in specific types
- **Helper functions**: Schema builders like `withMonitorBaseAttributes()` reduce duplication
- **Composition**: Embedded structs with helper functions, not inheritance hierarchies
- **Context management**: `context.Background()` in provider, Terraform context in resources
- **Error handling**: `resp.Diagnostics` pattern with early returns, never direct returns
- **Connection pooling**: Singleton pattern for acceptance tests to prevent rate limiting

See [internal/provider/CLAUDE.md](internal/provider/CLAUDE.md) for detailed implementation patterns.

## Documentation Index

### Module Documentation

- **[internal/client/CLAUDE.md](internal/client/CLAUDE.md)** - Client package documentation
  - Client creation patterns (direct connection vs. pooled)
  - Exponential backoff retry logic
  - Connection pooling for acceptance tests
  - Integration with provider
  - Testing considerations

- **[internal/provider/CLAUDE.md](internal/provider/CLAUDE.md)** - Provider implementation
  - Provider structure and configuration
  - Complete resource catalog (85+ resources, 81+ data sources)
  - Base models and helper functions
  - CRUD implementation patterns
  - State management
  - Testing infrastructure
  - Recent changes (v0.1.6 status page perpetual diff fix)

- **[CODE_STYLE.md](CODE_STYLE.md)** - Code quality standards
  - Go version and basic style
  - 80+ linter configuration details
  - Function complexity limits
  - Naming conventions
  - Error handling requirements
  - Git hooks (pre-commit, pre-push)
  - Common linting issues and solutions

### Quick Reference

- **Base Models**: See [internal/provider/CLAUDE.md § Base Models]
  (internal/provider/CLAUDE.md#base-models-and-patterns)
- **Helper Functions**: See [internal/provider/CLAUDE.md § Helper Functions]
  (internal/provider/CLAUDE.md#helper-functions)
- **Testing Patterns**: See [internal/provider/CLAUDE.md § Testing Infrastructure]
  (internal/provider/CLAUDE.md#testing-infrastructure)
- **Client Creation**: See [internal/client/CLAUDE.md § Client Creation Patterns]
  (internal/client/CLAUDE.md#client-creation-patterns)
- **Linting Issues**: See [CODE_STYLE.md § Common Issues]
  (CODE_STYLE.md#common-linting-issues--solutions)

## Dependencies

### Runtime Dependencies

- `github.com/breml/go-uptime-kuma-client` - Uptime Kuma API client (Socket.IO-based)
- `github.com/hashicorp/terraform-plugin-framework` - Terraform Plugin Framework (v6)
- `github.com/hashicorp/terraform-plugin-log/tflog` - Structured logging

**Note**: go.mod has a replace directive pointing to `../go-uptime-kuma-client` for local development.
Check `@.scratch/go-uptime-kuma-client` for the client source code.

### Development Dependencies

- `github.com/hashicorp/terraform-plugin-testing` - Acceptance testing framework
- `github.com/ory/dockertest/v3` - Docker container management for tests
- 80+ linters via golangci-lint (see [CODE_STYLE.md](CODE_STYLE.md))

## Code Quality Standards

This project maintains strict code quality standards:

- **Function complexity**: Max 50 statements OR 100 lines per function
- **Function parameters**: Max 6 parameters (use structs for more)
- **Error handling**: All errors checked and wrapped with context
- **No global state**: No global variables or init functions (except tests)
- **Test separation**: Unit tests use `_test` package (except internal/provider)
- **Documentation**: All exported symbols documented

See [CODE_STYLE.md](CODE_STYLE.md) for complete guidelines.

## Special Directories

- **.scratch/**: Temporary code for testing ideas - not linted, not tested, not in git
- **.scratch/uptime-kuma/**: Copy of Uptime Kuma source code for reference
- **examples/**: Terraform examples for documentation generation
- **tools/**: Contains `tools.go` for documentation generation dependencies

## Definition of Done

Before marking a task complete:

1. **Code is formatted**: Run `task fmt` - must pass with no changes
2. **Code passes linting**: Run `task lint` - must pass all 80+ linters
   - Pay special attention to function complexity limits
   - Ensure all errors are checked and wrapped appropriately
   - Verify all exported symbols have proper documentation
   - Confirm no global variables or init functions were added
3. **Unit tests cover new functionality**: Add comprehensive test coverage
   - Use separate `_test` package for unit tests (exception: provider tests)
   - Tests should be clear and focused on specific behaviors
4. **All tests pass**: Both `task test` and `task testacc` must pass
   - Unit tests run with shuffle and race detection
   - Acceptance tests require real Uptime Kuma instance
5. **Documentation is updated**: Run `task generate-docs` if adding/modifying resources
   - Ensure examples are provided in `examples/` directory
   - Verify generated documentation in `docs/` is correct
6. **Git hooks pass**: Pre-commit hooks must pass before committing
   - Run `task install-githooks` if not already set up
   - Pre-push hooks will run tests automatically
