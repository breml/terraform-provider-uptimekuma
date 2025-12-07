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

func TestAccNotificationDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotification")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_notification.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
			{
				Config: testAccNotificationDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_notification.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
		},
	})
}

func testAccNotificationDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification" "test" {
  name      = %[1]q
  type      = "webhook"
  is_active = true
  config = jsonencode({
    webhookURL = "https://example.com/webhook"
  })
}

data "uptimekuma_notification" "test" {
  name = uptimekuma_notification.test.name
}
`, name)
}

func testAccNotificationDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification" "test" {
  name      = %[1]q
  type      = "webhook"
  is_active = true
  config = jsonencode({
    webhookURL = "https://example.com/webhook"
  })
}

data "uptimekuma_notification" "test" {
  id = uptimekuma_notification.test.id
}
`, name)
}
