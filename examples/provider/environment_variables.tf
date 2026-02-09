# Configure the Uptime Kuma Provider using environment variables.
# Set the following environment variables before running Terraform:
#   export UPTIMEKUMA_ENDPOINT="http://localhost:3001"
#   export UPTIMEKUMA_USERNAME="admin"
#   export UPTIMEKUMA_PASSWORD="password"
#   export UPTIMEKUMA_TIMEOUT="30s"  # Optional: connection timeout
#
# The provider block can be empty when all configuration is provided via environment variables.

provider "uptimekuma" {
}

# If you need to override environment variables for specific provider instances,
# you can still provide values in the provider block. Terraform configuration
# takes precedence over environment variables.
#
# Example: Override only the endpoint
# provider "uptimekuma" {
#   endpoint = "http://alternative-host:3001"
#   # username and password will still come from environment variables
# }
