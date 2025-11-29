variable "uptimekuma_endpoint" {
  description = "Uptime Kuma API endpoint"
  type        = string
  default     = "http://localhost:3001"
}

variable "uptimekuma_username" {
  description = "Uptime Kuma username"
  type        = string
  sensitive   = true
}

variable "uptimekuma_password" {
  description = "Uptime Kuma password"
  type        = string
  sensitive   = true
}

variable "slack_webhook_url" {
  description = "Slack webhook URL for notifications"
  type        = string
  sensitive   = true
}

variable "db_host" {
  description = "Database host"
  type        = string
  default     = "db.example.com"
}

variable "db_user" {
  description = "Database user"
  type        = string
  default     = "monitoring"
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}
