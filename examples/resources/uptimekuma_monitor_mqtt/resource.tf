# MQTT monitor resource for basic connectivity checking.
# This monitor connects to an MQTT broker and checks for messages on a specific topic.
resource "uptimekuma_monitor_mqtt" "home_automation" {
  name            = "Home Automation MQTT Broker"
  description     = "Monitor MQTT broker for home automation system"
  hostname        = "mqtt.example.com"
  port            = 1883
  mqtt_topic      = "home/status"
  mqtt_check_type = "keyword"
  interval        = 60
  timeout         = 30
}

# MQTT monitor with authentication.
# Uses username and password credentials to connect to the MQTT broker.
resource "uptimekuma_monitor_mqtt" "secure_broker" {
  name                 = "Secure MQTT Broker"
  hostname             = "mqtt.internal.example.com"
  port                 = 8883
  mqtt_topic           = "sensors/temperature"
  mqtt_username        = "monitor_user"
  mqtt_password        = var.mqtt_password # Recommend using Terraform variables for sensitive data
  mqtt_check_type      = "keyword"
  mqtt_success_message = "online"
  interval             = 30
  max_retries          = 3
  notification_ids     = [1, 2] # Notification channel IDs
}

# MQTT monitor with WebSocket connection.
# Uses WebSocket protocol for MQTT communication over HTTP/HTTPS.
resource "uptimekuma_monitor_mqtt" "websocket_broker" {
  name                 = "WebSocket MQTT Monitor"
  hostname             = "mqtt.example.com"
  mqtt_topic           = "events/updates"
  mqtt_websocket_path  = "/mqtt"
  mqtt_check_type      = "keyword"
  mqtt_success_message = "received"
  interval             = 60
}

# MQTT monitor with JSON query check.
# Parses MQTT messages as JSON and validates specific values.
resource "uptimekuma_monitor_mqtt" "json_monitor" {
  name            = "JSON MQTT Monitor"
  hostname        = "mqtt.example.com"
  port            = 1883
  mqtt_topic      = "sensors/data"
  mqtt_username   = "user"
  mqtt_password   = var.mqtt_password
  mqtt_check_type = "json-query"
  json_path       = "$.status"
  expected_value  = "active"
  interval        = 60
  timeout         = 30
}

# MQTT monitor as part of a monitor group.
# Organizes related monitors under a parent group.
resource "uptimekuma_monitor_mqtt" "grouped_monitor" {
  name       = "Grouped MQTT Monitor"
  hostname   = "mqtt.example.com"
  mqtt_topic = "group/topic"
  parent     = uptimekuma_monitor_group.mqtt_monitors.id
  interval   = 60
  active     = true
}

resource "uptimekuma_monitor_group" "mqtt_monitors" {
  name     = "MQTT Monitors"
  interval = 60
}
