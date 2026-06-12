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

func TestAccNotificationTeltonikaDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTeltonika")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTeltonikaDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_teltonika.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationTeltonikaDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_teltonika.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationTeltonikaDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_teltonika" "test" {
  name         = %[1]q
  is_active    = true
  url          = "https://192.168.1.1"
  username     = "admin"
  password     = "test-password-123"
  phone_number = "+33600000000"
}

data "uptimekuma_notification_teltonika" "test" {
  name = uptimekuma_notification_teltonika.test.name
}
`, name)
}

func testAccNotificationTeltonikaDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_teltonika" "test" {
  name         = %[1]q
  is_active    = true
  url          = "https://192.168.1.1"
  username     = "admin"
  password     = "test-password-123"
  phone_number = "+33600000000"
}

data "uptimekuma_notification_teltonika" "test" {
  id = uptimekuma_notification_teltonika.test.id
}
`, name)
}
