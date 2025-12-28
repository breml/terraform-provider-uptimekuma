package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationWebhookResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWebhook")
	nameUpdated := acctest.RandomWithPrefix("NotificationWebhookUpdated")
	webhookURL := "https://example.com/webhook"
	webhookURLUpdated := "https://example.com/webhook-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWebhookResourceConfig(name, webhookURL, "json", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("webhook_content_type"),
						knownvalue.StringExact("json"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationWebhookResourceConfig(nameUpdated, webhookURLUpdated, "form-data", ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("webhook_content_type"),
						knownvalue.StringExact("form-data"),
					),
				},
			},
		},
	})
}

func TestAccNotificationWebhookResource_WithHeaders(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWebhookHeaders")
	webhookURL := "https://api.example.com/notify"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWebhookResourceConfigWithHeaders(name, webhookURL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("webhook_content_type"),
						knownvalue.StringExact("json"),
					),
				},
			},
		},
	})
}

func TestAccNotificationWebhookResource_CustomBody(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWebhookCustom")
	webhookURL := "https://api.example.com/alerts"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWebhookResourceConfigWithCustomBody(name, webhookURL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webhook.test",
						tfjsonpath.New("webhook_content_type"),
						knownvalue.StringExact("custom"),
					),
				},
			},
		},
	})
}

func testAccNotificationWebhookResourceConfig(
	name string,
	webhookURL string,
	contentType string,
	customBody string,
) string {
	config := fmt.Sprintf(`
resource "uptimekuma_notification_webhook" "test" {
  name                 = %[1]q
  webhook_url          = %[2]q
  webhook_content_type = %[3]q
  is_active            = true
}
`, name, webhookURL, contentType)

	if customBody != "" {
		config = fmt.Sprintf(`
resource "uptimekuma_notification_webhook" "test" {
  name                 = %[1]q
  webhook_url          = %[2]q
  webhook_content_type = %[3]q
  webhook_custom_body  = %[4]q
  is_active            = true
}
`, name, webhookURL, contentType, customBody)
	}

	return providerConfig() + config
}

func testAccNotificationWebhookResourceConfigWithHeaders(name string, webhookURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_webhook" "test" {
  name                   = %[1]q
  webhook_url            = %[2]q
  webhook_content_type   = "json"
  webhook_additional_headers = {
    "Authorization" = "Bearer secret-token"
    "X-App-ID"      = "uptime-kuma"
  }
  is_active = true
}
`, name, webhookURL)
}

func testAccNotificationWebhookResourceConfigWithCustomBody(name string, webhookURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_webhook" "test" {
  name                 = %[1]q
  webhook_url          = %[2]q
  webhook_content_type = "custom"
  webhook_custom_body  = jsonencode({
    "title"   = "Alert - $${monitorJSON['name']}"
    "message" = "$${msg}"
  })
  is_active = true
}
`, name, webhookURL)
}
