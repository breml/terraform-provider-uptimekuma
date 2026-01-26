package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationServerChanResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationServerChan")
	nameUpdated := acctest.RandomWithPrefix("NotificationServerChanUpdated")
	sendKey := "test-send-key-12345"
	sendKeyUpdated := "test-send-key-67890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationServerChanResourceConfig(name, sendKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serverchan.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serverchan.test",
						tfjsonpath.New("send_key"),
						knownvalue.StringExact(sendKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serverchan.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationServerChanResourceConfig(nameUpdated, sendKeyUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serverchan.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serverchan.test",
						tfjsonpath.New("send_key"),
						knownvalue.StringExact(sendKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serverchan.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_serverchan.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNotificationServerChanImportStateID,
			},
		},
	})
}

func testAccNotificationServerChanResourceConfig(name string, sendKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_serverchan" "test" {
  name     = %[1]q
  is_active = true
  send_key = %[2]q
}
`, name, sendKey)
}

func testAccNotificationServerChanImportStateID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["uptimekuma_notification_serverchan.test"]
	if !ok {
		return "", errors.New("Not found: uptimekuma_notification_serverchan.test")
	}

	return rs.Primary.Attributes["id"], nil
}
