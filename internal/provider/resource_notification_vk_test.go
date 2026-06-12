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

func TestAccNotificationVKResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationVK")
	nameUpdated := acctest.RandomWithPrefix("NotificationVKUpdated")
	accessToken := "vk1.a.abcdefghijklmnopqrstuvwxyz"
	accessTokenUpdated := "vk1.a.zyxwvutsrqponmlkjihgfedcba"
	peerID := "12345"
	peerIDUpdated := "2000000001"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Minimal config: verify optional fields default correctly.
			{
				Config: testAccNotificationVKResourceConfigMinimal(name, accessToken, peerID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("peer_id"),
						knownvalue.StringExact(peerID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("api_version"),
						knownvalue.StringExact("5.199"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("dont_parse_links"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				Config: testAccNotificationVKResourceConfig(name, accessToken, peerID, "5.131", true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("peer_id"),
						knownvalue.StringExact(peerID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("api_version"),
						knownvalue.StringExact("5.131"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("dont_parse_links"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationVKResourceConfig(
					nameUpdated, accessTokenUpdated, peerIDUpdated, "5.199", false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("peer_id"),
						knownvalue.StringExact(peerIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("api_version"),
						knownvalue.StringExact("5.199"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_vk.test",
						tfjsonpath.New("dont_parse_links"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_vk.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"access_token"},
			},
		},
	})
}

func testAccNotificationVKResourceConfigMinimal(name string, accessToken string, peerID string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_vk" "test" {
  name         = %[1]q
  is_active    = true
  access_token = %[2]q
  peer_id      = %[3]q
}
`, name, accessToken, peerID)
}

func testAccNotificationVKResourceConfig(
	name string, accessToken string, peerID string, apiVersion string, dontParseLinks bool,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_vk" "test" {
  name             = %[1]q
  is_active        = true
  access_token     = %[2]q
  peer_id          = %[3]q
  api_version      = %[4]q
  dont_parse_links = %[5]t
}
`, name, accessToken, peerID, apiVersion, dontParseLinks)
}
