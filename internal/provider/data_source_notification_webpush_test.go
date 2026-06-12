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

func TestAccNotificationWebpushDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationWebpush")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWebpushDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_webpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationWebpushDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_webpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationWebpushDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_webpush" "test" {
  name      = %[1]q
  is_active = true

  subscription = {
    endpoint = "https://fcm.googleapis.com/fcm/send/abc123"
    keys = {
      p256dh = "BGxi5eHcCnFv1example"
      auth   = "auth-secret-abc"
    }
  }
}

data "uptimekuma_notification_webpush" "test" {
  name = uptimekuma_notification_webpush.test.name
}
`, name)
}

func testAccNotificationWebpushDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_webpush" "test" {
  name      = %[1]q
  is_active = true

  subscription = {
    endpoint = "https://fcm.googleapis.com/fcm/send/abc123"
    keys = {
      p256dh = "BGxi5eHcCnFv1example"
      auth   = "auth-secret-abc"
    }
  }
}

data "uptimekuma_notification_webpush" "test" {
  id = uptimekuma_notification_webpush.test.id
}
`, name)
}
