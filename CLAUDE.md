# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider for Uptime Kuma, built using the Terraform Plugin Framework (not
the older Plugin SDK). The provider allows managing Uptime Kuma monitors and notifications via
Terraform.

## Essential Commands

- **Build**: `make build` or `go build -v ./...`
- **Install locally**: `make install` (builds then installs)
- **Format**: `make fmt` (uses `gofmt -s -w -e .`)
- **Lint**: `make lint` (uses golangci-lint v2)
- **Unit tests**: `make test` (runs with 120s timeout, parallel=10)
- **Single test**: `go test -v -timeout=120s ./internal/provider -run TestName`
- **Acceptance tests**: `make testacc` (requires `TF_ACC=1`, runs with 120m timeout)
- **Generate docs**: `make generate` (uses tools in `tools/`)
- **Default target**: `make` (runs fmt, lint, install, generate)

## Architecture

### Client Dependency

- Uses `github.com/breml/go-uptime-kuma-client` as the API client
- go.mod has a replace directive pointing to `../go-uptime-kuma-client` (local development)
- Check `@.scratch/go-uptime-kuma-client` for the client source code

### Provider Structure

- **Main entry**: `main.go` - standard Terraform provider entrypoint
- **Provider core**: `internal/provider/provider.go` - defines UptimeKumaProvider with endpoint/username/password config
- **Provider type name**: `uptimekuma` (all resources prefixed with `uptimekuma_`)

### Resource Organization

Resources follow a pattern-based architecture:

1. **Notification resources** - manage notification endpoints
   - Base: `resource_notification_base.go` defines `NotificationBaseModel` and `withNotificationBaseAttributes()` helper
   - Generic: `resource_notification.go` - generic notification resource
   - Specific types: `resource_notification_ntfy.go`, `resource_notification_slack.go`, `resource_notification_teams.go`
   - Each notification type extends the base with type-specific fields

2. **Monitor resources** - manage uptime monitors
   - `resource_monitor_http.go` - HTTP/HTTPS monitoring
   - `resource_monitor_group.go` - monitor groups for organization
   - Monitors support hierarchical organization via the `parent` field (can reference a monitor group)

### Client Usage Pattern

- Provider creates a single `*kuma.Client` instance in `Configure()` using
  `context.Background()` (not Terraform's context, which cancels too early)
- Client is passed to resources via `req.ProviderData` in resource `Configure()` methods
- Resources use client methods like `CreateMonitor()`, `GetMonitorAs()`, `UpdateMonitor()`, `DeleteMonitor()`

### Testing

- Tests use `terraform-plugin-testing` framework
- Acceptance tests (`*_test.go`) create real resources via Docker containers
- Tests require Uptime Kuma instance running (typically via testcontainers)
- Note: Authentication may need to be disabled to avoid rate limits in acceptance tests

## Code Style

- **Go version**: 1.25.0
- **Import grouping**: stdlib, then third-party, then local
- **Client alias**: Use `kuma` for `github.com/breml/go-uptime-kuma-client`
- **Terraform types**: Use `types.String`, `types.Int64`, `types.Bool`, `types.List` from terraform-plugin-framework
- **Schema patterns**: Use plan modifiers like `int64planmodifier.UseStateForUnknown()` for computed IDs
- **Defaults**: Use schema defaults like `int64default.StaticInt64()`, `booldefault.StaticBool()`, `stringdefault.StaticString()`
- **Error handling**: Add errors to `resp.Diagnostics`, not direct returns
- **Self-documenting code**: Avoid inline comments unless necessary

## Special Directories

- **.scratch/**: Temporary code for testing ideas - not linted, not tested, not in git
- **.scratch/uptime-kuma/**: Copy of Uptime Kuma source code for reference
- **examples/**: Terraform examples for documentation generation
- **tools/**: Contains `tools.go` for documentation generation dependencies

## Definition of Done

Before marking a task complete:

1. Code is formatted (`make fmt`)
2. Code passes linting (`make lint`)
3. Unit tests cover new functionality
4. Both unit tests (`make test`) and acceptance tests (`make testacc`) pass
5. Documentation is updated (run `make generate` if adding/modifying resources)
