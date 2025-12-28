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

func TestAccNotificationGotifyResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGotify")
	nameUpdated := acctest.RandomWithPrefix("NotificationGotifyUpdated")
	serverURL := "https://gotify.example.com"
	serverURLUpdated := "https://gotify.updated.com"
	token := "AGe0Ks4WV5fEJkX"
	tokenUpdated := "AGe0Ks4WV5fEJkY"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGotifyResourceConfig(
					name,
					serverURL,
					token,
					8,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gotify.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gotify.test",
						tfjsonpath.New("server_url"),
						knownvalue.StringExact(serverURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gotify.test",
						tfjsonpath.New("application_token"),
						knownvalue.StringExact(token),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gotify.test",
						tfjsonpath.New("priority"),
						knownvalue.Int64Exact(8),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gotify.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationGotifyResourceConfig(
					nameUpdated,
					serverURLUpdated,
					tokenUpdated,
					5,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gotify.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gotify.test",
						tfjsonpath.New("server_url"),
						knownvalue.StringExact(serverURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gotify.test",
						tfjsonpath.New("application_token"),
						knownvalue.StringExact(tokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gotify.test",
						tfjsonpath.New("priority"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gotify.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_gotify.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationGotifyResourceConfig(
	name string,
	serverURL string,
	applicationToken string,
	priority int64,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_gotify" "test" {
  name                 = %[1]q
  is_active            = true
  server_url           = %[2]q
  application_token    = %[3]q
  priority             = %[4]d
}
`, name, serverURL, applicationToken, priority)
}
