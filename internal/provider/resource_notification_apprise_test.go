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
	title := "Test Notification"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAppriseResourceConfig(name, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("apprise_url"), knownvalue.StringExact("discord://webhook_id/webhook_token")),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("title"), knownvalue.StringExact("")),
				},
			},
			{
				Config: testAccNotificationAppriseResourceConfigWithTitle(nameUpdated, title),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("apprise_url"), knownvalue.StringExact("discord://webhook_id/webhook_token")),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("title"), knownvalue.StringExact(title)),
				},
			},
			{
				Config: testAccNotificationAppriseResourceConfig(nameUpdated, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("apprise_url"), knownvalue.StringExact("discord://webhook_id/webhook_token")),
					statecheck.ExpectKnownValue("uptimekuma_notification_apprise.test", tfjsonpath.New("title"), knownvalue.StringExact("")),
				},
			},
		},
	})
}

func testAccNotificationAppriseResourceConfig(name, title string) string {
	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_apprise" "test" {
  name        = %[1]q
  is_active   = true
  apprise_url = "discord://webhook_id/webhook_token"
`, name)

	if title != "" {
		config += fmt.Sprintf(`  title       = %[1]q
`, title)
	}

	config += `}
`
	return config
}

func testAccNotificationAppriseResourceConfigWithTitle(name, title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_apprise" "test" {
  name        = %[1]q
  is_active   = true
  apprise_url = "discord://webhook_id/webhook_token"
  title       = %[2]q
}
`, name, title)
}
