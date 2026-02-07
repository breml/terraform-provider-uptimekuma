# Code Style & Linting

This document describes the code quality standards, linting configuration, and best practices for the
terraform-provider-uptimekuma codebase.

## Go Version & Basic Style

- **Go version**: 1.25.2
- **Import grouping**: stdlib, then third-party, then local (enforced by goimports/gci)
- **Local import prefix**: `github.com/breml/terraform-provider-uptimekuma`
- **Client alias**: Use `kuma` for `github.com/breml/go-uptime-kuma-client`
- **Terraform types**: Use `types.String`, `types.Int64`, `types.Bool`, `types.List` from terraform-plugin-framework
- **Schema patterns**: Use plan modifiers like `int64planmodifier.UseStateForUnknown()` for computed IDs
- **Defaults**: Use schema defaults like `int64default.StaticInt64()`, `booldefault.StaticBool()`, `stringdefault.StaticString()`
- **Error handling**: Add errors to `resp.Diagnostics`, not direct returns
- **Self-documenting code**: Avoid inline comments unless necessary

## Strict Linting Configuration

This project uses a comprehensive [.golangci.yml](.golangci.yml) configuration with 80+ linters enabled. Key requirements:

### Code Formatting

- **Formatters**: gofumpt (stricter than gofmt), goimports, golines, newline-after-block
- **Max line length**: 120 characters (enforced by golines)
- **Auto-fix**: `task fmt` and `task lint` both auto-fix issues when possible

### Function Complexity Limits

These limits keep functions maintainable and testable:

- **Arguments**: Max 6 parameters per function (revive:argument-limit)
- **Return values**: Max 3 return values (revive:function-result-limit)
- **Function length**: Max 50 statements OR 100 lines (revive:function-length)
- **Cognitive complexity**: Max 20 (revive:cognitive-complexity)
- **Cyclomatic complexity**: Max 30 (revive:cyclomatic)
- **Control nesting**: Max 5 levels (revive:max-control-nesting)
- **Naked returns**: Not allowed in any function (nakedret)

### Naming Conventions

#### Variables and Functions

- **Case style**: Use camelCase, no underscores except in test names
- **Test names**: Underscores allowed (e.g., `TestAcc_MonitorHTTP_Create`)

#### Errors

- **Sentinel errors**: Prefix with `Err` (e.g., `ErrNotFound`)
- **Error types**: Suffix with `Error` (e.g., `ValidationError`)

#### Import Aliases

- **Style**: Lowercase, no version numbers
- **Example**: Use `kuma` not `kuma2` for `github.com/breml/go-uptime-kuma-client`

#### Repeated Argument Types

Always use full type for each parameter (revive:enforce-repeated-arg-type-style: "full"):

```go
// Bad
func foo(a, b, c string) {}

// Good
func foo(a string, b string, c string) {}
```

#### Exported Naming

- **Documentation**: Must document all exported symbols
- **Package stuttering**: Avoid repeating package name in type names
  - Good: `monitor.HTTP` (in package `monitor`)
  - Bad: `monitor.HTTPMonitor` (in package `monitor`)

### Code Quality Requirements

#### Global State

- **No global variables**: gochecknoglobals enforces no package-level mutable state
  - Exception: Test package global variables allowed for test fixtures
  - Exception: internal/client pool uses globals with explicit `//nolint:gochecknoglobals` comment
- **No init functions**: gochecknoinits prevents `init()` functions
  - Exception: Test setup in `func TestMain(m *testing.M)` is allowed

#### Error Handling

- **Error wrapping**: All errors from external packages must be wrapped (wrapcheck)

  ```go
  // Bad
  return err

  // Good
  return fmt.Errorf("failed to create monitor: %w", err)
  ```

- **Error checking**: All errors must be checked, including type assertions (errcheck)

  ```go
  // Bad
  client, _ := req.ProviderData.(*kuma.Client)

  // Good
  client, ok := req.ProviderData.(*kuma.Client)
  if !ok {
      resp.Diagnostics.AddError("Unexpected type", "...")
      return
  }
  ```

#### Comments

- **Comments density**: Minimum 10% comment lines in functions (revive:comments-density, see [.golangci.yml](.golangci.yml))
- **Comments style**: Comments must end with a period (godot)
- **Exported symbols**: All exported functions, types, constants must have doc comments

#### Test Organization

- **Test separation**: Tests must use separate `_test` package (testpackage)

  ```go
  // File: internal/utils/helper_test.go

  // Bad
  package utils

  // Good
  package utils_test

  import "github.com/breml/terraform-provider-uptimekuma/internal/utils"
  ```

- **Exception**: [internal/provider/*_test.go](internal/provider/) tests can use same package for testing private methods

### Logging (slog)

When using structured logging with `log/slog`:

- **No global loggers**: Must not use global slog logger (sloglint:no-global)

  ```go
  // Bad
  slog.Info("message")

  // Good
  logger.InfoContext(ctx, "message")
  ```

- **Context required**: Use context-aware methods when context is in scope (sloglint:context)

  ```go
  // Bad (when ctx available)
  logger.Info("message")

  // Good
  logger.InfoContext(ctx, "message")
  ```

- **Attributes only**: Use `slog.Attr()`, not key-value pairs (sloglint:attr-only)

  ```go
  // Bad
  logger.Info("message", "key", value)

  // Good
  logger.Info("message", slog.Int("key", value))
  ```

- **Static messages**: Log messages must be string literals (sloglint:static-msg)

  ```go
  // Bad
  logger.Info(fmt.Sprintf("processed %d items", count))

  // Good
  logger.Info("processed items", slog.Int("count", count))
  ```

- **Key naming**: Use snake_case for log attribute keys (sloglint:key-naming-case)

  ```go
  // Bad
  slog.Int("monitorID", id)

  // Good
  slog.Int("monitor_id", id)
  ```

### Security & Best Practices

- **gosec**: Security vulnerability scanning enabled
  - Checks for common security issues (SQL injection, command injection, etc.)
  - Weak crypto usage detection
  - File permission issues

- **No shadowing**: Variable shadowing not allowed, strict mode (govet:shadow)
  - Especially important for `err` and `ctx` variables

- **Exhaustive switches**: All enum cases must be handled (exhaustive, gochecksumtype)

  ```go
  // Bad (missing case)
  switch status {
  case StatusUp:
      // handle
  }

  // Good
  switch status {
  case StatusUp:
      // handle
  case StatusDown:
      // handle
  default:
      // handle unknown
  }
  ```

- **Resource cleanup**: HTTP response bodies, SQL rows/statements must be closed

  ```go
  resp, err := http.Get(url)
  if err != nil {
      return err
  }
  defer resp.Body.Close()
  ```

- **No deprecated**: Use `math/rand/v2` not `math/rand`, use modern stdlib features

### Test-Specific Rules

Tests (`*_test.go`) have relaxed rules for:

- **Code duplication** (dupl): Test repetition is acceptable for clarity
- **Function complexity** (cognitive-complexity, cyclomatic, function-length): Tests can be longer
- **Security checks** (gosec, noctx): Test code security is less critical
- **Error wrapping** (wrapcheck): Test errors don't need wrapping
- **Deep exit calls**: `os.Exit` in `main_test.go` is allowed for test lifecycle management

## Git Hooks

The project uses [lefthook](https://github.com/evilmartians/lefthook) for git hooks. Configuration in [lefthook.yml](lefthook.yml).

### Pre-commit Hook

Automatically runs on every commit:

1. **Verify golangci-lint config**: Ensures `.golangci.yml` is valid
2. **Format code**: Runs `gofumpt` and `newline-after-block` formatters
3. **Lint code**: Runs `golangci-lint --fix` to auto-fix issues
4. **Lint markdown**: Runs markdown linting on documentation files

**What gets auto-fixed**:

- Code formatting (gofumpt)
- Import organization (goimports)
- Line length (golines)
- Many linter issues (golangci-lint --fix)

### Pre-push Hook

Automatically runs before pushing:

1. **Check go mod tidy**: Ensures `go.mod` and `go.sum` are up to date
2. **Check go generate**: Ensures generated code is current
3. **Run all tests**: Executes `task test` to verify nothing is broken

**Setup**: Run `task install-githooks` once to enable these hooks.

## Common Linting Issues & Solutions

### Function too complex

If you hit complexity limits (50 statements, 100 lines, cognitive complexity 20):

**Solutions**:

- Extract helper functions to break down logic
- Use early returns to reduce nesting
- Split large functions into smaller, focused ones
- For provider CRUD operations, extract common patterns into shared helpers

**Example**:

```go
// Before (too complex)
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // ... 60 lines of logic
}

// After (refactored)
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data ResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    apiObject := r.buildAPIObject(ctx, &data, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    r.createResource(ctx, apiObject, &data, &resp.Diagnostics)
}

func (r *Resource) buildAPIObject(ctx context.Context, data *ResourceModel, diags *diag.Diagnostics) *APIObject {
    // ... extraction logic
}

func (r *Resource) createResource(ctx context.Context, apiObject *APIObject, data *ResourceModel, diags *diag.Diagnostics) {
    // ... creation logic
}
```

### Too many function parameters

If a function needs more than 6 parameters:

**Solutions**:

- Group related parameters into a config struct
- Use functional options pattern
- Consider if the function is doing too much (should it be split?)

**Example**:

```go
// Before (too many parameters)
func createMonitor(name string, url string, interval int64, retries int64, timeout int64, active bool, notifications []int64) error

// After (grouped into struct)
type MonitorConfig struct {
    Name          string
    URL           string
    Interval      int64
    Retries       int64
    Timeout       int64
    Active        bool
    Notifications []int64
}

func createMonitor(config *MonitorConfig) error
```

### Error wrapping

All errors from external packages must be wrapped:

```go
// Bad
if err := r.client.CreateMonitor(ctx, monitor); err != nil {
    return err
}

// Good
if err := r.client.CreateMonitor(ctx, monitor); err != nil {
    return fmt.Errorf("failed to create monitor: %w", err)
}
```

**Benefits**:

- Adds context to error messages
- Preserves error chain for `errors.Is()` and `errors.As()`
- Makes debugging easier

### Repeated argument types

Must specify type for each parameter (revive:enforce-repeated-arg-type-style: "full"):

```go
// Bad
func buildURL(protocol, host, path string) string

// Good
func buildURL(protocol string, host string, path string) string
```

**Rationale**: Explicit types improve readability and prevent type confusion.

### Variable shadowing

Avoid shadowing variables, especially `err` and `ctx`:

```go
// Bad
if err := doSomething(); err != nil {
    if err := doOther(); err != nil {  // shadows outer err
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

**Rationale**: Shadowing makes code harder to understand and can hide bugs.

### Test package separation

Unit tests should use `_test` package suffix:

```go
// File: internal/utils/helper_test.go

// Bad
package utils

func TestHelper(t *testing.T) {
    // Can access private functions but couples tests to implementation
}

// Good
package utils_test

import (
    "testing"
    "github.com/breml/terraform-provider-uptimekuma/internal/utils"
)

func TestHelper(t *testing.T) {
    // Tests public API only, better encapsulation
}
```

**Exception**: [internal/provider/*_test.go](../internal/provider/) can use same package for testing private CRUD helpers.

## Running Linters

### Format Code

```bash
task fmt
```

Runs:

- `gofumpt -w .` - Format all Go files
- `newline-after-block` - Ensure newlines after blocks

### Lint Code

```bash
task lint
```

Runs:

- `golangci-lint run --fix ./...` - Run all linters with auto-fix
- Markdown linting

**Note**: Both `task fmt` and `task lint` auto-fix issues when possible.

### Manual Linting

```bash
# Run specific linter
golangci-lint run --disable-all --enable=errcheck ./...

# Run without auto-fix (CI mode)
golangci-lint run ./...

# Check config validity
golangci-lint linters
```

## Integration with Development Workflow

### Local Development

1. Write code normally
2. Run `task fmt` before committing (or rely on pre-commit hook)
3. Run `task lint` to check for issues
4. Run `task test` to verify tests pass
5. Commit (pre-commit hook runs automatically)
6. Push (pre-push hook runs automatically)

### Continuous Integration

CI should run:

1. `task lint` - Verify all linting passes (no --fix, strict mode)
2. `task test` - Run unit tests
3. `task testacc` - Run acceptance tests (if applicable)

### IDE Integration

**VS Code** ([.vscode/settings.json](.vscode/settings.json) recommended):

```json
{
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "go.formatTool": "gofumpt",
    "editor.formatOnSave": true
}
```

**GoLand/IntelliJ IDEA**:

- Enable golangci-lint integration in Settings → Go → Linter
- Set gofumpt as formatter in Settings → Go → Fmt

**Vim/Neovim**:

- Use [ale](https://github.com/dense-analysis/ale) or [nvim-lint](https://github.com/mfussenegger/nvim-lint)
- Configure golangci-lint as linter and gofumpt as formatter

## Disabling Linters (Use Sparingly)

Sometimes linters need to be disabled for specific lines:

```go
//nolint:lintername // Reason why this is necessary
problematicCode()

//nolint:lintername1,lintername2 // Multiple linters
problematicCode()

//nolint:all // Disable all linters (very rarely needed)
problematicCode()
```

**Guidelines**:

- Always include a comment explaining why the linter is disabled
- Be as specific as possible (use linter name, not `all`)
- Consider if the code can be refactored instead of disabling the linter
- Review nolint comments during code review

**Common exceptions**:

- `//nolint:gochecknoglobals` - For test package globals or internal/client pool
- `//nolint:gosec` - For non-security-critical use of weak crypto (tests, non-sensitive data)

## Linter Reference

### Enabled Linters (Partial List)

**Code Quality**:

- `gofumpt` - Stricter gofmt
- `goimports` - Import organization
- `golines` - Line length enforcement
- `revive` - Fast, configurable linter (many rules)
- `staticcheck` - Advanced static analysis
- `govet` - Official Go static analyzer

**Error Handling**:

- `errcheck` - Check all errors are handled
- `wrapcheck` - Errors from external packages are wrapped
- `errname` - Error naming conventions

**Performance**:

- `prealloc` - Suggest slice preallocation
- `ineffassign` - Detect ineffective assignments

**Security**:

- `gosec` - Security vulnerability scanner

**Style**:

- `gci` - Import order enforcement
- `godot` - Comments must end with period
- `nakedret` - No naked returns

**Complexity**:

- `funlen` - Function length
- `gocognit` - Cognitive complexity
- `gocyclo` - Cyclomatic complexity
- `cyclop` - Package cyclomatic complexity
- `nestif` - Nesting depth

**Tests**:

- `testpackage` - Test package separation
- `tparallel` - Parallel test detection

See [.golangci.yml](.golangci.yml) for complete list and configuration.

## Related Documentation

- [CLAUDE.md](CLAUDE.md) - Project overview and navigation
- [internal/client/CLAUDE.md](internal/client/CLAUDE.md) - Client implementation patterns
- [internal/provider/CLAUDE.md](internal/provider/CLAUDE.md) - Provider implementation patterns
- [.golangci.yml](.golangci.yml) - Complete linter configuration
- [lefthook.yml](lefthook.yml) - Git hooks configuration
- [Taskfile.yml](Taskfile.yml) - Task runner configuration
