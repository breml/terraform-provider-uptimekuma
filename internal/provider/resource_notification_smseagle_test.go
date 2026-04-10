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

func TestAccNotificationSMSEagleResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSEagle")
	nameUpdated := acctest.RandomWithPrefix("NotificationSMSEagleUpdated")
	url := "https://192.168.1.100"
	urlUpdated := "https://smseagle.example.com"
	token := "test-api-token-123"
	tokenUpdated := "test-api-token-456"
	recipientTo := "+1234567890"
	recipientToUpdated := "+0987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSEagleResourceConfig(
					name,
					url,
					token,
					recipientTo,
					"smseagle-sms",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("token"),
						knownvalue.StringExact(token),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("recipient_type"),
						knownvalue.StringExact("smseagle-to"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("recipient_to"),
						knownvalue.StringExact(recipientTo),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("msg_type"),
						knownvalue.StringExact("smseagle-sms"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("api_type"),
						knownvalue.StringExact("smseagle-apiv2"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("priority"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("encoding"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("duration"),
						knownvalue.Int64Exact(10),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("tts_model"),
						knownvalue.Int64Exact(1),
					),
				},
			},
			{
				Config: testAccNotificationSMSEagleResourceConfig(
					nameUpdated,
					urlUpdated,
					tokenUpdated,
					recipientToUpdated,
					"smseagle-ring",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(urlUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("token"),
						knownvalue.StringExact(tokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("recipient_to"),
						knownvalue.StringExact(recipientToUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("msg_type"),
						knownvalue.StringExact("smseagle-ring"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smseagle.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_smseagle.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"url", "token"},
			},
		},
	})
}

func testAccNotificationSMSEagleResourceConfig(
	name string,
	url string,
	token string,
	recipientTo string,
	msgType string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smseagle" "test" {
  name           = %[1]q
  is_active      = true
  url            = %[2]q
  token          = %[3]q
  recipient_type = "smseagle-to"
  recipient_to   = %[4]q
  msg_type       = %[5]q
  api_type       = "smseagle-apiv2"
}
`, name, url, token, recipientTo, msgType)
}
