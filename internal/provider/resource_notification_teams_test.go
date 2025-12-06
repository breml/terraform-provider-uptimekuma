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

func TestAccNotificationTeamsResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTeams")
	nameUpdated := acctest.RandomWithPrefix("NotificationTeamsUpdated")
	webhookURL := "https://example.webhook.office.com/webhookb2/test"
	webhookURLUpdated := "https://example.webhook.office.com/webhookb2/test-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTeamsResourceConfig(name, webhookURL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_teams.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_notification_teams.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_teams.test", tfjsonpath.New("webhook_url"), knownvalue.StringExact(webhookURL)),
				},
			},
			{
				Config: testAccNotificationTeamsResourceConfig(nameUpdated, webhookURLUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_teams.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_teams.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_teams.test", tfjsonpath.New("webhook_url"), knownvalue.StringExact(webhookURLUpdated)),
				},
			},
		},
	})
}

func testAccNotificationTeamsResourceConfig(name, webhookURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_teams" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
}
`, name, webhookURL)
}

func TestAccNotificationTeamsResourceImport(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTeamsImport")
	webhookURL := "https://outlook.office.com/webhook/00000000-0000-0000-0000-000000000000@00000000-0000-0000-0000-000000000000/IncomingWebhook/00000000000000000000000000000000/00000000-0000-0000-0000-000000000000"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTeamsResourceConfig(name, webhookURL),
			},
			{
				ResourceName:      "uptimekuma_notification_teams.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
