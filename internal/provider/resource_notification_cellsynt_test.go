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

func TestAccNotificationCellsyntResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationCellsynt")
	nameUpdated := acctest.RandomWithPrefix("NotificationCellsyntUpdated")
	login := "testuser"
	loginUpdated := "testuserupdated"
	password := "testpass123"
	passwordUpdated := "testpass456"
	destination := "+46701234567"
	destinationUpdated := "+46707654321"
	originator := "TestSender"
	originatorUpdated := "UpdatedSender"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationCellsyntResourceConfig(
					name,
					login,
					password,
					destination,
					originator,
					"Numeric",
					false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("login"),
						knownvalue.StringExact(login),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(password),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("destination"),
						knownvalue.StringExact(destination),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("originator"),
						knownvalue.StringExact(originator),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("originator_type"),
						knownvalue.StringExact("Numeric"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("allow_long_sms"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationCellsyntResourceConfig(
					nameUpdated,
					loginUpdated,
					passwordUpdated,
					destinationUpdated,
					originatorUpdated,
					"Alphanumeric",
					true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("login"),
						knownvalue.StringExact(loginUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(passwordUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("destination"),
						knownvalue.StringExact(destinationUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("originator"),
						knownvalue.StringExact(originatorUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("originator_type"),
						knownvalue.StringExact("Alphanumeric"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("allow_long_sms"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_cellsynt.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_cellsynt.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccNotificationCellsyntResourceConfig(
	name string, login string, password string, destination string, originator string,
	originatorType string, allowLongSMS bool,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_cellsynt" "test" {
  name              = %[1]q
  is_active         = true
  login             = %[2]q
  password          = %[3]q
  destination       = %[4]q
  originator        = %[5]q
  originator_type   = %[6]q
  allow_long_sms    = %[7]t
}
`, name, login, password, destination, originator, originatorType, allowLongSMS)
}
