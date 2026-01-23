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

func TestAccNotificationPushPlusResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPushPlus")
	nameUpdated := acctest.RandomWithPrefix("NotificationPushPlusUpdated")
	sendKey := "test-send-key-123456789"
	sendKeyUpdated := "updated-send-key-987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPushPlusResourceConfig(name, sendKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushplus.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushplus.test",
						tfjsonpath.New("send_key"),
						knownvalue.StringExact(sendKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushplus.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationPushPlusResourceConfig(nameUpdated, sendKeyUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushplus.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushplus.test",
						tfjsonpath.New("send_key"),
						knownvalue.StringExact(sendKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushplus.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func TestAccNotificationPushPlusResourceImport(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPushPlus")
	sendKey := "test-send-key-import-123456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPushPlusResourceConfig(name, sendKey),
			},
			{
				ResourceName:      "uptimekuma_notification_pushplus.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationPushPlusResourceConfig(name string, sendKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pushplus" "test" {
  name     = %[1]q
  is_active = true
  send_key = %[2]q
}
`, name, sendKey)
}
