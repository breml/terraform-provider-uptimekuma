package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccNotificationMilkyResource covers create, update and import, and both shapes of msg_type:
// unset, which the server reads as private, and group. The updated http_addr ends in a slash, which
// has to round-trip verbatim. Going back to an unset msg_type checks that the key is dropped rather
// than kept by the server, which would show up as a non-empty plan.
func TestAccNotificationMilkyResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationMilky")
	nameUpdated := acctest.RandomWithPrefix("NotificationMilkyUpdated")
	addr := "uptimekuma_notification_milky.test"
	httpAddr := "http://localhost:3000"
	httpAddrUpdated := "http://milky.example.com:3000/"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationMilkyResourceConfig(name, httpAddr, "token-1", `"channel"`, "1"),
				ExpectError: regexp.MustCompile(`(?i)msg_type`),
			},
			{
				Config: testAccNotificationMilkyResourceConfig(name, httpAddr, "token-1", "null", "1001"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("is_default"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("apply_existing"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("http_addr"), knownvalue.StringExact(httpAddr)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("access_token"), knownvalue.StringExact("token-1")),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("msg_type"), knownvalue.Null()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("receiver_id"), knownvalue.StringExact("1001")),
				},
			},
			{
				Config: testAccNotificationMilkyResourceConfig(nameUpdated, httpAddrUpdated, "token-2", `"group"`, "2002"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("http_addr"), knownvalue.StringExact(httpAddrUpdated)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("access_token"), knownvalue.StringExact("token-2")),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("msg_type"), knownvalue.StringExact("group")),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("receiver_id"), knownvalue.StringExact("2002")),
				},
			},
			{
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationMilkyResourceConfig(nameUpdated, httpAddr, "token-2", "null", "2002"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("msg_type"), knownvalue.Null()),
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

// testAccNotificationMilkyResourceConfig takes msgType as an HCL expression, so that null leaves the
// attribute unset.
func testAccNotificationMilkyResourceConfig(
	name string,
	httpAddr string,
	accessToken string,
	msgType string,
	receiverID string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_milky" "test" {
  name         = %[1]q
  is_active    = true
  http_addr    = %[2]q
  access_token = %[3]q
  msg_type     = %[4]s
  receiver_id  = %[5]q
}
`, name, httpAddr, accessToken, msgType, receiverID)
}
