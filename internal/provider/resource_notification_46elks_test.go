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

func TestAccNotification46ElksResource(t *testing.T) {
	name := acctest.RandomWithPrefix("Notification46Elks")
	nameUpdated := acctest.RandomWithPrefix("Notification46ElksUpdated")
	username := "test_user_46elks"
	usernameUpdated := "test_user_46elks_updated"
	authToken := "test_auth_token_46elks"
	authTokenUpdated := "test_auth_token_46elks_updated"
	fromNumber := "+1234567890"
	fromNumberUpdated := "+1234567891"
	toNumber := "+0987654321"
	toNumberUpdated := "+0987654322"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotification46ElksResourceConfig(
					name,
					username,
					authToken,
					fromNumber,
					toNumber,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact(authToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("from_number"),
						knownvalue.StringExact(fromNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact(toNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotification46ElksResourceConfig(
					nameUpdated,
					usernameUpdated,
					authTokenUpdated,
					fromNumberUpdated,
					toNumberUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(usernameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact(authTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("from_number"),
						knownvalue.StringExact(fromNumberUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_46elks.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact(toNumberUpdated),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_46elks.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotification46ElksResourceConfig(
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
`, name, username, authToken, fromNumber, toNumber)
}
