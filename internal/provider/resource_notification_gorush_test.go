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

func TestAccNotificationGorushResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGorush")
	nameUpdated := acctest.RandomWithPrefix("NotificationGorushUpdated")
	serverURL := "https://gorush.example.com"
	serverURLUpdated := "https://gorush-updated.example.com"
	deviceToken := "device-token-12345"
	deviceTokenUpdated := "device-token-67890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGorushResourceConfig(
					name,
					serverURL,
					deviceToken,
					"ios",
					"Test Title",
					"high",
					"3",
					"test-topic",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("server_url"),
						knownvalue.StringExact(serverURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("device_token"),
						knownvalue.StringExact(deviceToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("platform"),
						knownvalue.StringExact("ios"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Test Title"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("priority"),
						knownvalue.StringExact("high"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationGorushResourceConfig(
					nameUpdated,
					serverURLUpdated,
					deviceTokenUpdated,
					"android",
					"Updated Title",
					"low",
					"1",
					"updated-topic",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("server_url"),
						knownvalue.StringExact(serverURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("device_token"),
						knownvalue.StringExact(deviceTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("platform"),
						knownvalue.StringExact("android"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Updated Title"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("priority"),
						knownvalue.StringExact("low"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gorush.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccNotificationGorushResourceConfig(
	name string,
	serverURL string,
	deviceToken string,
	platform string,
	title string,
	priority string,
	retry string,
	topic string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_gorush" "test" {
  name           = %[1]q
  is_active      = true
  server_url     = %[2]q
  device_token   = %[3]q
  platform       = %[4]q
  title          = %[5]q
  priority       = %[6]q
  retry          = %[7]s
  topic          = %[8]q
}
`, name, serverURL, deviceToken, platform, title, priority, retry, topic)
}
