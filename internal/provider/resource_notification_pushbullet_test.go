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

func TestAccNotificationPushbulletResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPushbullet")
	nameUpdated := acctest.RandomWithPrefix("NotificationPushbulletUpdated")
	accessToken := "o.test1234567890abcdefghijklmnopqrst"
	accessTokenUpdated := "o.updated1234567890abcdefghijklmnop"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPushbulletResourceConfig(name, accessToken),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushbullet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushbullet.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushbullet.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationPushbulletResourceConfig(nameUpdated, accessTokenUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushbullet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushbullet.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushbullet.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_pushbullet.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationPushbulletImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"access_token"},
			},
		},
	})
}

func testAccNotificationPushbulletResourceConfig(name string, accessToken string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pushbullet" "test" {
  name          = %[1]q
  is_active     = true
  access_token  = %[2]q
}
`, name, accessToken)
}

func testAccNotificationPushbulletImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_pushbullet.test"]
	return rs.Primary.Attributes["id"], nil
}
