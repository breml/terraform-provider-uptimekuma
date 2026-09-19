package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationHomeAssistantDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationHomeAssistant")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationHomeAssistantDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.uptimekuma_notification_homeassistant.by_name",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_notification_homeassistant.by_name",
						"name",
						name,
					),
					resource.TestCheckResourceAttrSet(
						"data.uptimekuma_notification_homeassistant.by_id",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_notification_homeassistant.by_id",
						"name",
						name,
					),
				),
			},
		},
	})
}

func testAccNotificationHomeAssistantDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_homeassistant" "test" {
  name                    = %[1]q
  is_active               = true
  home_assistant_url      = "https://homeassistant.example.com"
  long_lived_access_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"
  notification_service    = "notify.mobile_app"
}

data "uptimekuma_notification_homeassistant" "by_name" {
  name = uptimekuma_notification_homeassistant.test.name
}

data "uptimekuma_notification_homeassistant" "by_id" {
  id = uptimekuma_notification_homeassistant.test.id
}
`, name)
}
