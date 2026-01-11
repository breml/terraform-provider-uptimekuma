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

func TestAccNotificationNostrDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNostrDS")
	sender := "nsec1vl5dr5z69cydfy5kqsruqh84zkyv684jpu7v0unv2dl0aq2t5spe5nqq"
	recipients := "npub1qypt4l5elx7qjxapqvzc3gw7nj5zxwq5r5rzc5yqgj5j5j5j5j5j5j5j5j5j"
	relays := "wss://relay.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNostrDataSourceConfigByName(name, sender, recipients, relays),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_nostr.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationNostrDataSourceConfigByID(name, sender, recipients, relays),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_nostr.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationNostrDataSourceConfigByName(
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

data "uptimekuma_notification_nostr" "by_name" {
  name = uptimekuma_notification_nostr.test.name
}
`, name, sender, recipients, relays)
}

func testAccNotificationNostrDataSourceConfigByID(
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

data "uptimekuma_notification_nostr" "by_id" {
  id = uptimekuma_notification_nostr.test.id
}
`, name, sender, recipients, relays)
}
