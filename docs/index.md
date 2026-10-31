# Uptime Kuma Provider

The Uptime Kuma provider is used to interact with [Uptime Kuma](https://uptime.kuma.pet/)
resources through Terraform. The provider allows you to manage monitors and notification channels
for uptime monitoring.

## Example Usage

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

resource "uptimekuma_monitor_http" "example" {
  name     = "Example Monitor"
  url      = "https://example.com/health"
  interval = 60
  timeout  = 30
  active   = true
}
```

## Authentication

The provider uses basic authentication to connect to your Uptime Kuma instance. This is verified
to work with Uptime Kuma instances that support basic authentication via API. You need to
provide:

- `endpoint` - The URL of your Uptime Kuma instance
- `username` - Your Uptime Kuma username
- `password` - Your Uptime Kuma password

## Compatibility

This provider targets **Uptime Kuma 2.5.0 or later**. The acceptance tests run against
`louislam/uptime-kuma:2.5.5`. Against older servers some attributes behave differently:

- `interval` and `retry_interval` lost their 24 day maximum in 2.5.0. The provider no longer
  enforces an upper bound, so earlier servers reject large values at apply time.
- `screenshot_delay` on `uptimekuma_monitor_real_browser` is only echoed back since 2.5.0.
  Earlier servers cannot report drift on it or recover it on import.
- `analytics_type = "rybbit"` on `uptimekuma_status_page` is only accepted from 2.5.1. Uptime Kuma
  2.5.0 ships the server side support but still carries the older SQLite `CHECK` constraint on the
  column, so the write fails at apply time with a constraint violation.
- `system_service_name` on `uptimekuma_monitor_system_service` is validated against
  `^[a-zA-Z0-9._\-@]+$`, which matches the server side check added in 2.5.0. On Windows the
  Service Control Manager does not accept `@`, so a name containing it is accepted here but
  fails when the check runs.
- `uptimekuma_notification_openwa` needs Uptime Kuma 2.5.1 or later. Uptime Kuma 2.5.0 saves the
  notification without complaint, but fails every send with `Notification type is not supported`.
- `uptimekuma_notification_signalgrid` needs Uptime Kuma 2.5.4 or later. Earlier versions save the
  notification without complaint, but fail every send with `Notification type is not supported`.
- `uptimekuma_notification_notifyapp` needs Uptime Kuma 2.5.4 or later. Earlier versions save the
  notification without complaint, but fail every send with `Notification type is not supported`.

## Upgrading: `is_active` removed from notifications

All notification resources no longer have an `is_active` attribute. Delete it from your
configurations, along with any references to it, because Terraform rejects both. Existing state needs
no action.

Uptime Kuma does not store this value and does not check it when sending alerts, so
`is_active = false` never silenced a notification. To stop alerts, remove the notification from the
`notification_ids` of its monitors.

## Upgrading: `smtp_security` values

`smtp_security` on `uptimekuma_monitor_smtp` now accepts only the values Uptime Kuma recognises.
The removed values were stored as is and all behaved as opportunistic STARTTLS, so choose by the
behaviour you want:

| Removed value | Replacement |
| --- | --- |
| `TLS` | `secure` (SMTPS) |
| `STARTTLS` | `starttls` (use STARTTLS or fail) |
| `None` | `nostarttls` to never use STARTTLS, or leave unset to keep opportunistic STARTTLS |

`secure` and `starttls` are stricter than the removed values, which never enforced TLS. A check that
passed before can go down: use `secure` only on an implicit TLS port such as 465, and `starttls`
only on a server that offers STARTTLS.

The attribute also no longer defaults to `None`: unset now means null. A monitor that carries a
value on the server, including one created with a removed value or set in the web UI, plans
`<value> -> null` unless the attribute is configured. The same applies to the new `smtp_security`
attribute on `uptimekuma_monitor_tcp_port`, so configure it before applying to a TCP port monitor
whose TLS option was set outside Terraform. Configure `expiry_notification` on that resource too:
it defaults to `false`, so a monitor with the option enabled plans `true -> false`.

This provider does not manage the "expected TLS alert" option of a TCP port monitor. Any apply that
updates such a monitor clears the option without showing it in the plan.

## Upgrading to Uptime Kuma 2.5.0

Uptime Kuma 2.5.0 renamed eight notification type identifiers. Notifications created with an
earlier version of this provider are stored under the old identifier:

| Resource | Stored by provider <= 0.4.x | Uptime Kuma 2.5.0 |
| --- | --- | --- |
| `uptimekuma_notification_46elks` | `46elks` | `Elks` |
| `uptimekuma_notification_bark` | `bark` | `Bark` |
| `uptimekuma_notification_brevo` | `brevo` | `Brevo` |
| `uptimekuma_notification_evolution` | `EvolutionApi` | `evolution` |
| `uptimekuma_notification_nextcloudtalk` | `NextcloudTalk` | `nextcloudtalk` |
| `uptimekuma_notification_onesender` | `onesender` | `Onesender` |
| `uptimekuma_notification_pumble` | `Pumble` | `pumble` |
| `uptimekuma_notification_sevenio` | `sevenio` | `SevenIO` |

Data sources for these eight types no longer find such a notification, by ID or by name. The
matching resource still reads it, but the stored identifier is only corrected once the resource
is written again. To migrate, change any attribute of the resource and apply, or taint and
recreate it. Notifications of the other types are unaffected.

## Supported Resources

The provider supports managing the following resources:

- **Monitors** for various protocols and types (with support for tags)
- **Monitor Groups** for organizing monitors
- **Notifications** for alerting when monitors fail
- **Tags** for organizing and filtering monitors and notifications

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `endpoint` (String) Uptime Kuma endpoint. Can be set via `UPTIMEKUMA_ENDPOINT` environment variable.
- `max_retries` (Number) Maximum number of connection retry attempts (default: `3`). All retry attempts must complete within the overall `timeout` budget. Can be set via `UPTIMEKUMA_MAX_RETRIES` environment variable.
- `operation_timeout` (String) Timeout for a single Uptime Kuma command as a Go duration string (e.g. `30s`, `2m`) (default: `60s`). Bounds how long the provider waits for the server to acknowledge one create, read, update or delete and to confirm it, so an instance that accepts a command and never answers it fails the operation instead of blocking the apply indefinitely. It also bounds the login and setup performed while connecting, so during connect the effective bound is whichever of `timeout` and `operation_timeout` expires first. The bound is on the round trip and not on the whole call: a write that has to queue behind other writes to the same list waits its turn outside this budget. A value of `0s` means "not configured" and falls back to the default; the bound cannot be disabled, so set a large value instead. Can be set via `UPTIMEKUMA_OPERATION_TIMEOUT` environment variable.
- `password` (String, Sensitive) Uptime Kuma password. Can be set via `UPTIMEKUMA_PASSWORD` environment variable.
- `per_attempt_timeout` (String) Optional per-attempt connection timeout as a Go duration string (e.g. `5s`, `10s`). Caps the time spent on each individual connection attempt. The effective per-attempt timeout is the smaller of this value and the remaining `timeout` budget. When unset, each attempt may use the full remaining `timeout` budget. Can be set via `UPTIMEKUMA_PER_ATTEMPT_TIMEOUT` environment variable.
- `timeout` (String) Overall connection timeout as a Go duration string (e.g. `30s`, `2m`). Bounds the total time spent attempting to connect to Uptime Kuma, including all retry attempts and backoff. Defaults to `30s` if not specified. Can be set via `UPTIMEKUMA_TIMEOUT` environment variable.
- `username` (String) Uptime Kuma username. Can be set via `UPTIMEKUMA_USERNAME` environment variable.
