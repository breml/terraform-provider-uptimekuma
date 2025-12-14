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

func TestAccNotificationAppriseResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationApprise")
	nameUpdated := acctest.RandomWithPrefix("NotificationAppriseUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAppriseResourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("apprise_url"), knownvalue.StringExact("discord://webhook_id/webhook_token")),
				},
			},
			{
				Config: testAccNotificationAppriseResourceConfig(nameUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("apprise_url"), knownvalue.StringExact("discord://webhook_id/webhook_token")),
				},
			},
		},
	})
}

func testAccNotificationAppriseResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_apprise" "test" {
  name        = %[1]q
  is_active   = true
  apprise_url = "discord://webhook_id/webhook_token"
}
`, name)
}
