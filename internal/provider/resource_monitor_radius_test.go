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

func TestAccMonitorRadiusResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRadiusMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestRadiusMonitorUpdated")
	description := "Test Radius monitor description"
	descriptionUpdated := "Updated test Radius monitor description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRadiusResourceConfig(
					name,
					description,
					"radius.example.com",
					"testuser",
					"testpass",
					"testsecret",
					1812,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("radius.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("radius_username"),
						knownvalue.StringExact("testuser"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(1812),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				// Refresh-only step: ensure no perpetual diff is produced
				// when the API does not return sensitive fields.
				RefreshState:       true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMonitorRadiusResourceConfig(
					nameUpdated,
					descriptionUpdated,
					"radius2.example.com",
					"newuser",
					"newpass",
					"newsecret",
					1813,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(descriptionUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("radius2.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("radius_username"),
						knownvalue.StringExact("newuser"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(1813),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:                         "uptimekuma_monitor_radius.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
				ImportStateVerifyIgnore:              []string{"radius_password", "radius_secret"},
			},
		},
	})
}

func testAccMonitorRadiusResourceConfig(
	name string,
	description string,
	hostname string,
	radiusUsername string,
	radiusPassword string,
	radiusSecret string,
	port int64,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_radius" "test" {
  name            = %[1]q
  description     = %[2]q
  hostname        = %[3]q
  radius_username = %[4]q
  radius_password = %[5]q
  radius_secret   = %[6]q
  port            = %[7]d
  active          = true
}
`, name, description, hostname, radiusUsername, radiusPassword, radiusSecret, port)
}

func TestAccMonitorRadiusResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRadiusMonitorMinimal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRadiusResourceConfigMinimal(
					name,
					"radius.example.com",
					"testuser",
					"testpass",
					"testsecret",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("radius.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("radius_username"),
						knownvalue.StringExact("testuser"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(1812),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorRadiusResourceConfigMinimal(
	name string,
	hostname string,
	radiusUsername string,
	radiusPassword string,
	radiusSecret string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_radius" "test" {
  name            = %[1]q
  hostname        = %[2]q
  radius_username = %[3]q
  radius_password = %[4]q
  radius_secret   = %[5]q
}
`, name, hostname, radiusUsername, radiusPassword, radiusSecret)
}

func TestAccMonitorRadiusResourceWithAllOptions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRadiusMonitorFull")
	description := "Full test Radius monitor"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRadiusResourceConfigWithAllOptions(name, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("radius.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("radius_username"),
						knownvalue.StringExact("testuser"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(1812),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("called_station_id"),
						knownvalue.StringExact("called-station"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("calling_station_id"),
						knownvalue.StringExact("calling-station"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccMonitorRadiusResourceConfigWithAllOptions(name string, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_radius" "test" {
  name               = %[1]q
  description        = %[2]q
  hostname           = "radius.example.com"
  radius_username    = "testuser"
  radius_password    = "testpass"
  radius_secret      = "testsecret"
  called_station_id  = "called-station"
  calling_station_id = "calling-station"
  port               = 1812
  interval           = 120
  retry_interval     = 60
  max_retries        = 5
  active             = true
  upside_down        = false
}
`, name, description)
}

func TestAccMonitorRadiusResourceWithParent(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestRadiusGroup")
	monitorName := acctest.RandomWithPrefix("TestRadiusMonitorWithParent")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRadiusResourceConfigWithParent(groupName, monitorName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(groupName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(monitorName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_radius.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccMonitorRadiusResourceConfigWithParent(groupName string, monitorName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_monitor_radius" "test" {
  name            = %[2]q
  hostname        = "radius.example.com"
  radius_username = "testuser"
  radius_password = "testpass"
  radius_secret   = "testsecret"
  parent          = uptimekuma_monitor_group.test.id
}
`, groupName, monitorName)
}
