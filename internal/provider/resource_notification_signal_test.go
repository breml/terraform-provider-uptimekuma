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

func TestAccNotificationSignalResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSignal")
	nameUpdated := acctest.RandomWithPrefix("NotificationSignalUpdated")
	url := "http://signal.example.com:8080"
	urlUpdated := "http://signal.example.com:8081"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSignalResourceConfig(
					name,
					url,
					"+1234567890",
					"+9876543210,+1111111111",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signal.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signal.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signal.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signal.test",
						tfjsonpath.New("number"),
						knownvalue.StringExact("+1234567890"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signal.test",
						tfjsonpath.New("recipients"),
						knownvalue.StringExact("+9876543210,+1111111111"),
					),
				},
			},
			{
				Config: testAccNotificationSignalResourceConfig(
					nameUpdated,
					urlUpdated,
					"+1111111111",
					"+9876543210",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signal.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signal.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(urlUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signal.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signal.test",
						tfjsonpath.New("number"),
						knownvalue.StringExact("+1111111111"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_signal.test",
						tfjsonpath.New("recipients"),
						knownvalue.StringExact("+9876543210"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_signal.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationSignalImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"url"},
			},
		},
	})
}

func testAccNotificationSignalImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_signal.test"]
	return rs.Primary.Attributes["id"], nil
}

func testAccNotificationSignalResourceConfig(
	name string, url string, number string, recipients string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_signal" "test" {
  name       = %[1]q
  is_active  = true
  url        = %[2]q
  number     = %[3]q
  recipients = %[4]q
}
`, name, url, number, recipients)
}
