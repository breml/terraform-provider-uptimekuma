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

func TestAccNotificationOnesenderResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOnesender")
	nameUpdated := acctest.RandomWithPrefix("NotificationOnesenderUpdated")
	url := "https://onesender.example.com/api/v1/send"
	urlUpdated := "https://onesender.example.com/api/v2/send"
	token := "test-token-abc123"
	tokenUpdated := "test-token-def456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOnesenderResourceConfig(
					name,
					url,
					token,
					"+6281234567890",
					"private",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("token"),
						knownvalue.StringExact(token),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("receiver"),
						knownvalue.StringExact("+6281234567890"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("type_receiver"),
						knownvalue.StringExact("private"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationOnesenderResourceConfig(
					nameUpdated,
					urlUpdated,
					tokenUpdated,
					"group-id-123",
					"group",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(urlUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("token"),
						knownvalue.StringExact(tokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("receiver"),
						knownvalue.StringExact("group-id-123"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("type_receiver"),
						knownvalue.StringExact("group"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onesender.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_onesender.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationOnesenderResourceConfig(
	name string,
	url string,
	token string,
	receiver string,
	typeReceiver string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_onesender" "test" {
  name          = %[1]q
  is_active     = true
  url           = %[2]q
  token         = %[3]q
  receiver      = %[4]q
  type_receiver = %[5]q
}
`, name, url, token, receiver, typeReceiver)
}
