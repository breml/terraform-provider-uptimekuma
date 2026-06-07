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

func TestAccNotificationHaloPSAResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationHaloPSA")
	nameUpdated := acctest.RandomWithPrefix("NotificationHaloPSAUpdated")
	webhookURL := "https://halopsa.example.com/webhook/XXXXXXXX"
	webhookURLUpdated := "https://halopsa.example.com/webhook/YYYYYYYY"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationHaloPSAResourceConfig(name, webhookURL, "user1", "pass1"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_halopsa.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_halopsa.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_halopsa.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("user1"),
					),
				},
			},
			{
				Config: testAccNotificationHaloPSAResourceConfig(
					nameUpdated,
					webhookURLUpdated,
					"user2",
					"pass2",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_halopsa.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_halopsa.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("user2"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_halopsa.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"webhook_url", "password"},
			},
		},
	})
}

func testAccNotificationHaloPSAResourceConfig(
	name string, webhookURL string, username string, password string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_halopsa" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
  username    = %[3]q
  password    = %[4]q
}
`, name, webhookURL, username, password)
}
