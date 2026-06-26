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

func TestAccNotificationWebpushResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWebpush")
	nameUpdated := acctest.RandomWithPrefix("NotificationWebpushUpdated")
	pushEndpoint := "https://fcm.googleapis.com/fcm/send/abc123"
	pushEndpointUpdated := "https://fcm.googleapis.com/fcm/send/xyz789"
	p256dh := "BGxi5eHcCnFv1example"
	p256dhUpdated := "BNcExampleUpdatedKey"
	auth := "auth-secret-abc"
	authUpdated := "auth-secret-xyz"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWebpushResourceConfig(name, pushEndpoint, p256dh, auth),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webpush.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webpush.test",
						tfjsonpath.New("subscription").AtMapKey("endpoint"),
						knownvalue.StringExact(pushEndpoint),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webpush.test",
						tfjsonpath.New("subscription").AtMapKey("keys").AtMapKey("p256dh"),
						knownvalue.StringExact(p256dh),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webpush.test",
						tfjsonpath.New("subscription").AtMapKey("keys").AtMapKey("auth"),
						knownvalue.StringExact(auth),
					),
				},
			},
			{
				Config: testAccNotificationWebpushResourceConfig(nameUpdated, pushEndpointUpdated, p256dhUpdated, authUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webpush.test",
						tfjsonpath.New("subscription").AtMapKey("endpoint"),
						knownvalue.StringExact(pushEndpointUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webpush.test",
						tfjsonpath.New("subscription").AtMapKey("keys").AtMapKey("p256dh"),
						knownvalue.StringExact(p256dhUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_webpush.test",
						tfjsonpath.New("subscription").AtMapKey("keys").AtMapKey("auth"),
						knownvalue.StringExact(authUpdated),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_webpush.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationWebpushResourceConfig(name string, endpoint string, p256dh string, auth string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_webpush" "test" {
  name      = %[1]q
  is_active = true

  subscription = {
    endpoint = %[2]q
    keys = {
      p256dh = %[3]q
      auth   = %[4]q
    }
  }
}
`, name, endpoint, p256dh, auth)
}
