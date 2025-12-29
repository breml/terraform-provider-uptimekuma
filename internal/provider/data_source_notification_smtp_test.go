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

func TestAccNotificationSMTPDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationSMTP")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMTPDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smtp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationSMTPDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smtp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSMTPDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smtp" "test" {
  name  = %[1]q
  host  = "smtp.example.com"
  from  = "uptime-kuma@example.com"
  to    = "admin@example.com"
  port  = 587
  is_active = true
}

data "uptimekuma_notification_smtp" "test" {
  name = uptimekuma_notification_smtp.test.name
}
`, name)
}

func testAccNotificationSMTPDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smtp" "test" {
  name  = %[1]q
  host  = "smtp.example.com"
  from  = "uptime-kuma@example.com"
  to    = "admin@example.com"
  port  = 587
  is_active = true
}

data "uptimekuma_notification_smtp" "test" {
  id = uptimekuma_notification_smtp.test.id
}
`, name)
}
