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

func TestAccNotificationStackfieldResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationStackfield")
	nameUpdated := acctest.RandomWithPrefix("NotificationStackfieldUpdated")
	webhookURL := "https://api.stackfield.com/hooks/XXXXXXXXXXXXXXXXXXXXXXXX"
	webhookURLUpdated := "https://api.stackfield.com/hooks/YYYYYYYYYYYYYYYYYYYYYYYY"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationStackfieldResourceConfig(name, webhookURL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_stackfield.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_stackfield.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_stackfield.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationStackfieldResourceConfig(nameUpdated, webhookURLUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_stackfield.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_stackfield.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_stackfield.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_stackfield.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"webhook_url"},
			},
		},
	})
}

func testAccNotificationStackfieldResourceConfig(name string, webhookURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_stackfield" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
}
`, name, webhookURL)
}
