# Terraform Provider Uptime Kuma - Agent Guidelines

## Commands

- **Build**: `go build -v ./...` or `make build`
- **Install**: `go install -v ./...` or `make install`
- **Lint**: `golangci-lint run` or `make lint`
- **Format**: `gofmt -s -w -e .` or `make fmt`
- **Test (unit)**: `go test -v -cover -timeout=120s -parallel=10 ./...` or `make test`
- **Test (single)**: `go test -v -timeout=120s ./internal/provider -run TestName`
- **Test (acceptance)**: `TF_ACC=1 go test -v -cover -timeout 120m ./...` or `make testacc`
- **Generate docs**: `make generate` (uses tools in `tools/`)

## Architecture

- **Provider**: Terraform provider for Uptime Kuma using Plugin Framework (not SDK)
- **Main package**: `internal/provider/` contains all provider implementation
- **Resources**: `resource_notification.go`, `resource_notification_ntfy.go` (notification resources)
- **Base types**: `resource_notification_base.go` (shared notification logic)
- **Client**: Uses `github.com/breml/go-uptime-kuma-client` (local replace in go.mod, check @.scratch/go-uptime-kuma-client for source code)
- **Tests**: Acceptance tests use `terraform-plugin-testing`, create real resources via Docker
- **.scratch/uptime-kuma/**: Code of Uptime Kuma itself, copied here for reference
- **.scratch/**: Temporary code for testing ideas, not linted, not tested, not checked into git
- **admin/pep/**: Project Enhancement Proposals (PEP) for changes to the codebase

## Code Style

- **Imports**: Group stdlib, then third-party, then local; use alias `kuma` for uptime-kuma-client
- **Formatting**: Use `gofmt -s` (via `make fmt`)
- **Linters**: golangci-lint v2 with strict presets (see `.golangci.yml`)
- **Go version**: 1.25.0
- **Naming**: Follow Terraform Plugin Framework conventions; resources use `uptimekuma_` prefix
- **Error handling**: Return errors from framework methods; use diagnostics for user-facing errors
- **Types**: Use `types.String`, `types.Int64`, etc. from terraform-plugin-framework for schema
- **Documentation**: Self-documenting code, avoid inline comments
