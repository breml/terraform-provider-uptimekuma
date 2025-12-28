package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationGoogleChatResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGoogleChat")
	nameUpdated := acctest.RandomWithPrefix("NotificationGoogleChatUpdated")
	webhookURL := "https://chat.googleapis.com/v1/spaces/SPACE_ID/messages?key=KEY&token=TOKEN"
	webhookURLUpdated := "https://chat.googleapis.com/v1/spaces/UPDATED_SPACE_ID/messages?key=KEY&token=TOKEN"
	template := "Alert: {msg}"
	templateUpdated := "Updated Alert: {msg}"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGoogleChatResourceConfig(
					name,
					webhookURL,
					false,
					template,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_googlechat.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_googlechat.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_googlechat.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_googlechat.test",
						tfjsonpath.New("use_template"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				Config: testAccNotificationGoogleChatResourceConfig(
					nameUpdated,
					webhookURLUpdated,
					true,
					templateUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_googlechat.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_googlechat.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_googlechat.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_googlechat.test",
						tfjsonpath.New("use_template"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_googlechat.test",
						tfjsonpath.New("template"),
						knownvalue.StringExact(templateUpdated),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_googlechat.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationGoogleChatImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"webhook_url"},
			},
		},
	})
}

//nolint:revive // flag-parameter is fine for test helper function
func testAccNotificationGoogleChatResourceConfig(
	name string,
	webhookURL string,
	useTemplate bool,
	template string,
) string {
	templateField := ""
	if useTemplate {
		templateField = fmt.Sprintf("\n  template = %q", template)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_googlechat" "test" {
  name         = %[1]q
  webhook_url  = %[2]q
  use_template = %[3]t%[4]s
}
`, name, webhookURL, useTemplate, templateField)
}

func testAccNotificationGoogleChatImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_googlechat.test"]
	return rs.Primary.Attributes["id"], nil
}
