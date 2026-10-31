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

// TestAccNotificationNotifyAppResource covers create without icon_url, an update of all attributes
// except apply_existing, removing icon_url again, and an import round-trip with and without icon_url.
func TestAccNotificationNotifyAppResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNotifyApp")
	nameUpdated := acctest.RandomWithPrefix("NotificationNotifyAppUpdated")

	const address = "uptimekuma_notification_notifyapp.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNotifyAppResourceConfig(name, "ABC12345", "token-1"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("is_default"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("apply_existing"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("device_id"), knownvalue.StringExact("ABC12345")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("token"), knownvalue.StringExact("token-1")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("icon_url"), knownvalue.Null()),
				},
			},
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationNotifyAppResourceUpdatedConfig(
					nameUpdated,
					"DEF67890",
					"token-2",
					"https://example.com/icon.png",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("is_default"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("apply_existing"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("device_id"), knownvalue.StringExact("DEF67890")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("token"), knownvalue.StringExact("token-2")),
					statecheck.ExpectKnownValue(
						address,
						tfjsonpath.New("icon_url"),
						knownvalue.StringExact("https://example.com/icon.png"),
					),
				},
			},
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationNotifyAppResourceConfig(nameUpdated, "DEF67890", "token-2"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, tfjsonpath.New("icon_url"), knownvalue.Null()),
				},
			},
			// The import reads the server's value, so this proves the update removed icon_url there too.
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationNotifyAppResourceCreateWithIconURL covers icon_url set on create.
func TestAccNotificationNotifyAppResourceCreateWithIconURL(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNotifyAppIcon")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNotifyAppResourceUpdatedConfig(
					name,
					"ABC12345",
					"token-1",
					"https://example.com/icon.png",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifyapp.test",
						tfjsonpath.New("icon_url"),
						knownvalue.StringExact("https://example.com/icon.png"),
					),
				},
			},
		},
	})
}

// TestAccNotificationNotifyAppResourceInvalidConfig covers the length validators and the errors
// for a configuration that omits a required attribute.
func TestAccNotificationNotifyAppResourceInvalidConfig(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNotifyAppInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationNotifyAppResourceConfig(name, "", "token-1"),
				ExpectError: regexp.MustCompile(`Attribute device_id string length must be at least 1`),
			},
			{
				Config:      testAccNotificationNotifyAppResourceConfig(name, "ABC12345", ""),
				ExpectError: regexp.MustCompile(`Attribute token string length must be at least 1`),
			},
			{
				Config:      testAccNotificationNotifyAppResourceUpdatedConfig(name, "ABC12345", "token-1", ""),
				ExpectError: regexp.MustCompile(`Attribute icon_url string length must be at least 1`),
			},
			{
				Config:      testAccNotificationNotifyAppResourceMissingTokenConfig(name),
				ExpectError: regexp.MustCompile(`one \(and only one\) of \[token_wo\] is required`),
			},
			{
				Config:      testAccNotificationNotifyAppResourceMissingDeviceIDConfig(name),
				ExpectError: regexp.MustCompile(`The argument "device_id" is required`),
			},
		},
	})
}

func testAccNotificationNotifyAppResourceConfig(name string, deviceID string, token string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_notifyapp" "test" {
  name      = %[1]q
  device_id = %[2]q
  token     = %[3]q
}
`, name, deviceID, token)
}

func testAccNotificationNotifyAppResourceUpdatedConfig(
	name string,
	deviceID string,
	token string,
	iconURL string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_notifyapp" "test" {
  name       = %[1]q
  is_default = true
  device_id  = %[2]q
  token      = %[3]q
  icon_url   = %[4]q
}
`, name, deviceID, token, iconURL)
}

func testAccNotificationNotifyAppResourceMissingTokenConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_notifyapp" "invalid" {
  name      = %[1]q
  device_id = "ABC12345"
}
`, name)
}

func testAccNotificationNotifyAppResourceMissingDeviceIDConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_notifyapp" "invalid" {
  name  = %[1]q
  token = "token-1"
}
`, name)
}
