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

func TestAccNotificationLineDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationLine")
	channelAccessToken := "channel_access_token_123456789abcdef"
	userID := "U1234567890abcdef1234567890abcdef"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationLineDataSourceConfig(
					name,
					channelAccessToken,
					userID,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_line.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_line.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationLineDataSourceConfig(name string, channelAccessToken string, userID string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_line" "test" {
  name                   = %[1]q
  is_active              = true
  channel_access_token   = %[2]q
  user_id                = %[3]q
}

data "uptimekuma_notification_line" "test" {
  name = uptimekuma_notification_line.test.name
}
`, name, channelAccessToken, userID)
}
