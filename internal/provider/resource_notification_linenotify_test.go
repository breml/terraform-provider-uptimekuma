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

func TestAccNotificationLinenotifyResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationLinenotify")
	nameUpdated := acctest.RandomWithPrefix("NotificationLinenotifyUpdated")
	accessToken := "abcdef1234567890"
	accessTokenUpdated := "zyxwvu9876543210"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationLinenotifyResourceConfig(name, accessToken),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_linenotify.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_linenotify.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_linenotify.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationLinenotifyResourceConfig(nameUpdated, accessTokenUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_linenotify.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_linenotify.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_linenotify.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func TestAccNotificationLinenotifyResourceImport(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationLinenotify")
	accessToken := "abcdef1234567890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationLinenotifyResourceConfig(name, accessToken),
			},
			{
				ResourceName:            "uptimekuma_notification_linenotify.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationLinenotifyImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"access_token"},
			},
		},
	})
}

func testAccNotificationLinenotifyImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_linenotify.test"]
	return rs.Primary.Attributes["id"], nil
}

func testAccNotificationLinenotifyResourceConfig(name string, accessToken string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_linenotify" "test" {
  name           = %[1]q
  is_active      = true
  access_token   = %[2]q
}
`, name, accessToken)
}
