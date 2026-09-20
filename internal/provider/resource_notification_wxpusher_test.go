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

// TestAccNotificationWxPusherResource covers the single-token shape, an update of name and spt, and
// an import round-trip.
func TestAccNotificationWxPusherResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWxPusher")
	nameUpdated := acctest.RandomWithPrefix("NotificationWxPusherUpdated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWxPusherResourceConfig(name, "SPT_0123456789ab"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wxpusher.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wxpusher.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wxpusher.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wxpusher.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wxpusher.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wxpusher.test",
						tfjsonpath.New("spt"),
						knownvalue.StringExact("SPT_0123456789ab"),
					),
				},
			},
			{
				Config: testAccNotificationWxPusherResourceConfig(nameUpdated, "SPT_cdef01234567"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wxpusher.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wxpusher.test",
						tfjsonpath.New("spt"),
						knownvalue.StringExact("SPT_cdef01234567"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_wxpusher.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationWxPusherResourceMultipleTokens asserts that a comma separated list of tokens
// round-trips verbatim through create, update and import: the server splits and trims it only at
// send time, so the provider must not normalize it on either write path.
func TestAccNotificationWxPusherResourceMultipleTokens(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWxPusherMulti")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWxPusherResourceConfig(
					name,
					"SPT_0123456789ab, SPT_cdef01234567 ,SPT_89abcdef0123",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wxpusher.test",
						tfjsonpath.New("spt"),
						knownvalue.StringExact("SPT_0123456789ab, SPT_cdef01234567 ,SPT_89abcdef0123"),
					),
				},
			},
			{
				Config: testAccNotificationWxPusherResourceConfig(
					name,
					"SPT_89abcdef0123 ,SPT_0123456789ab",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wxpusher.test",
						tfjsonpath.New("spt"),
						knownvalue.StringExact("SPT_89abcdef0123 ,SPT_0123456789ab"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_wxpusher.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationWxPusherResourceInvalidConfig covers the spt length validator and the
// required-argument error.
func TestAccNotificationWxPusherResourceInvalidConfig(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWxPusherInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationWxPusherResourceEmptySPTConfig(name),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length`),
			},
			{
				Config:      testAccNotificationWxPusherResourceMissingSPTConfig(name),
				ExpectError: regexp.MustCompile(`The argument "spt" is required`),
			},
		},
	})
}

func testAccNotificationWxPusherResourceConfig(name string, spt string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_wxpusher" "test" {
  name      = %[1]q
  is_active = true
  spt       = %[2]q
}
`, name, spt)
}

func testAccNotificationWxPusherResourceEmptySPTConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_wxpusher" "invalid" {
  name = %[1]q
  spt  = ""
}
`, name)
}

func testAccNotificationWxPusherResourceMissingSPTConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_wxpusher" "invalid" {
  name = %[1]q
}
`, name)
}
