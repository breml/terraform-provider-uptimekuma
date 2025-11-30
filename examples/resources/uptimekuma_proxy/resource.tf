# Example 1: Basic HTTP proxy without authentication
resource "uptimekuma_proxy" "http_basic" {
  host     = "proxy.example.com"
  port     = 8080
  protocol = "http"
  active   = true
}

# Example 2: HTTPS proxy with authentication
resource "uptimekuma_proxy" "https_with_auth" {
  host     = "secure-proxy.example.com"
  port     = 8443
  protocol = "https"
  auth     = true
  username = "proxyuser"
  password = "proxypassword"
  active   = true
}

# Example 3: SOCKS5 proxy as default proxy
resource "uptimekuma_proxy" "socks5_default" {
  host     = "socks.example.com"
  port     = 1080
  protocol = "socks5"
  auth     = true
  username = "socksuser"
  password = "sockspass"
  active   = true
  default  = true
}

# Example 4: Proxy with apply_existing flag to apply to all monitors on creation
resource "uptimekuma_proxy" "apply_to_monitors" {
  host           = "new-proxy.example.com"
  port           = 3128
  protocol       = "http"
  active         = true
  apply_existing = true
}

# Example 5: Inactive proxy (disabled)
resource "uptimekuma_proxy" "disabled_proxy" {
  host     = "legacy-proxy.example.com"
  port     = 8080
  protocol = "http"
  active   = false
}
