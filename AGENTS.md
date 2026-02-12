# Terraform Provider Uptime Kuma - Agent Guidelines

## Commands

- **Install tools**: `task install` - builds all development tools to `bin/` directory
- **Install git hooks**: `task install-githooks` - sets up lefthook pre-commit and pre-push hooks
- **Build**: `go build -v ./...` or `task build`
- **Format**: `task fmt` - runs gofumpt, goimports, golines, and newline-after-block formatters
- **Lint**: `golangci-lint run` or `task lint` - runs 80+ linters with strict configuration
- **Test (unit)**: `go test -v -cover -timeout=120s -parallel=10 ./...` or `task test`
- **Test (single)**: `go test -v -timeout=120s ./internal/provider -run TestName`
- **Test (acceptance)**: `TF_ACC=1 go test -v -cover -timeout 480s ./...` or `task testacc`
- **Generate docs**: `task generate-docs` - generates Terraform provider documentation
- **Clean**: `task clean` - removes coverage files and build artifacts

## Architecture

### Project Structure

- **Provider**: Terraform provider for Uptime Kuma using Plugin Framework (not SDK)
- **Main entry**: `main.go` - standard Terraform provider entrypoint
- **Provider core**: `internal/provider/provider.go` - defines UptimeKumaProvider with endpoint/username/password config
- **Provider type name**: `uptimekuma` (all resources prefixed with `uptimekuma_`)
- **Main package**: `internal/provider/` contains all provider implementation

### Resource Organization

Resources follow a pattern-based architecture:

1. **Notification resources** - manage notification endpoints
   - Base: `resource_notification_base.go` defines `NotificationBaseModel` and helper functions
   - Generic: `resource_notification.go` - generic notification resource
   - Specific types: `resource_notification_ntfy.go`, `resource_notification_slack.go`, `resource_notification_teams.go`
   - Each type extends the base with type-specific fields

2. **Monitor resources** - manage uptime monitors
   - `resource_monitor_http.go` - HTTP/HTTPS monitoring
   - `resource_monitor_group.go` - monitor groups for organization
   - Monitors support hierarchical organization via the `parent` field (can reference a monitor group)
   - **Base types**: `resource_monitor_base.go`, `resource_monitor_http_base.go` (shared logic)
   - Supports: HTTP, DNS, TCP, Postgres, Redis, gRPC, Push, Ping, Real Browser

### Client Dependency

- Uses `github.com/breml/go-uptime-kuma-client` as the API client
- go.mod has a replace directive pointing to `../go-uptime-kuma-client` (local development)
- Check `.scratch/go-uptime-kuma-client` for the client source code
- Provider creates a single `*kuma.Client` instance in `Configure()` using `context.Background()` (not Terraform's context)
- Resources use client methods like `CreateMonitor()`, `GetMonitorAs()`, `UpdateMonitor()`, `DeleteMonitor()`

### Testing

- Tests use `terraform-plugin-testing` framework
- Acceptance tests (`*_test.go`) create real resources via Docker containers
- Tests require Uptime Kuma instance running (typically via testcontainers)

### Special Directories

- **.scratch/**: Temporary code for testing ideas - not linted, not tested, not in git
- **examples/**: Terraform examples for documentation generation
- **tools/**: Contains `tools.go` for documentation generation dependencies

## Code Style

### Go Version & Basic Style

- **Go version**: 1.25.2
- **Import grouping**: stdlib, then third-party, then local (enforced by goimports/gci)
- **Local import prefix**: `github.com/breml/terraform-provider-uptimekuma`
- **Client alias**: Use `kuma` for `github.com/breml/go-uptime-kuma-client`
- **Terraform types**: Use `types.String`, `types.Int64`, `types.Bool`, `types.List` from terraform-plugin-framework
- **Schema patterns**: Use plan modifiers like `int64planmodifier.UseStateForUnknown()` for computed IDs
- **Defaults**: Use schema defaults like `int64default.StaticInt64()`, `booldefault.StaticBool()`, `stringdefault.StaticString()`
- **Error handling**: Add errors to `resp.Diagnostics`, not direct returns
- **Self-documenting code**: Avoid inline comments unless necessary

### Strict Linting Configuration

This project uses a comprehensive `.golangci.yml` configuration with 80+ linters enabled.

#### Code Formatting

- **Formatters**: gofumpt (stricter than gofmt), goimports, golines, newline-after-block
- **Max line length**: 120 characters (enforced by golines)
- **Auto-fix**: `task fmt` and `task lint` both auto-fix issues when possible

#### Function Complexity Limits

- **Arguments**: Max 6 parameters per function (revive:argument-limit)
- **Return values**: Max 3 return values (revive:function-result-limit)
- **Function length**: Max 50 statements OR 100 lines (revive:function-length)
- **Cognitive complexity**: Max 20 (revive:cognitive-complexity)
- **Cyclomatic complexity**: Max 30 (revive:cyclomatic)
- **Control nesting**: Max 5 levels (revive:max-control-nesting)
- **Naked returns**: Not allowed in any function (nakedret)

#### Naming Conventions

- **Variables/functions**: Use camelCase, no underscores except in test names
- **Errors**: Prefix sentinel errors with `Err`, suffix error types with `Error`
- **Import aliases**: Lowercase, no version numbers (e.g., use `kuma` not `kuma2`)
- **Repeated arg types**: Always use full type for each parameter (revive:enforce-repeated-arg-type-style: "full")
  - Good: `func foo(a int, b int, c int)`
  - Bad: `func foo(a, b, c int)`
- **Exported naming**: Must document all exported symbols, avoid stuttering package names
  - Good: `monitor.HTTP` not `monitor.HTTPMonitor` in package `monitor`

#### Code Quality Requirements

- **No global variables**: gochecknoglobals enforces no package-level mutable state
- **No init functions**: gochecknoinits prevents init() functions
- **No variable shadowing**: Strict mode (govet:shadow) - avoid shadowing variables, especially `err` and `ctx`
- **Error wrapping**: All errors from external packages must be wrapped (wrapcheck)
  - Good: `return fmt.Errorf("failed to create monitor: %w", err)`
  - Bad: `return err`
- **Error checking**: All errors must be checked, including type assertions (errcheck)
- **Comments density**: Min 15% comment lines in functions (revive:comments-density)
- **Comments style**: Comments must end with a period (godot)
- **Test separation**: Tests must use separate `_test` package (testpackage)
  - Exception: `internal/provider` tests can be in same package for testing private methods

#### Logging (slog)

- **No global loggers**: Must not use global slog logger (sloglint:no-global)
- **Context required**: Use context-aware methods when context is in scope (sloglint:context)
- **Attributes only**: Use `slog.Attr()`, not key-value pairs (sloglint:attr-only)
- **Static messages**: Log messages must be string literals (sloglint:static-msg)
- **Key naming**: Use snake_case for log attribute keys (sloglint:key-naming-case)

#### Security & Best Practices

- **gosec**: Security vulnerability scanning enabled
- **Exhaustive switches**: All enum cases must be handled (exhaustive, gochecksumtype)
- **Resource cleanup**: HTTP response bodies, SQL rows/statements must be closed
- **No deprecated**: Use math/rand/v2 not math/rand, use modern stdlib features

#### Test-Specific Rules

Tests (`*_test.go`) have relaxed rules for:

- Code duplication (dupl)
- Function complexity (cognitive-complexity, cyclomatic, function-length)
- Security checks (gosec, noctx)
- Error wrapping (wrapcheck)

### Git Hooks

The project uses lefthook for git hooks:

#### Pre-commit

Automatically runs on every commit:

- Verifies golangci-lint config
- Runs gofumpt, newline-after-block, and golangci-lint --fix
- Lints markdown files

#### Pre-push

Automatically runs before pushing:

- Checks `go mod tidy` is up to date
- Checks `go generate` is up to date
- Runs all tests

**Setup**: Run `task install-githooks` once to enable these hooks.

### Common Linting Issues & Solutions

#### Function too complex

If you hit complexity limits, consider:

- Extracting helper functions to break down logic
- Using early returns to reduce nesting
- Splitting large functions into smaller, focused ones
- For provider CRUD operations, extract common patterns into shared helpers

#### Too many function parameters

If a function needs more than 6 parameters:

- Group related parameters into a config struct
- Use functional options pattern
- Consider if the function is doing too much

#### Error wrapping

All errors from external packages must be wrapped:

```go
// Bad
return err

// Good
return fmt.Errorf("failed to create monitor: %w", err)
```

#### Repeated argument types

Must specify type for each parameter:

```go
// Bad
func foo(a, b, c string) {}

// Good
func foo(a string, b string, c string) {}
```

#### Variable shadowing

Avoid shadowing variables, especially `err` and `ctx`:

```go
// Bad
if err := doSomething(); err != nil {
    if err := doOther(); err != nil { // shadows outer err
        return err
    }
}

// Good
if err := doSomething(); err != nil {
    return fmt.Errorf("do something: %w", err)
}

if err := doOther(); err != nil {
    return fmt.Errorf("do other: %w", err)
}
```

#### Test package separation

Unit tests should use `_test` package suffix:

```go
// File: internal/utils/helper_test.go

// Bad
package utils

// Good
package utils_test

import "github.com/breml/terraform-provider-uptimekuma/internal/utils"
```

Exception: `internal/provider/*_test.go` can use same package for testing private methods.

## Known Pitfalls

### Status Page `public_group_list` State Management

The `public_group_list` in `uptimekuma_status_page` requires special handling because:

1. **`GetStatusPage` does not return `public_group_list`** — the API endpoint omits it.
   The `Read()` function preserves whatever is in state without modification.
2. **`SaveStatusPage` response is lossy** — the Uptime Kuma server omits optional
   fields (e.g. `sendUrl` when `false`) in the response. If state is rebuilt from
   the server response, values like `send_url = false` become `null` in state,
   causing a perpetual diff.
3. **Solution**: After `SaveStatusPage`, use `mergeGroupIDsIntoPlan()` to preserve
   all plan values and only inject computed group IDs from the server response.
   Never replace the entire `public_group_list` with data from the server response.

**Pattern**: For nested computed+optional attributes where the API response may
omit default/falsy values, always treat the Terraform plan as the source of truth
and only merge in server-assigned computed values (like IDs).

## Definition of Done

The following criteria must be met before a task is considered done:

1. **Code is formatted**: Run `task fmt` - must pass with no changes
2. **Code passes linting**: Run `task lint` - must pass all 80+ linters
   - Pay special attention to function complexity limits
   - Ensure all errors are checked and wrapped appropriately
   - Verify all exported symbols have proper documentation
   - Confirm no global variables or init functions were added
   - Verify no variable shadowing
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
