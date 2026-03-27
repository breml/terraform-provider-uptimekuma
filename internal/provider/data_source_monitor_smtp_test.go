package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceMonitorSMTPByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMonitorSMTPByIDConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.uptimekuma_monitor_smtp.test", "id"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_smtp.test", "name", "smtp-datasource-test"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_smtp.test", "hostname", "smtp.example.com"),
				),
			},
		},
	})
}

func TestAccDataSourceMonitorSMTPByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMonitorSMTPByNameConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.uptimekuma_monitor_smtp.test", "id"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_smtp.test", "name", "smtp-datasource-test"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_smtp.test", "hostname", "smtp.example.com"),
				),
			},
		},
	})
}

func testAccDataSourceMonitorSMTPByIDConfig() string {
	return providerConfig() + `
resource "uptimekuma_monitor_smtp" "test" {
  name     = "smtp-datasource-test"
  hostname = "smtp.example.com"
  port     = 587
}

data "uptimekuma_monitor_smtp" "test" {
  id = uptimekuma_monitor_smtp.test.id
}
`
}

func testAccDataSourceMonitorSMTPByNameConfig() string {
	return providerConfig() + `
resource "uptimekuma_monitor_smtp" "test" {
  name     = "smtp-datasource-test"
  hostname = "smtp.example.com"
  port     = 587
}

data "uptimekuma_monitor_smtp" "test" {
  name = uptimekuma_monitor_smtp.test.name
}
`
}
