package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccNotificationSignalgridResource covers create, an update of all attributes except
// apply_existing, and an import round-trip.
func TestAccNotificationSignalgridResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSignalgrid")
	nameUpdated := acctest.RandomWithPrefix("NotificationSignalgridUpdated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSignalgridResourceConfig(name, "client-key-1", "alerts"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("client_key"),
						knownvalue.StringExact("client-key-1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("channel"),
						knownvalue.StringExact("alerts"),
					),
				},
			},
			{
				Config: testAccNotificationSignalgridResourceUpdatedConfig(nameUpdated, "client-key-2", "operations"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("client_key"),
						knownvalue.StringExact("client-key-2"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("channel"),
						knownvalue.StringExact("operations"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_signalgrid.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationSignalgridResourceInvalidConfig covers the length validators and the errors
// for a configuration that omits a required attribute.
func TestAccNotificationSignalgridResourceInvalidConfig(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSignalgridInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationSignalgridResourceConfig(name, "", "alerts"),
				ExpectError: regexp.MustCompile(`Attribute client_key string length must be at least 1`),
			},
			{
				Config:      testAccNotificationSignalgridResourceConfig(name, "client-key-1", ""),
				ExpectError: regexp.MustCompile(`Attribute channel string length must be at least 1`),
			},
			{
				Config:      testAccNotificationSignalgridResourceMissingClientKeyConfig(name),
				ExpectError: regexp.MustCompile(`one \(and only one\) of \[client_key_wo\] is required`),
			},
			{
				Config:      testAccNotificationSignalgridResourceMissingChannelConfig(name),
				ExpectError: regexp.MustCompile(`The argument "channel" is required`),
			},
		},
	})
}

func testAccNotificationSignalgridResourceConfig(name string, clientKey string, channel string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_signalgrid" "test" {
  name       = %[1]q
  client_key = %[2]q
  channel    = %[3]q
}
`, name, clientKey, channel)
}

func testAccNotificationSignalgridResourceUpdatedConfig(name string, clientKey string, channel string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_signalgrid" "test" {
  name       = %[1]q
  is_default = true
  client_key = %[2]q
  channel    = %[3]q
}
`, name, clientKey, channel)
}

func testAccNotificationSignalgridResourceMissingClientKeyConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_signalgrid" "invalid" {
  name    = %[1]q
  channel = "alerts"
}
`, name)
}

func testAccNotificationSignalgridResourceMissingChannelConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_signalgrid" "invalid" {
  name       = %[1]q
  client_key = "client-key-1"
}
`, name)
}
