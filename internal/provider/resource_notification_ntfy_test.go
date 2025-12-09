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

func TestAccNotificationNtfyResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNtfy")
	nameUpdated := acctest.RandomWithPrefix("NotificationNtfyUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNotificationNtfyResourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("authentication_method"), knownvalue.StringExact("none")),
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("server_url"), knownvalue.StringExact("https://ntfy.sh")),
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("priority"), knownvalue.Int32Exact(5)),
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("topic"), knownvalue.StringExact(name)),
				},
			},
			// Update and Read testing
			{
				Config: testAccNotificationNtfyResourceConfig(nameUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("authentication_method"), knownvalue.StringExact("none")),
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("server_url"), knownvalue.StringExact("https://ntfy.sh")),
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("priority"), knownvalue.Int32Exact(5)),
					statecheck.ExpectKnownValue("uptimekuma_notification_ntfy.test", tfjsonpath.New("topic"), knownvalue.StringExact(nameUpdated)),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccNotificationNtfyResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ntfy" "test" {
  name      = %[1]q
  is_active = true

  authentication_method = "none"
  server_url            = "https://ntfy.sh"
  priority              = 5
  topic                 = %[1]q
}
`, name)
}
