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

`endpoint`, the URL of your Uptime Kuma instance, is always required. The credentials next to it
form one of three shapes:

- **Username and password** - the usual way in. Every attribute can also come from an environment
  variable, `UPTIMEKUMA_ENDPOINT`, `UPTIMEKUMA_USERNAME` and `UPTIMEKUMA_PASSWORD`.

  ```hcl
  provider "uptimekuma" {
    endpoint = "http://localhost:3001"
    username = "admin"
    password = "password"
  }
  ```

- **Username, password and a one-time code**, for an account with two-factor authentication
  enabled. `totp_secret` is the base32 `secret` from the `otpauth://` URI Uptime Kuma shows while
  two-factor authentication is set up - the provider derives the code the server asks for from it,
  so nothing has to be typed at apply time. Spaces, hyphens, lower case and missing padding are
  accepted. The secret is inert on an account without two-factor authentication.

  ```hcl
  provider "uptimekuma" {
    endpoint    = "http://localhost:3001"
    username    = "admin"
    password    = "password"
    totp_secret = "JBSWY3DPEHPK3PXP" # or UPTIMEKUMA_TOTP_SECRET
  }
  ```

- **A session token**, the credential Uptime Kuma hands out on every successful login and accepts
  in place of one. It bypasses two-factor authentication entirely, so it needs neither the password
  nor a code.

  ```hcl
  provider "uptimekuma" {
    endpoint      = "http://localhost:3001"
    session_token = var.uptimekuma_session_token # or UPTIMEKUMA_SESSION_TOKEN
  }
  ```

Set with a username and password as well, the token is tried first and the password login is the
fallback for a token the server refuses - which is what a changed password leaves behind. The
provider warns when that fallback happens, so the stored token can be replaced.

### Obtaining a session token

Uptime Kuma answers a successful login with the token; it is not shown anywhere in the web
interface. Either read it out of the browser (the `token` entry the web interface stores for a
logged-in session), or log in once with the Go client this provider is built on and print what it
was given:

```go
client, err := kuma.New(ctx, "http://localhost:3001", "admin", "password",
    kuma.WithTOTPSecret("JBSWY3DPEHPK3PXP")) // only for a 2FA account
if err != nil {
    log.Fatal(err)
}

fmt.Println(client.SessionToken())
```

### Why a token rather than a secret for a 2FA account

Uptime Kuma records the last one-time code it accepted for an account and refuses to see it again,
and that guard is per account rather than per connection. Two Terraform runs against the same 2FA
account starting within the same 30 second step therefore cannot both log in, however correct the
secret is. A token has no such guard and can be shared by any number of runs, which makes it the
credential to use in CI.

### Storing the credentials

A session token is a bearer credential that does not expire: only a password change or a
deactivated account invalidates it, and for a 2FA account it is a *stronger* credential than the
password, because it needs no second factor. Store it - and `totp_secret`, which is the whole
second factor - the way the password is stored, and keep both out of the state file by passing them
through a variable or an environment variable.

### Authentication and `uptimekuma_settings`

Uptime Kuma asks for the account password again when settings are written, so
`uptimekuma_settings` needs `password` on the provider. A configuration authenticated only with
`session_token` can manage everything else.

### Servers with authentication disabled

A server that logs clients in by itself takes no credentials at all - leave `username`, `password`,
`totp_secret` and `session_token` unset. Note that such a provider cannot ask Uptime Kuma to resend
a list it missed, so it reports a resource it cannot find rather than removing it from state.

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
- `session_token` (String, Sensitive) Session token from an earlier login, used instead of a password. Uptime Kuma hands one out on every successful login and accepts it in place of one; it bypasses two-factor authentication entirely, which makes it the way to run several Terraform configurations against an account that has it enabled. Set it alone, or together with `username` and `password`, in which case the token is tried first and the password login is the fallback for a token the server refuses. The token does not expire: only a password change or a deactivated account invalidates it, so store it the way the password is stored. Can be set via `UPTIMEKUMA_SESSION_TOKEN` environment variable.
- `timeout` (String) Overall connection timeout as a Go duration string (e.g. `30s`, `2m`). Bounds the total time spent attempting to connect to Uptime Kuma, including all retry attempts and backoff. Defaults to `30s` if not specified. Can be set via `UPTIMEKUMA_TIMEOUT` environment variable.
- `totp_secret` (String, Sensitive) Shared secret of an account with two-factor authentication enabled. This is the base32 `secret` from the `otpauth://` URI Uptime Kuma shows while two-factor authentication is set up; the spaces, hyphens, lower case and missing padding a copied secret carries are all accepted. The provider derives the one-time code the server asks for from it, so no code has to be supplied by hand. It is only used once the server asks for a code, which makes it inert on an account without two-factor authentication. Uptime Kuma refuses a code it has already accepted and the guard is per account, so several Terraform runs starting within the same 30 second step cannot all log in - share a `session_token` instead. Can be set via `UPTIMEKUMA_TOTP_SECRET` environment variable.
- `username` (String) Uptime Kuma username. Can be set via `UPTIMEKUMA_USERNAME` environment variable.
