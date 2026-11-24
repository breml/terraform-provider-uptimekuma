provider "uptimekuma" {
  endpoint = "http://localhost:3001"
  username = "admin"
  password = "password"
  timeout  = "30s"
}

# An account with two-factor authentication enabled: the provider derives the
# one-time code from the base32 secret of the otpauth:// URI.
provider "uptimekuma" {
  alias       = "two_factor"
  endpoint    = "http://localhost:3001"
  username    = "admin"
  password    = var.uptimekuma_password
  totp_secret = var.uptimekuma_totp_secret
}

# A session token from an earlier login, which needs no one-time code. The
# username and password are the fallback for a token the server refuses.
provider "uptimekuma" {
  alias         = "session_token"
  endpoint      = "http://localhost:3001"
  session_token = var.uptimekuma_session_token
  username      = "admin"
  password      = var.uptimekuma_password
}

# An API key of Uptime Kuma 3.0, with "API and MCP" enabled in the settings.
# The key is the only credential: no username, password or session token.
provider "uptimekuma" {
  alias    = "api_key"
  endpoint = "https://uptime-kuma.example.com"
  api_key  = var.uptimekuma_api_key
}

# Uptime Kuma behind Cloudflare Access: the service token goes in headers.
provider "uptimekuma" {
  alias    = "cloudflare_access"
  endpoint = "https://uptime-kuma.example.com"
  username = "admin"
  password = var.uptimekuma_password

  headers = {
    "CF-Access-Client-Id"     = var.cf_access_client_id
    "CF-Access-Client-Secret" = var.cf_access_client_secret
  }
}

# Ephemeral variables (Terraform 1.10+) keep the secrets out of saved plans.
variable "uptimekuma_password" {
  type      = string
  sensitive = true
  ephemeral = true
}

variable "uptimekuma_totp_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

variable "uptimekuma_session_token" {
  type      = string
  sensitive = true
  ephemeral = true
}

variable "uptimekuma_api_key" {
  type      = string
  sensitive = true
  ephemeral = true
}

variable "cf_access_client_id" {
  type      = string
  sensitive = true
  ephemeral = true
}

variable "cf_access_client_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}
