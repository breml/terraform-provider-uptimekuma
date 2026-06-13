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

func TestAccNotificationWhatsapp360messengerDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWhatsapp360messenger")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWhatsapp360messengerDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccNotificationWhatsapp360messengerDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationWhatsapp360messengerDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_whatsapp360messenger" "test" {
  name       = %[1]q
  is_active  = true
  auth_token = "auth-token-xxxxxxxx"
  recipient  = "+15551234567"
}

data "uptimekuma_notification_whatsapp360messenger" "test" {
  name = uptimekuma_notification_whatsapp360messenger.test.name
}
`, name)
}

func testAccNotificationWhatsapp360messengerDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_whatsapp360messenger" "test" {
  name       = %[1]q
  is_active  = true
  auth_token = "auth-token-xxxxxxxx"
  recipient  = "+15551234567"
}

data "uptimekuma_notification_whatsapp360messenger" "test" {
  id = uptimekuma_notification_whatsapp360messenger.test.id
}
`, name)
}
