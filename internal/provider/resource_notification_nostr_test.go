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

func TestAccNotificationNostrResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNostr")
	nameUpdated := acctest.RandomWithPrefix("NotificationNostrUpdated")
	sender := "nsec1vl5dr5z69cydfy5kqsruqh84zkyv684jpu7v0unv2dl0aq2t5spe5nqq"
	senderUpdated := "nsec1uy5dr5z69cydfy5kqsruqh84zkyv684jpu7v0unv2dl0aq2t5spe6nqq"
	recipients := "npub1qypt4l5elx7qjxapqvzc3gw7nj5zxwq5r5rzc5yqgj5j5j5j5j5j5j5j5j5j"
	recipientsUpdated := "npub2qypt4l5elx7qjxapqvzc3gw7nj5zxwq5r5rzc5yqgj5j5j5j5j5j5j5j5j"
	relays := "wss://relay.example.com\nwss://relay2.example.com"
	relaysUpdated := "wss://relay-updated.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNostrResourceConfig(
					name,
					sender,
					recipients,
					relays,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nostr.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nostr.test",
						tfjsonpath.New("sender"),
						knownvalue.StringExact(sender),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nostr.test",
						tfjsonpath.New("recipients"),
						knownvalue.StringExact(recipients),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nostr.test",
						tfjsonpath.New("relays"),
						knownvalue.StringExact(relays),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nostr.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationNostrResourceConfig(
					nameUpdated,
					senderUpdated,
					recipientsUpdated,
					relaysUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nostr.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nostr.test",
						tfjsonpath.New("sender"),
						knownvalue.StringExact(senderUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nostr.test",
						tfjsonpath.New("recipients"),
						knownvalue.StringExact(recipientsUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nostr.test",
						tfjsonpath.New("relays"),
						knownvalue.StringExact(relaysUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nostr.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_nostr.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationNostrResourceConfig(
	name string, sender string, recipients string, relays string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_nostr" "test" {
  name       = %[1]q
  is_active  = true
  sender     = %[2]q
  recipients = %[3]q
  relays     = %[4]q
}
`, name, sender, recipients, relays)
}
