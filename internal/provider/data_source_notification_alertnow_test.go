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

func TestAccNotificationAlertNowDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationAlertNow")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAlertNowDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_alertnow.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationAlertNowDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_alertnow.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationAlertNowDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_alertnow" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://alertnow.example.com/webhook"
}

data "uptimekuma_notification_alertnow" "test" {
  name = uptimekuma_notification_alertnow.test.name
}
`, name)
}

func testAccNotificationAlertNowDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_alertnow" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://alertnow.example.com/webhook"
}

data "uptimekuma_notification_alertnow" "test" {
  id = uptimekuma_notification_alertnow.test.id
}
`, name)
}
