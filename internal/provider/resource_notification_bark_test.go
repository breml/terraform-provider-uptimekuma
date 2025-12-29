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

func TestAccNotificationBarkResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationBark")
	nameUpdated := acctest.RandomWithPrefix("NotificationBarkUpdated")
	endpointURL := "https://api.bark.com"
	endpointURLUpdated := "https://api-updated.bark.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBarkResourceConfig(
					name,
					endpointURL,
					"test-group",
					"default",
					"v1",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("endpoint"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("test-group"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("sound"),
						knownvalue.StringExact("default"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("api_version"),
						knownvalue.StringExact("v1"),
					),
				},
			},
			{
				Config: testAccNotificationBarkResourceConfig(
					nameUpdated,
					endpointURLUpdated,
					"updated-group",
					"custom",
					"v2",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("endpoint"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("updated-group"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("sound"),
						knownvalue.StringExact("custom"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bark.test",
						tfjsonpath.New("api_version"),
						knownvalue.StringExact("v2"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_bark.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationBarkResourceConfig(
	name string,
	endpoint string,
	group string,
	sound string,
	apiVersion string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bark" "test" {
  name        = %[1]q
  is_active   = true
  endpoint    = %[2]q
  group       = %[3]q
  sound       = %[4]q
  api_version = %[5]q
}
`, name, endpoint, group, sound, apiVersion)
}
