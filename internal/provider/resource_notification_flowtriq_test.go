package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccNotificationFlowtriqResource covers the minimal shape, where api_key stays unset and must
// therefore round-trip as null rather than as the empty string.
func TestAccNotificationFlowtriqResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationFlowtriq")
	nameUpdated := acctest.RandomWithPrefix("NotificationFlowtriqUpdated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationFlowtriqResourceConfig(
					name,
					"https://app.flowtriq.com/api/webhooks/0123456789abcdef",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact("https://app.flowtriq.com/api/webhooks/0123456789abcdef"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.test",
						tfjsonpath.New("api_key"),
						knownvalue.Null(),
					),
				},
			},
			{
				Config: testAccNotificationFlowtriqResourceConfig(
					nameUpdated,
					"https://app.flowtriq.com/api/webhooks/fedcba9876543210",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact("https://app.flowtriq.com/api/webhooks/fedcba9876543210"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.test",
						tfjsonpath.New("api_key"),
						knownvalue.Null(),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_flowtriq.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationFlowtriqResourceAPIKey covers the authenticated shape, including clearing the
// API key again, which has to fall back to null so Uptime Kuma omits the X-API-Key header.
func TestAccNotificationFlowtriqResourceAPIKey(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationFlowtriqAPIKey")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationFlowtriqResourceAPIKeyConfig(name, "apikey0000000000"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.with_key",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.with_key",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact("apikey0000000000"),
					),
				},
			},
			{
				Config: testAccNotificationFlowtriqResourceAPIKeyConfig(name, "apikey1111111111"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.with_key",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact("apikey1111111111"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_flowtriq.with_key",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationFlowtriqResourceNoAPIKeyConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flowtriq.with_key",
						tfjsonpath.New("api_key"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

// TestAccNotificationFlowtriqResourceInvalidConfig covers the schema validation of the attributes.
func TestAccNotificationFlowtriqResourceInvalidConfig(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationFlowtriqInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationFlowtriqResourceInvalidWebhookURLConfig(name),
				ExpectError: regexp.MustCompile(`Invalid URL`),
			},
			{
				Config:      testAccNotificationFlowtriqResourceEmptyAPIKeyConfig(name),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length`),
			},
		},
	})
}

func testAccNotificationFlowtriqResourceConfig(name string, webhookURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_flowtriq" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
}
`, name, webhookURL)
}

func testAccNotificationFlowtriqResourceAPIKeyConfig(name string, apiKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_flowtriq" "with_key" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://app.flowtriq.com/api/webhooks/0123456789abcdef"
  api_key     = %[2]q
}
`, name, apiKey)
}

func testAccNotificationFlowtriqResourceNoAPIKeyConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_flowtriq" "with_key" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://app.flowtriq.com/api/webhooks/0123456789abcdef"
}
`, name)
}

func testAccNotificationFlowtriqResourceInvalidWebhookURLConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_flowtriq" "invalid" {
  name        = %[1]q
  webhook_url = "not-a-url"
}
`, name)
}

func testAccNotificationFlowtriqResourceEmptyAPIKeyConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_flowtriq" "invalid" {
  name        = %[1]q
  webhook_url = "https://app.flowtriq.com/api/webhooks/0123456789abcdef"
  api_key     = ""
}
`, name)
}
