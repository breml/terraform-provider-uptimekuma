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

func TestAccNotificationCellsyntDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationCellsynt")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationCellsyntDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationCellsyntDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationCellsyntDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_cellsynt" "test" {
  name            = %[1]q
  is_active       = true
  login           = "testuser"
  password        = "testpass123"
  destination     = "+46701234567"
  originator      = "TestSender"
  originator_type = "Numeric"
  allow_long_sms  = false
}

data "uptimekuma_notification_cellsynt" "test" {
  name = uptimekuma_notification_cellsynt.test.name
}
`, name)
}

func testAccNotificationCellsyntDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_cellsynt" "test" {
  name            = %[1]q
  is_active       = true
  login           = "testuser"
  password        = "testpass123"
  destination     = "+46701234567"
  originator      = "TestSender"
  originator_type = "Numeric"
  allow_long_sms  = false
}

data "uptimekuma_notification_cellsynt" "test" {
  id = uptimekuma_notification_cellsynt.test.id
}
`, name)
}
