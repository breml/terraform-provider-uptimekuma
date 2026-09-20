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

// TestAccNotificationOoredooResource covers the default shape, where server_url stays unset and
// must therefore round-trip as null rather than as the empty string.
func TestAccNotificationOoredooResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOoredoo")
	nameUpdated := acctest.RandomWithPrefix("NotificationOoredooUpdated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOoredooResourceConfig(
					name,
					"ooredoo-user",
					"accesskey0000000000",
					"bearertoken0000000000000000",
					"9601234567",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("ooredoo-user"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("access_key"),
						knownvalue.StringExact("accesskey0000000000"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("bearer_token"),
						knownvalue.StringExact("bearertoken0000000000000000"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact("9601234567"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("server_url"),
						knownvalue.Null(),
					),
				},
			},
			{
				Config: testAccNotificationOoredooResourceConfig(
					nameUpdated,
					"ooredoo-user-updated",
					"accesskey1111111111",
					"bearertoken1111111111111111",
					"1234567,7654321",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("ooredoo-user-updated"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("access_key"),
						knownvalue.StringExact("accesskey1111111111"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("bearer_token"),
						knownvalue.StringExact("bearertoken1111111111111111"),
					),
					// Multiple recipients are passed through as configured, the separator handling
					// and the 960 prefixing happen in Uptime Kuma at send time.
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact("1234567,7654321"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.test",
						tfjsonpath.New("server_url"),
						knownvalue.Null(),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_ooredoo.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationOoredooResourceServerURL covers the custom endpoint, including clearing it
// again, which has to fall back to null so Uptime Kuma applies its own default endpoint.
func TestAccNotificationOoredooResourceServerURL(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOoredooServerURL")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOoredooResourceServerURLConfig(name, "https://sms.example.com/bulk_sms/v2"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.custom",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.custom",
						tfjsonpath.New("server_url"),
						knownvalue.StringExact("https://sms.example.com/bulk_sms/v2"),
					),
				},
			},
			{
				Config: testAccNotificationOoredooResourceServerURLConfig(name, "https://other.example.com/bulk_sms/v2"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.custom",
						tfjsonpath.New("server_url"),
						knownvalue.StringExact("https://other.example.com/bulk_sms/v2"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_ooredoo.custom",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationOoredooResourceDefaultServerURLConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_ooredoo.custom",
						tfjsonpath.New("server_url"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

// TestAccNotificationOoredooResourceInvalidConfig covers the schema validation of the attributes.
func TestAccNotificationOoredooResourceInvalidConfig(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOoredooInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationOoredooResourceEmptyToNumberConfig(name),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length`),
			},
			{
				Config:      testAccNotificationOoredooResourceInvalidServerURLConfig(name),
				ExpectError: regexp.MustCompile(`Invalid URL`),
			},
		},
	})
}

func testAccNotificationOoredooResourceConfig(
	name string, username string, accessKey string, bearerToken string, toNumber string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ooredoo" "test" {
  name         = %[1]q
  is_active    = true
  username     = %[2]q
  access_key   = %[3]q
  bearer_token = %[4]q
  to_number    = %[5]q
}
`, name, username, accessKey, bearerToken, toNumber)
}

func testAccNotificationOoredooResourceServerURLConfig(name string, serverURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ooredoo" "custom" {
  name         = %[1]q
  is_active    = true
  username     = "ooredoo-user"
  access_key   = "accesskey0000000000"
  bearer_token = "bearertoken0000000000000000"
  to_number    = "9601234567"
  server_url   = %[2]q
}
`, name, serverURL)
}

func testAccNotificationOoredooResourceDefaultServerURLConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ooredoo" "custom" {
  name         = %[1]q
  is_active    = true
  username     = "ooredoo-user"
  access_key   = "accesskey0000000000"
  bearer_token = "bearertoken0000000000000000"
  to_number    = "9601234567"
}
`, name)
}

func testAccNotificationOoredooResourceEmptyToNumberConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ooredoo" "invalid" {
  name         = %[1]q
  username     = "ooredoo-user"
  access_key   = "accesskey0000000000"
  bearer_token = "bearertoken0000000000000000"
  to_number    = ""
}
`, name)
}

func testAccNotificationOoredooResourceInvalidServerURLConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ooredoo" "invalid" {
  name         = %[1]q
  username     = "ooredoo-user"
  access_key   = "accesskey0000000000"
  bearer_token = "bearertoken0000000000000000"
  to_number    = "9601234567"
  server_url   = "not-a-url"
}
`, name)
}
