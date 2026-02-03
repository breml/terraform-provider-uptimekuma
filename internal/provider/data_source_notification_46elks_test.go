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

func TestAccNotification46ElksDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("Notification46Elks")
	username := "test_user_46elks"
	authToken := "test_auth_token_46elks"
	fromNumber := "+1234567890"
	toNumber := "+0987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotification46ElksDataSourceConfig(
					name,
					username,
					authToken,
					fromNumber,
					toNumber,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_46elks.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_46elks.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccNotification46ElksDataSourceConfig(
	name string,
	username string,
	authToken string,
	fromNumber string,
	toNumber string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_46elks" "test" {
  name        = %[1]q
  is_active   = true
  username    = %[2]q
  auth_token  = %[3]q
  from_number = %[4]q
  to_number   = %[5]q
}

data "uptimekuma_notification_46elks" "test" {
  name = uptimekuma_notification_46elks.test.name
}
`, name, username, authToken, fromNumber, toNumber)
}
