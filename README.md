# Terraform Provider for Uptime Kuma

A Terraform provider for managing Uptime Kuma monitors and notifications through infrastructure-as-code.

## Features

- Manage HTTP/HTTPS monitors with advanced options (auth, TLS, redirects, body/header validation)
- Create monitors for various protocols: DNS, gRPC, TCP, Push, Ping, PostgreSQL, Redis, Real Browser
- Monitor groups for organizing related monitors
- Manage notification channels (webhook, Slack, Teams, ntfy)
- Configure generic notifications with JSON config for custom types
- Tag monitors and notifications for organization and filtering

## Dependencies

This provider uses the [go-uptime-kuma-client](https://github.com/breml/go-uptime-kuma-client) to interact with Uptime Kuma. The capabilities are limited to the features supported by the client library. If you need a feature not yet available, first check if it's supported in the client library.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (for development)
- Uptime Kuma instance with API access
- [Docker](https://www.docker.com/) (for running integration tests)

## Installation

The provider is available on the [Terraform Registry](https://registry.terraform.io/providers/breml/uptimekuma). Configure it in your Terraform code:

```hcl
terraform {
  required_providers {
    uptimekuma = {
      source  = "breml/uptimekuma"
      version = "~> 0.1"
    }
  }
}

provider "uptimekuma" {
  endpoint = "http://localhost:3001"
  username = "admin"
  password = "password"
}
```

## Quick Start

### Create an HTTP Monitor

```hcl
resource "uptimekuma_monitor_http" "example" {
  name     = "Example API"
  url      = "https://api.example.com/health"
  interval = 60
  timeout  = 30
  active   = true
}
```

### Create a Webhook Notification

```hcl
resource "uptimekuma_notification_webhook" "example" {
  name     = "Slack Webhook"
  url      = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
  method   = "POST"
  is_active = true
}
```

## Supported Resources

### Monitors
- `uptimekuma_monitor_http` - HTTP/HTTPS monitoring
- `uptimekuma_monitor_http_keyword` - HTTP monitoring with keyword detection
- `uptimekuma_monitor_http_json_query` - HTTP monitoring with JSON query validation
- `uptimekuma_monitor_grpc_keyword` - gRPC monitoring with keyword detection
- `uptimekuma_monitor_ping` - ICMP ping monitoring
- `uptimekuma_monitor_dns` - DNS query monitoring
- `uptimekuma_monitor_tcp_port` - TCP port connectivity
- `uptimekuma_monitor_push` - Push monitoring (external push events)
- `uptimekuma_monitor_postgres` - PostgreSQL database monitoring
- `uptimekuma_monitor_redis` - Redis database monitoring
- `uptimekuma_monitor_real_browser` - Browser-based monitoring
- `uptimekuma_monitor_group` - Monitor groups for organization

### Notifications
- `uptimekuma_notification` - Generic notification with JSON config
- `uptimekuma_notification_webhook` - Webhook notifications
- `uptimekuma_notification_slack` - Slack integration
- `uptimekuma_notification_teams` - Microsoft Teams integration
- `uptimekuma_notification_ntfy` - ntfy.sh notifications

## Documentation

Full documentation including all resource attributes and examples is available on the [Terraform Registry](https://registry.terraform.io/providers/breml/uptimekuma/latest/docs).

## Development

### Building

```bash
make build
```

### Running Tests

```bash
# Unit tests
make test

# Acceptance tests (requires Uptime Kuma instance)
make testacc
```

### Generating Documentation

```bash
make generate
```

### Code Quality

```bash
make fmt      # Format code
make lint     # Run linters
```

## Contributing

Contributions are welcome! Please ensure:
- Code is formatted with `make fmt`
- Code passes linting with `make lint`
- Tests pass with `make test` and `make testacc`
- Documentation is updated with `make generate`

## License

This provider is licensed under the Mozilla Public License Version 2.0. See the `LICENSE` file for details.

## Support

For issues, feature requests, or questions, please use the [GitHub repository](https://github.com/breml/terraform-provider-uptimekuma).
