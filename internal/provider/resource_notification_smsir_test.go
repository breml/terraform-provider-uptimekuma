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

func TestAccNotificationSMSIRResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSIR")
	nameUpdated := acctest.RandomWithPrefix("NotificationSMSIRUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSIRResourceConfig(name, "test-api-key", "09123456789", "12345"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact("test-api-key"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("number"),
						knownvalue.StringExact("09123456789"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("template"),
						knownvalue.StringExact("12345"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccNotificationSMSIRResourceConfig(
					nameUpdated, "test-api-key-updated", "09123456780,09987654321", "54321",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("number"),
						knownvalue.StringExact("09123456780,09987654321"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsir.test",
						tfjsonpath.New("template"),
						knownvalue.StringExact("54321"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_smsir.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccNotificationSMSIRResourceConfig(
	name string, apiKey string, number string, template string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsir" "test" {
  name      = %[1]q
  is_active = true
  api_key   = %[2]q
  number    = %[3]q
  template  = %[4]q
}
`, name, apiKey, number, template)
}
