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

func TestAccNotificationGoAlertResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGoAlert")
	nameUpdated := acctest.RandomWithPrefix("NotificationGoAlertUpdated")
	baseURL := "https://goalert.example.com"
	baseURLUpdated := "https://goalert-updated.example.com"
	token := "test-token-123456789"
	tokenUpdated := "updated-token-987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGoAlertResourceConfig(
					name,
					baseURL,
					token,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_goalert.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_goalert.test",
						tfjsonpath.New("base_url"),
						knownvalue.StringExact(baseURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_goalert.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationGoAlertResourceConfig(
					nameUpdated,
					baseURLUpdated,
					tokenUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_goalert.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_goalert.test",
						tfjsonpath.New("base_url"),
						knownvalue.StringExact(baseURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_goalert.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccNotificationGoAlertResourceConfig(
	name string, baseURL string, token string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_goalert" "test" {
  name     = %[1]q
  is_active = true
  base_url = %[2]q
  token    = %[3]q
}
`, name, baseURL, token)
}
