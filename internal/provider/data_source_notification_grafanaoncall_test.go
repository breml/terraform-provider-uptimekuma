package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationGrafanaOncallDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGrafanaOncall")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGrafanaOncallDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.uptimekuma_notification_grafanaoncall.by_name",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_notification_grafanaoncall.by_name",
						"name",
						name,
					),
					resource.TestCheckResourceAttrSet(
						"data.uptimekuma_notification_grafanaoncall.by_id",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_notification_grafanaoncall.by_id",
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

data "uptimekuma_notification_grafanaoncall" "by_name" {
  name = uptimekuma_notification_grafanaoncall.test.name
}

data "uptimekuma_notification_grafanaoncall" "by_id" {
  id = uptimekuma_notification_grafanaoncall.test.id
}
`, name)
}
