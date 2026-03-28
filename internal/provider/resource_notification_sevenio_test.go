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

func TestAccNotificationSevenioResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSevenio")
	nameUpdated := acctest.RandomWithPrefix("NotificationSevenioUpdated")
	apiKey := "test-api-key-123"
	apiKeyUpdated := "test-api-key-456"
	sender := "+491234567890"
	senderUpdated := "+490987654321"
	to := "+491111111111"
	toUpdated := "+492222222222"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSevenioResourceConfig(
					name,
					apiKey,
					sender,
					to,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sevenio.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sevenio.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sevenio.test",
						tfjsonpath.New("sender"),
						knownvalue.StringExact(sender),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sevenio.test",
						tfjsonpath.New("to"),
						knownvalue.StringExact(to),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sevenio.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationSevenioResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					senderUpdated,
					toUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sevenio.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sevenio.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sevenio.test",
						tfjsonpath.New("sender"),
						knownvalue.StringExact(senderUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sevenio.test",
						tfjsonpath.New("to"),
						knownvalue.StringExact(toUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sevenio.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_sevenio.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccNotificationSevenioResourceConfig(
	name string, apiKey string, sender string, to string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_sevenio" "test" {
  name      = %[1]q
  is_active = true
  api_key   = %[2]q
  sender    = %[3]q
  to        = %[4]q
}
`, name, apiKey, sender, to)
}
