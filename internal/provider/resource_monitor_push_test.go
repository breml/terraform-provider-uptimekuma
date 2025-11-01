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

func TestAccMonitorPushResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPushMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestPushMonitorUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorPushResourceConfig(name, 60),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("push_token"), knownvalue.NotNull()),
				},
			},
			{
				Config: testAccMonitorPushResourceConfig(nameUpdated, 120),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(120)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("push_token"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccMonitorPushResourceConfig(name string, interval int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_push" "test" {
  name     = %[1]q
  interval = %[2]d
  active   = true
}
`, name, interval)
}

func TestAccMonitorPushResourceWithOptionalFields(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPushMonitorWithOptional")
	description := "Test push monitor with optional fields"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPushResourceConfigWithOptionalFields(name, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("retry_interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("resend_interval"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("max_retries"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("upside_down"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("push_token"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccMonitorPushResourceConfigWithOptionalFields(name, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_push" "test" {
  name            = %[1]q
  description     = %[2]q
  interval        = 60
  retry_interval  = 60
  resend_interval = 0
  max_retries     = 0
  upside_down     = false
  active          = true
}
`, name, description)
}

func TestAccMonitorPushResourceWithParent(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestPushGroup")
	monitorName := acctest.RandomWithPrefix("TestPushMonitorWithParent")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPushResourceConfigWithParent(groupName, monitorName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_group.test", tfjsonpath.New("name"), knownvalue.StringExact(groupName)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("name"), knownvalue.StringExact(monitorName)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("parent"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("uptimekuma_monitor_push.test", tfjsonpath.New("push_token"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccMonitorPushResourceConfigWithParent(groupName, monitorName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_monitor_push" "test" {
  name   = %[2]q
  parent = uptimekuma_monitor_group.test.id
}
`, groupName, monitorName)
}
