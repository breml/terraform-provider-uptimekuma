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

func TestAccNotificationYZJResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationYZJ")
	nameUpdated := acctest.RandomWithPrefix("NotificationYZJUpdated")
	webhookURL := "https://yzj.example.com/webhook/test"
	webhookURLUpdated := "https://yzj.example.com/webhook/updated"
	token := "test-token-placeholder"
	tokenUpdated := "test-token-updated-placeholder"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationYZJResourceConfig(name, webhookURL, token),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_yzj.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_yzj.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_yzj.test",
						tfjsonpath.New("token"),
						knownvalue.StringExact(token),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_yzj.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationYZJResourceConfig(nameUpdated, webhookURLUpdated, tokenUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_yzj.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_yzj.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_yzj.test",
						tfjsonpath.New("token"),
						knownvalue.StringExact(tokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_yzj.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_yzj.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationYZJImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"webhook_url", "token"},
			},
		},
	})
}

func testAccNotificationYZJImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_yzj.test"]
	return rs.Primary.Attributes["id"], nil
}

func testAccNotificationYZJResourceConfig(name string, webhookURL string, token string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_yzj" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
  token       = %[3]q
}
`, name, webhookURL, token)
}
