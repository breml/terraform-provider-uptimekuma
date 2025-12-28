package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationMatrixResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationMatrix")
	nameUpdated := acctest.RandomWithPrefix("NotificationMatrixUpdated")
	homeserverURL := "https://matrix.example.com"
	homeserverURLUpdated := "https://matrix-updated.example.com"
	internalRoomID := "!abc123:example.com"
	internalRoomIDUpdated := "!xyz789:example.com"
	accessToken := "syt_access_token_example_123"
	accessTokenUpdated := "syt_access_token_updated_456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationMatrixResourceConfig(
					name,
					homeserverURL,
					internalRoomID,
					accessToken,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_matrix.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_matrix.test",
						tfjsonpath.New("homeserver_url"),
						knownvalue.StringExact(homeserverURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_matrix.test",
						tfjsonpath.New("internal_room_id"),
						knownvalue.StringExact(internalRoomID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_matrix.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationMatrixResourceConfig(
					nameUpdated,
					homeserverURLUpdated,
					internalRoomIDUpdated,
					accessTokenUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_matrix.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_matrix.test",
						tfjsonpath.New("homeserver_url"),
						knownvalue.StringExact(homeserverURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_matrix.test",
						tfjsonpath.New("internal_room_id"),
						knownvalue.StringExact(internalRoomIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_matrix.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_matrix.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationMatrixImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"access_token"},
			},
		},
	})
}

func testAccNotificationMatrixResourceConfig(
	name string, homeserverURL string, internalRoomID string, accessToken string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_matrix" "test" {
  name               = %[1]q
  is_active          = true
  homeserver_url     = %[2]q
  internal_room_id   = %[3]q
  access_token       = %[4]q
}
`, name, homeserverURL, internalRoomID, accessToken)
}

func testAccNotificationMatrixImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_matrix.test"]
	return rs.Primary.Attributes["id"], nil
}
