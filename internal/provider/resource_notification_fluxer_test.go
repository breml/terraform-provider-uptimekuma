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

func TestAccNotificationFluxerResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationFluxer")
	nameUpdated := acctest.RandomWithPrefix("NotificationFluxerUpdated")
	webhookURL := "https://fluxer.example.com/webhook/XXXXXXXX"
	webhookURLUpdated := "https://fluxer.example.com/webhook/YYYYYYYY"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Minimal config: verify optional fields default to null.
			{
				Config: testAccNotificationFluxerResourceConfigMinimal(name, webhookURL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("username"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("prefix_message"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("disable_url"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("use_message_template"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("message_format"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("message_template"),
						knownvalue.Null(),
					),
				},
			},
			{
				Config: testAccNotificationFluxerResourceConfig(
					name,
					webhookURL,
					"test-bot",
					"alert:",
					true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("test-bot"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("prefix_message"),
						knownvalue.StringExact("alert:"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("disable_url"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationFluxerResourceConfig(
					nameUpdated,
					webhookURLUpdated,
					"updated-bot",
					"warning:",
					false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("updated-bot"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("prefix_message"),
						knownvalue.StringExact("warning:"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("disable_url"),
						knownvalue.Bool(false),
					),
				},
			},
			// Set pointer-based template fields and verify they round-trip correctly.
			{
				Config: testAccNotificationFluxerResourceConfigWithTemplate(
					nameUpdated,
					webhookURLUpdated,
					true,
					"minimalist",
					"Alert: {{name}}",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("use_message_template"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("message_format"),
						knownvalue.StringExact("minimalist"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("message_template"),
						knownvalue.StringExact("Alert: {{name}}"),
					),
				},
			},
			// Clear pointer-based fields and verify they return to null.
			{
				Config: testAccNotificationFluxerResourceConfigMinimal(nameUpdated, webhookURLUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("use_message_template"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("message_format"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_fluxer.test",
						tfjsonpath.New("message_template"),
						knownvalue.Null(),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_fluxer.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"webhook_url"},
			},
		},
	})
}

func testAccNotificationFluxerResourceConfigMinimal(name string, webhookURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_fluxer" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
}
`, name, webhookURL)
}

func testAccNotificationFluxerResourceConfig(
	name string, webhookURL string, username string, prefixMessage string, disableURL bool,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_fluxer" "test" {
  name           = %[1]q
  is_active      = true
  webhook_url    = %[2]q
  username       = %[3]q
  prefix_message = %[4]q
  disable_url    = %[5]t
}
`, name, webhookURL, username, prefixMessage, disableURL)
}

func testAccNotificationFluxerResourceConfigWithTemplate(
	name string, webhookURL string, useMessageTemplate bool, messageFormat string, messageTemplate string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_fluxer" "test" {
  name                 = %[1]q
  is_active            = true
  webhook_url          = %[2]q
  use_message_template = %[3]t
  message_format       = %[4]q
  message_template     = %[5]q
}
`, name, webhookURL, useMessageTemplate, messageFormat, messageTemplate)
}
