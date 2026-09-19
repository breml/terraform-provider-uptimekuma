package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceMonitorSMTP(t *testing.T) {
	name := acctest.RandomWithPrefix("smtp-datasource-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMonitorSMTPConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.uptimekuma_monitor_smtp.by_id", "id"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_smtp.by_id", "name", name),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_monitor_smtp.by_id", "hostname", "smtp.example.com",
					),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_monitor_smtp.by_id", "domain_expiry_notification", "true",
					),
					resource.TestCheckResourceAttrSet("data.uptimekuma_monitor_smtp.by_name", "id"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_smtp.by_name", "name", name),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_monitor_smtp.by_name", "hostname", "smtp.example.com",
					),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_monitor_smtp.by_name", "domain_expiry_notification", "true",
					),
				),
			},
		},
	})
}

func testAccDataSourceMonitorSMTPConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_smtp" "test" {
  name                       = %[1]q
  hostname                   = "smtp.example.com"
  port                       = 587
  domain_expiry_notification = true
}

data "uptimekuma_monitor_smtp" "by_id" {
  id = uptimekuma_monitor_smtp.test.id
}

data "uptimekuma_monitor_smtp" "by_name" {
  name = uptimekuma_monitor_smtp.test.name
}
`, name)
}
