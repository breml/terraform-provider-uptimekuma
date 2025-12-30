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

func TestAccNotificationLineResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationLine")
	nameUpdated := acctest.RandomWithPrefix("NotificationLineUpdated")
	channelAccessToken := "channel_access_token_123456789abcdef"
	channelAccessTokenUpdated := "channel_access_token_987654321fedcba"
	userID := "U1234567890abcdef1234567890abcdef"
	userIDUpdated := "U0987654321fedcba0987654321fedcba"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationLineResourceConfig(
					name,
					channelAccessToken,
					userID,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_line.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_line.test",
						tfjsonpath.New("channel_access_token"),
						knownvalue.StringExact(channelAccessToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_line.test",
						tfjsonpath.New("user_id"),
						knownvalue.StringExact(userID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_line.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationLineResourceConfig(
					nameUpdated,
					channelAccessTokenUpdated,
					userIDUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_line.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_line.test",
						tfjsonpath.New("channel_access_token"),
						knownvalue.StringExact(channelAccessTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_line.test",
						tfjsonpath.New("user_id"),
						knownvalue.StringExact(userIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_line.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccNotificationLineResourceConfig(name string, channelAccessToken string, userID string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_line" "test" {
  name                   = %[1]q
  is_active              = true
  channel_access_token   = %[2]q
  user_id                = %[3]q
}
`, name, channelAccessToken, userID)
}
