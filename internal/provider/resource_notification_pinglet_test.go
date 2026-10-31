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

// TestAccNotificationPingletResource covers create, update and import. The updated publish_url ends
// in a slash, which Uptime Kuma strips only at send time, so it has to round-trip verbatim.
func TestAccNotificationPingletResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPinglet")
	nameUpdated := acctest.RandomWithPrefix("NotificationPingletUpdated")
	addr := "uptimekuma_notification_pinglet.test"
	publishURL := "https://app.pinglet.co.uk/test-namespace/alerts"
	publishURLUpdated := "https://app.pinglet.co.uk/test-namespace/critical/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPingletResourceConfig(name, publishURL, "api-key-1"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("is_default"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("apply_existing"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("publish_url"), knownvalue.StringExact(publishURL)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("api_key"), knownvalue.StringExact("api-key-1")),
				},
			},
			{
				Config: testAccNotificationPingletResourceConfig(nameUpdated, publishURLUpdated, "api-key-2"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("publish_url"), knownvalue.StringExact(publishURLUpdated),
					),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("api_key"), knownvalue.StringExact("api-key-2")),
				},
			},
			{
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationPingletResourceConfig(name string, publishURL string, apiKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pinglet" "test" {
  name        = %[1]q
  is_active   = true
  publish_url = %[2]q
  api_key     = %[3]q
}
`, name, publishURL, apiKey)
}
