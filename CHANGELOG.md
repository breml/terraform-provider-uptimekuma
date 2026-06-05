# Changelog

## Unreleased

BREAKING CHANGES:

- Removed the `uptimekuma_notification_linenotify` resource and data source. LINE discontinued the
  LINE Notify service and Uptime Kuma 2.3.2 removed the provider. Remove any
  `uptimekuma_notification_linenotify` resources from your Terraform configuration and state. This
  is unrelated to the still-supported `uptimekuma_notification_line` (LINE) notification.

FEATURES:

- Bumped `go-uptime-kuma-client` to v0.4.0 (Uptime Kuma 2.3.2 parity).
- Added `conditions` support to the `mqtt`, `redis`, `snmp`, `mongodb`, `mysql`, `postgres`,
  `sqlserver`, and `dns` monitors.
- Added `snmp_v3_username` to the SNMP monitor.
- Added `screenshot_delay` to the Real Browser monitor.
- Added `oauth_audience` to HTTP-based monitors (`http`, `http_keyword`, `http_json_query`).
- DNS monitors now accept multiple comma-separated resolver servers in `dns_resolve_server`.

## 0.1.0 (Unreleased)

FEATURES:

- Support for HTTP monitors (HTTP, HTTP Keyword, HTTP JSON Query)
- Support for protocol monitors (gRPC, Ping, DNS, TCP, PostgreSQL, Redis, Real Browser, Push)
- Support for monitor groups
- Support for notifications (Webhook, Slack, Teams, ntfy, Generic)
- Support for proxy configuration (HTTP, HTTPS, SOCKS5 with optional authentication)
- Support for tags
- Support for status pages and incidents
- Terraform registry documentation and examples
