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

func TestAccNotificationFreemobileResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationFreemobile")
	nameUpdated := acctest.RandomWithPrefix("NotificationFreemobileUpdated")
	user := "1234567890"
	userUpdated := "0987654321"
	pass := "test_api_key_1"
	passUpdated := "test_api_key_2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationFreemobileResourceConfig(
					name,
					user,
					pass,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_freemobile.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_freemobile.test",
						tfjsonpath.New("user"),
						knownvalue.StringExact(user),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_freemobile.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationFreemobileResourceConfig(
					nameUpdated,
					userUpdated,
					passUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_freemobile.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_freemobile.test",
						tfjsonpath.New("user"),
						knownvalue.StringExact(userUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_freemobile.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_freemobile.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"pass"},
			},
		},
	})
}

func testAccNotificationFreemobileResourceConfig(
	name string,
	user string,
	pass string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_freemobile" "test" {
  name      = %[1]q
  is_active = true
  user      = %[2]q
  pass      = %[3]q
}
`, name, user, pass)
}
