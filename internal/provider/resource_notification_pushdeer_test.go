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

func TestAccNotificationPushDeerResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPushDeer")
	nameUpdated := acctest.RandomWithPrefix("NotificationPushDeerUpdated")
	key := "pushkey123"
	keyUpdated := "pushkey456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPushDeerResourceConfig(name, key, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushdeer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushdeer.test",
						tfjsonpath.New("key"),
						knownvalue.StringExact(key),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushdeer.test",
						tfjsonpath.New("server"),
						knownvalue.StringExact("https://api2.pushdeer.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushdeer.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationPushDeerResourceConfig(
					nameUpdated,
					keyUpdated,
					"https://custom.pushdeer.server.com",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushdeer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushdeer.test",
						tfjsonpath.New("key"),
						knownvalue.StringExact(keyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushdeer.test",
						tfjsonpath.New("server"),
						knownvalue.StringExact("https://custom.pushdeer.server.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushdeer.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_pushdeer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationPushDeerResourceConfig(name string, key string, server string) string {
	baseConfig := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pushdeer" "test" {
  name      = %[1]q
  is_active = true
  key       = %[2]q
`, name, key)

	if server != "" {
		baseConfig += fmt.Sprintf("  server = %[1]q\n", server)
	}

	baseConfig += "}\n"
	return baseConfig
}
