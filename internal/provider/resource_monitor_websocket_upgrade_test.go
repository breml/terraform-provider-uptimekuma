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

func TestAccMonitorWebsocketUpgradeResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestWebsocketUpgradeMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestWebsocketUpgradeMonitorUpdated")
	url := "wss://echo.websocket.org"
	description := "Test Websocket Upgrade monitor with description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorWebsocketUpgradeResourceConfigWithDescription(name, url, 60, 48, description),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("method"),
						knownvalue.StringExact("GET"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("timeout"),
						knownvalue.Int64Exact(48),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("ws_ignore_sec_websocket_accept_header"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("ws_subprotocol"),
						knownvalue.Null(),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_websocket_upgrade.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitorWebsocketUpgradeResourceConfigWithDescription(nameUpdated, url, 120, 60, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("description"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("timeout"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorWebsocketUpgradeResourceConfigWithDescription(
	name string, url string,
	interval int64, timeout int64,
	description string,
) string {
	descField := ""
	if description != "" {
		descField = fmt.Sprintf("  description = %q", description)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_websocket_upgrade" "test" {
  name     = %[1]q
  url      = %[2]q
%[3]s
  interval = %[4]d
  timeout  = %[5]d
  active   = true
}
`, name, url, descField, interval, timeout)
}

func TestAccMonitorWebsocketUpgradeResourceWithAuth(t *testing.T) {
	name := acctest.RandomWithPrefix("TestWebsocketUpgradeMonitorWithAuth")
	url := "wss://echo.websocket.org"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorWebsocketUpgradeResourceConfigWithAuth(name, url, "user", "pass"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("auth_method"),
						knownvalue.StringExact("basic"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("basic_auth_user"),
						knownvalue.StringExact("user"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("basic_auth_pass"),
						knownvalue.StringExact("pass"),
					),
				},
			},
		},
	})
}

func testAccMonitorWebsocketUpgradeResourceConfigWithAuth(name string, url string, user string, pass string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_websocket_upgrade" "test" {
  name            = %[1]q
  url             = %[2]q
  auth_method     = "basic"
  basic_auth_user = %[3]q
  basic_auth_pass = %[4]q
}
`, name, url, user, pass)
}

func TestAccMonitorWebsocketUpgradeResourceWithOAuthAudience(t *testing.T) {
	name := acctest.RandomWithPrefix("TestWebsocketUpgradeOAuthAudience")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorWebsocketUpgradeResourceConfigWithOAuthAudience(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("oauth_audience"),
						knownvalue.StringExact("https://api.example.com/resource"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_monitor_websocket_upgrade.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth_client_secret"},
			},
		},
	})
}

func testAccMonitorWebsocketUpgradeResourceConfigWithOAuthAudience(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_websocket_upgrade" "test" {
  name                = %[1]q
  url                 = "wss://echo.websocket.org"
  auth_method         = "oauth2-cc"
  oauth_auth_method   = "client_secret_basic"
  oauth_token_url     = "https://auth.example.com/token"
  oauth_client_id     = "client-id"
  oauth_client_secret = "client-secret"
  oauth_scopes        = "read"
  oauth_audience      = "https://api.example.com/resource"
}
`, name)
}

func TestAccMonitorWebsocketUpgradeResourceWithWSOptions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestWebsocketUpgradeMonitorWithWSOptions")
	url := "wss://echo.websocket.org"
	subprotocol := "chat,superchat"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorWebsocketUpgradeResourceConfigWithWSOptions(name, url, subprotocol, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("ws_subprotocol"),
						knownvalue.StringExact(subprotocol),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_websocket_upgrade.test",
						tfjsonpath.New("ws_ignore_sec_websocket_accept_header"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorWebsocketUpgradeResourceConfigWithWSOptions(
	name string, url string, subprotocol string, ignoreAcceptHeader bool,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_websocket_upgrade" "test" {
  name                                  = %[1]q
  url                                   = %[2]q
  ws_subprotocol                        = %[3]q
  ws_ignore_sec_websocket_accept_header = %[4]t
}
`, name, url, subprotocol, ignoreAcceptHeader)
}
