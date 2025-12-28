package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationGrafanaOncallDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGrafanaOncall")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGrafanaOncallDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.uptimekuma_notification_grafanaoncall.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_notification_grafanaoncall.test",
						"name",
						name,
					),
				),
			},
			{
				Config: testAccNotificationGrafanaOncallDataSourceConfigByID(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.uptimekuma_notification_grafanaoncall.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_notification_grafanaoncall.test",
						"name",
						name,
					),
				),
			},
		},
	})
}

func testAccNotificationGrafanaOncallDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_grafanaoncall" "test" {
  name               = %[1]q
  is_active          = true
  grafana_oncall_url = "https://grafana-oncall.example.com/integrations/v1/webhook/abc123/"
}

data "uptimekuma_notification_grafanaoncall" "test" {
  name = uptimekuma_notification_grafanaoncall.test.name
}
`, name)
}

func testAccNotificationGrafanaOncallDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_grafanaoncall" "test" {
  name               = %[1]q
  is_active          = true
  grafana_oncall_url = "https://grafana-oncall.example.com/integrations/v1/webhook/abc123/"
}

data "uptimekuma_notification_grafanaoncall" "test" {
  id = uptimekuma_notification_grafanaoncall.test.id
}
`, name)
}
