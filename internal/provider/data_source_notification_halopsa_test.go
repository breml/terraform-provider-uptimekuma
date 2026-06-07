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

func TestAccNotificationHaloPSADataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationHaloPSA")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationHaloPSADataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_halopsa.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_halopsa.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccNotificationHaloPSADataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_halopsa.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationHaloPSADataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_halopsa" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://halopsa.example.com/webhook/XXXXXXXX"
  username    = "testuser"
  password    = "testpassword"
}

data "uptimekuma_notification_halopsa" "test" {
  name = uptimekuma_notification_halopsa.test.name
}
`, name)
}

func testAccNotificationHaloPSADataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_halopsa" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://halopsa.example.com/webhook/XXXXXXXX"
  username    = "testuser"
  password    = "testpassword"
}

data "uptimekuma_notification_halopsa" "test" {
  id = uptimekuma_notification_halopsa.test.id
}
`, name)
}
