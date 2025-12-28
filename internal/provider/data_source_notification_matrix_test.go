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

func TestAccNotificationMatrixDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationMatrix")
	homeserverURL := "https://matrix.example.com"
	internalRoomID := "!abc123:example.com"
	accessToken := "syt_access_token_example_123"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationMatrixDataSourceConfig(
					name,
					homeserverURL,
					internalRoomID,
					accessToken,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_matrix.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_matrix.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationMatrixDataSourceConfig(
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

data "uptimekuma_notification_matrix" "test" {
  name = uptimekuma_notification_matrix.test.name
}
`, name, homeserverURL, internalRoomID, accessToken)
}
