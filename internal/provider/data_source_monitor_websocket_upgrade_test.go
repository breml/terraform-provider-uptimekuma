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

func TestAccMonitorWebsocketUpgradeDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestWebsocketUpgradeMonitor")
	url := "wss://echo.websocket.org"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorWebsocketUpgradeDataSourceConfig(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("ws_subprotocol"),
						knownvalue.StringExact("chat"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("ws_ignore_sec_websocket_accept_header"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorWebsocketUpgradeDataSourceConfigByID(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("ws_subprotocol"),
						knownvalue.StringExact("chat"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("ws_ignore_sec_websocket_accept_header"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorWebsocketUpgradeDataSourceConfig(name string, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_websocket_upgrade" "test" {
  name                                   = %[1]q
  url                                    = %[2]q
  ws_subprotocol                         = "chat"
  ws_ignore_sec_websocket_accept_header  = true
}

data "uptimekuma_monitor_websocket_upgrade" "test" {
  name = uptimekuma_monitor_websocket_upgrade.test.name
}
`, name, url)
}

func testAccMonitorWebsocketUpgradeDataSourceConfigByID(name string, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_websocket_upgrade" "test" {
  name                                   = %[1]q
  url                                    = %[2]q
  ws_subprotocol                         = "chat"
  ws_ignore_sec_websocket_accept_header  = true
}

data "uptimekuma_monitor_websocket_upgrade" "test" {
  id = uptimekuma_monitor_websocket_upgrade.test.id
}
`, name, url)
}
