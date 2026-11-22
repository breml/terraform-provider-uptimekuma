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

`endpoint`, the URL of your Uptime Kuma instance, is always required. The provider follows no
redirect that leaves its origin (scheme, host and port), and that includes an upgrade from `http` to
`https` on the same host: a redirect would turn the POST requests of the login and of the Socket.IO
long polling into GET requests without a body, and would carry `api_key` and `headers` to another
server. The error names the URL to use, so set `endpoint` to the URL Uptime Kuma is finally served
from. Behind an access proxy, a redirect to another host such as `<team>.cloudflareaccess.com`
usually means that the proxy rejected the request, see
[Access proxies](#access-proxies-headers-and-user_agent).

The credentials next to the endpoint take one of these shapes:

- **Nothing**, for a server with authentication disabled.
- **`username` and `password`**, the usual login. For an account with two-factor authentication
  (2FA) enabled, add `totp_secret`: the base32 `secret` from the `otpauth://` URI Uptime Kuma shows
  while 2FA is set up. The provider derives the one-time code from it whenever the server asks for
  one, so nothing has to be typed at apply time. Spaces, hyphens, lower case and missing padding are
  accepted, and the secret is inert on an account without 2FA.
- **`session_token`**, the credential Uptime Kuma hands out on a successful login and accepts in
  place of one. It needs neither the password nor a one-time code. Set it alone, or together with
  `username` and `password`: the token is tried first, and the password login is the fallback for a
  token the server refuses. The provider warns when that fallback happens, so the stored token can
  be replaced.
- **`api_key`**, on Uptime Kuma 3.0 only, see [API keys on Uptime Kuma 3.0](#api-keys-on-uptime-kuma-30).
  The key is the only credential: it cannot be combined with `username`, `password`,
  `session_token` or `totp_secret`.

`totp_secret` without `username` and `password` is rejected, and so is `username` without
`password` or the other way round. The endpoint, the credentials and the scheme can also come from
an environment variable: `UPTIMEKUMA_ENDPOINT`, `UPTIMEKUMA_USERNAME`, `UPTIMEKUMA_PASSWORD`,
`UPTIMEKUMA_TOTP_SECRET`, `UPTIMEKUMA_SESSION_TOKEN`, `UPTIMEKUMA_API_KEY` and
`UPTIMEKUMA_AUTH_SCHEME`, and `headers` from `UPTIMEKUMA_HEADERS` as a JSON object of strings. The
connection options have their own, see the schema below. For the endpoint, the credentials, the
scheme and the headers an empty attribute counts as unset, so the environment variable applies to it
too; that is what an attribute fed from a variable without a value leaves behind.

`api_key` and the login credentials (`username`, `password`, `session_token`, `totp_secret`) fall
back to the environment by group. A provider block that sets `api_key` does not read
`UPTIMEKUMA_USERNAME`, `UPTIMEKUMA_PASSWORD`, `UPTIMEKUMA_SESSION_TOKEN` or
`UPTIMEKUMA_TOTP_SECRET`, and one that sets any login credential does not read
`UPTIMEKUMA_API_KEY`. Both groups set in the configuration, or both in the environment, are a
conflict, and the error names where each value came from.

The credentials must be known when Terraform plans, because the provider logs in to plan at all. A
provider attribute that refers to a resource created in the same run is rejected.

```hcl
# Username and password, with a TOTP secret for an account with 2FA enabled.
provider "uptimekuma" {
  endpoint    = "https://uptime-kuma.example.com"
  username    = "admin"
  password    = var.uptimekuma_password
  totp_secret = var.uptimekuma_totp_secret
}

# A session token, with the password to fall back on.
provider "uptimekuma" {
  alias         = "token"
  endpoint      = "https://uptime-kuma.example.com"
  session_token = var.uptimekuma_session_token
  username      = "admin"
  password      = var.uptimekuma_password
}
```

### Session tokens on Uptime Kuma 2.x and 3.x

- On **Uptime Kuma 2.x** the session token is the JWT the server answers a login with, the `token`
  entry the web interface keeps in the browser's local storage (or session storage, when "Remember
  me" was not ticked). It does not expire: only a password change or a deactivated account
  invalidates it. The server accepts it without asking for a one-time code, even for an account
  with 2FA enabled.
- On **Uptime Kuma 3.x** the session token is the value of the `better-auth.session_token` cookie.
  It lapses after 7 days without use and on a password change, so keep `username` and `password`
  next to it for the provider to fall back on.

A session token is a bearer credential. For an account with 2FA enabled it is a *stronger*
credential than the password, because it needs no second factor, and `totp_secret` is the whole
second factor. Store both the way the password is stored.

### Rate limits and two-factor authentication

Every Terraform command that configures the provider logs in once: `terraform plan` once, and
`terraform apply` twice, because the plan and the apply run in separate provider processes. A login
the server rejected is not repeated on the next connection attempt. Uptime Kuma limits logins: 2.x
allows 20 per minute for the whole server, and a login that answers a one-time code costs two of
them; 3.x allows three sign-ins per ten seconds per client address. Many Terraform runs starting at
once can reach the limit.

Uptime Kuma 2.x also refuses a one-time code it has already accepted for the account. The two logins
of one `terraform apply` therefore often collide when they fall into the same 30 second step, and so
do parallel runs with the same `totp_secret`. The provider then retries once in the next step,
which settles two logins in one step but not more, and only when `timeout` can cover the wait of up
to 30 seconds. With `totp_secret`, set `timeout` to at least `60s`. A `session_token` next to the
password avoids the one-time code altogether, and is the credential to share between parallel runs
such as CI jobs.

```hcl
provider "uptimekuma" {
  endpoint    = "https://uptime-kuma.example.com"
  username    = "admin"
  password    = var.uptimekuma_password
  totp_secret = var.uptimekuma_totp_secret
  timeout     = "60s"
}
```

### API keys on Uptime Kuma 3.0

Uptime Kuma 3.0 accepts an API key on the Socket.IO connection, presented as
`Authorization: Bearer <key>`, once "API and MCP" (`apiEnabled`) is enabled in its settings. Create
the key under Settings > API Keys, then configure it alone:

```hcl
provider "uptimekuma" {
  endpoint = "https://uptime-kuma.example.com"
  api_key  = var.uptimekuma_api_key
}
```

- The "API and MCP" setting is marked alpha upstream, so its behaviour may still change.
- The server authenticates the connection as the user who created the key, without a second factor.
  The key is valid until it expires or is revoked: store it the way a password is stored.
- A refused key and a disabled "API and MCP" setting look the same to the provider; the error names
  both causes.
- Uptime Kuma 2.x and earlier have no API key login on the Socket.IO API, so `api_key` fails there
  with an error saying so, and `auth_scheme = "socket"` next to `api_key` is rejected.
- A server with authentication disabled ignores the key and lets every connection in.
- No session is created, so the key needs no sign-in.

### Access proxies: `headers` and `user_agent`

`headers` adds HTTP headers to every request the provider makes, the Socket.IO handshake included.
It is meant for what sits in front of Uptime Kuma, such as a Cloudflare Access service token or a
forward-auth proxy:

```hcl
provider "uptimekuma" {
  endpoint = "https://uptime-kuma.example.com"
  username = "admin"
  password = var.uptimekuma_password

  headers = {
    "CF-Access-Client-Id"     = var.cf_access_client_id
    "CF-Access-Client-Secret" = var.cf_access_client_secret
  }
}
```

- The headers are sent to the origin (scheme, host and port) of `endpoint` only, and the provider
  follows no redirect away from it, see [Authentication](#authentication). A redirect to another
  host such as `<team>.cloudflareaccess.com` usually means that the proxy rejected the request:
  check the service token in `headers`.
- The headers the client and its transport set themselves cannot be configured: `Host`, `Origin`,
  `Connection`, `Upgrade`, `Accept`, `Accept-Encoding`, `Content-Type`, `Content-Length`,
  `Transfer-Encoding`, `TE`, `Trailer` and `Sec-WebSocket-*`. Neither can an `Authorization`
  header with a `Bearer` token, which is what `api_key` is for; `Authorization: Basic` for a reverse
  proxy is accepted, but not next to `api_key`.
- `User-Agent` is rejected in `headers`; set it with `user_agent`.
- Every header needs a value, and two names that differ only in case are rejected, because HTTP
  treats them as one header.
- `headers` is sensitive, because its values are often credentials.

`user_agent` sets the User-Agent of every request. It defaults to
`terraform-provider-uptimekuma/<provider version>`.

### Authentication and `uptimekuma_settings`

Uptime Kuma asks for the account password again whenever settings are written, so
`uptimekuma_settings` needs `password` on the provider. A provider that authenticates with
`session_token` alone can manage every other resource. A provider with `api_key` has no password
either, and cannot have one next to the key, so manage `uptimekuma_settings` through a second
provider configuration (an alias) that logs in with `username` and `password`. That works with the
key in `UPTIMEKUMA_API_KEY` too, because a provider block with a login credential does not read it:

```hcl
provider "uptimekuma" {
  endpoint = "https://uptime-kuma.example.com"
  # api_key from UPTIMEKUMA_API_KEY
}

provider "uptimekuma" {
  alias    = "settings"
  endpoint = "https://uptime-kuma.example.com"
  username = "admin"
  password = var.uptimekuma_password
}

resource "uptimekuma_settings" "this" {
  provider              = uptimekuma.settings
  keep_data_period_days = 180
}
```

### Keeping credentials out of the configuration

Provider attributes cannot be write-only, but Terraform never stores the provider configuration in
the state. Pass credentials through environment variables, or through
[ephemeral input variables](https://developer.hashicorp.com/terraform/language/values/variables#exclude-values-from-state-and-plan)
(Terraform 1.10 or later), which also keeps them out of saved plan files:

```hcl
variable "uptimekuma_session_token" {
  type      = string
  sensitive = true
  ephemeral = true
}
```

### Pinning the authentication scheme

Uptime Kuma 3.0 replaces the login command on the Socket.IO connection with an HTTP sign-in. The
provider detects which one the server speaks with one request before it connects. `auth_scheme`
pins it instead (`socket` for 2.x, `http` for 3.x), which is only needed for a reverse proxy that
answers that request itself. Provider versions up to 0.6.0 took any answer to it that was not JSON
for Uptime Kuma 2.x; the provider now reports such an answer as an error, so behind such a proxy pin
`auth_scheme = "socket"` for a 2.x server. A `4xx` answer is not retried, because it is a proxy
refusing the request.

## Compatibility

This provider targets **Uptime Kuma 2.5.x and 3.0.x**. The acceptance tests run against
`louislam/uptime-kuma:2.5.9`, and against the 3.0 nightly `louislam/uptime-kuma:nightly3` until
Uptime Kuma 3.0 is released.

Uptime Kuma 3.0 replaces the socket login with an HTTP sign-in. The provider detects which one the
server speaks, so the provider configuration is the same for both. On 3.0 the following applies:

- `username` and `password` work as for 2.5.x, and so does a `session_token`; on 3.0 the token is a
  session cookie that lapses after 7 days without use, so keep `username` and `password` next to
  it, see [Session tokens on Uptime Kuma 2.x and 3.x](#session-tokens-on-uptime-kuma-2x-and-3x).
- An API key authenticates the socket connection once "API and MCP" (`apiEnabled`, alpha) is enabled
  in the settings, see [API keys on Uptime Kuma 3.0](#api-keys-on-uptime-kuma-30).
- The server allows three sign-ins per ten seconds from one address. Many provider instances
  signing in at once, such as several `terraform apply` runs in parallel, can hit that limit. See
  [Rate limits and two-factor authentication](#rate-limits-and-two-factor-authentication).
- The first user's password must have at least eight characters and is checked against the Have I
  Been Pwned database.

Against servers older than 2.5.0 some attributes behave differently:

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
- `uptimekuma_notification_amootsms` needs Uptime Kuma 2.5.4 or later. Earlier versions save the
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

- `api_key` (String, Sensitive) API key of an Uptime Kuma 3.0 server, used instead of a login. The provider presents it as `Authorization: Bearer <key>` on the Socket.IO handshake, which the server accepts only while "API and MCP" is enabled in its settings (`apiEnabled`, alpha upstream). It is the only credential of a provider that sets it, so it conflicts with `username`, `password`, `session_token` and `totp_secret`, with an `Authorization` header in `headers`, and with `auth_scheme = "socket"`: Uptime Kuma 2.x has no API key login on the Socket.IO API. Set in the configuration, it keeps the provider from reading `UPTIMEKUMA_USERNAME`, `UPTIMEKUMA_PASSWORD`, `UPTIMEKUMA_SESSION_TOKEN` and `UPTIMEKUMA_TOTP_SECRET`; a login credential set in the configuration keeps it from reading `UPTIMEKUMA_API_KEY`. A server with authentication disabled ignores the key. `uptimekuma_settings` cannot be written with it, because Uptime Kuma asks for the account password whenever settings are written. Can be set via `UPTIMEKUMA_API_KEY` environment variable, which applies when the attribute is unset or empty.
- `auth_scheme` (String) How the provider authenticates: `auto` (default) detects it with one request before connecting, `socket` is the login command on the Socket.IO connection of Uptime Kuma 2.x, and `http` is the HTTP sign-in of Uptime Kuma 3.x. The detection is right for every Uptime Kuma release; pin the scheme only for a reverse proxy that answers the detection request itself. A pinned scheme the server does not speak fails the login. Can be set via `UPTIMEKUMA_AUTH_SCHEME` environment variable, which applies when the attribute is unset or empty.
- `endpoint` (String) Uptime Kuma endpoint. The provider follows no redirect that leaves the origin (scheme, host and port) of the endpoint, an upgrade from `http` to `https` included, so set it to the URL Uptime Kuma is finally served from. Can be set via `UPTIMEKUMA_ENDPOINT` environment variable, which applies when the attribute is unset or empty.
- `headers` (Map of String, Sensitive) HTTP headers added to every request to Uptime Kuma, the Socket.IO handshake included, for an access proxy in front of it such as Cloudflare Access (`CF-Access-Client-Id` and `CF-Access-Client-Secret`). They are sent to the origin (scheme, host and port) of `endpoint` only, and the provider follows no redirect that leaves that origin, an upgrade from `http` to `https` included: set `endpoint` to the URL the redirect names instead. The headers the client and its transport set themselves (`Host`, `Origin`, `Connection`, `Upgrade`, `Accept`, `Accept-Encoding`, `Content-Type`, `Content-Length`, `Transfer-Encoding`, `TE`, `Trailer` and `Sec-WebSocket-*`) are rejected, and so is an `Authorization` header with a `Bearer` token, which belongs in `api_key`; `Authorization: Basic` for a reverse proxy is accepted, except next to `api_key`. `User-Agent` is rejected too, set it with `user_agent`. A value must not be empty, and two names that differ only in case are rejected. Can be set via `UPTIMEKUMA_HEADERS` environment variable as a JSON object of strings (e.g. `{"CF-Access-Client-Id":"..."}`), which applies when the attribute is unset or empty.
- `max_retries` (Number) Maximum number of connection retry attempts (default: `3`). All retry attempts must complete within the overall `timeout` budget. Can be set via `UPTIMEKUMA_MAX_RETRIES` environment variable.
- `operation_timeout` (String) Timeout for a single Uptime Kuma command as a Go duration string (e.g. `30s`, `2m`) (default: `60s`). Bounds how long the provider waits for the server to acknowledge one create, read, update or delete and to confirm it, so an instance that accepts a command and never answers it fails the operation instead of blocking the apply indefinitely. It also bounds the login and setup performed while connecting, so during connect the effective bound is whichever of `timeout` and `operation_timeout` expires first. The bound is on the round trip and not on the whole call: a write that has to queue behind other writes to the same list waits its turn outside this budget. A value of `0s` means "not configured" and falls back to the default; the bound cannot be disabled, so set a large value instead. Can be set via `UPTIMEKUMA_OPERATION_TIMEOUT` environment variable.
- `password` (String, Sensitive) Uptime Kuma password. Can be set via `UPTIMEKUMA_PASSWORD` environment variable, which applies when the attribute is unset or empty.
- `per_attempt_timeout` (String) Optional per-attempt connection timeout as a Go duration string (e.g. `5s`, `10s`). Caps the time spent on each individual connection attempt. The effective per-attempt timeout is the smaller of this value and the remaining `timeout` budget. When unset, each attempt may use the full remaining `timeout` budget. Can be set via `UPTIMEKUMA_PER_ATTEMPT_TIMEOUT` environment variable.
- `session_token` (String, Sensitive) Session token from an earlier login, used instead of the password. It bypasses two-factor authentication, so it needs neither the password nor a one-time code. On Uptime Kuma 2.x it is the JWT the server hands out on every login; it does not expire, and only a password change or a deactivated account invalidates it. On Uptime Kuma 3.x it is the value of the `better-auth.session_token` cookie, which lapses after 7 days without use and on a password change, so keep `username` and `password` next to it there. Set it alone, or together with `username` and `password`: the token is tried first, and the password login is the fallback for a token the server refuses, which the provider reports with a warning. `uptimekuma_settings` needs `password` on the provider, because Uptime Kuma asks for the account password again when settings are written. Can be set via `UPTIMEKUMA_SESSION_TOKEN` environment variable, which applies when the attribute is unset or empty.
- `timeout` (String) Overall connection timeout as a Go duration string (e.g. `30s`, `2m`). Bounds the total time spent attempting to connect to Uptime Kuma, including all retry attempts and backoff. Defaults to `30s` if not specified. Can be set via `UPTIMEKUMA_TIMEOUT` environment variable.
- `totp_secret` (String, Sensitive) Shared secret of an account with two-factor authentication enabled, for the password login. This is the base32 `secret` from the `otpauth://` URI Uptime Kuma shows while two-factor authentication is set up; spaces, hyphens, lower case and missing padding are accepted. The provider derives the one-time code the server asks for from it, and only once the server asks, so it is inert on an account without two-factor authentication. Requires `username` and `password`. Uptime Kuma 2.x refuses a code it has already accepted for the account, and every Terraform command logs in on its own: `terraform apply` logs in twice, for the plan and for the apply, and the second login often falls into the 30 second step of the first. The provider then retries once in the next step, so set `timeout` to at least 60s. The same applies to parallel runs, where the single retry settles two logins in one step but not more; set `session_token` next to the password to need no code. Can be set via `UPTIMEKUMA_TOTP_SECRET` environment variable, which applies when the attribute is unset or empty.
- `user_agent` (String) User-Agent of every request to Uptime Kuma. Defaults to `terraform-provider-uptimekuma/<provider version>`.
- `username` (String) Uptime Kuma username. Can be set via `UPTIMEKUMA_USERNAME` environment variable, which applies when the attribute is unset or empty.
