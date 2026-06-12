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

func TestAccNotificationTeltonikaResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTeltonika")
	nameUpdated := acctest.RandomWithPrefix("NotificationTeltonikaUpdated")
	url := "https://192.168.1.1"
	urlUpdated := "https://teltonika.example.com"
	username := "admin"
	usernameUpdated := "admin2"
	password := "test-password-123"
	passwordUpdated := "test-password-456"
	phoneNumber := "+33600000000"
	phoneNumberUpdated := "+33600000001"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Minimal config: verify optional fields default correctly.
			{
				Config: testAccNotificationTeltonikaResourceConfigMinimal(
					name, url, username, password, phoneNumber,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(password),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact(phoneNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("modem"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("unsafe_tls"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			// Full config: set all optional fields.
			{
				Config: testAccNotificationTeltonikaResourceConfig(
					name, url, username, password, "1-1", phoneNumber, true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("modem"),
						knownvalue.StringExact("1-1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("unsafe_tls"),
						knownvalue.Bool(true),
					),
				},
			},
			// Update: change all fields.
			{
				Config: testAccNotificationTeltonikaResourceConfig(
					nameUpdated, urlUpdated, usernameUpdated, passwordUpdated, "1-2", phoneNumberUpdated, false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(urlUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(usernameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(passwordUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("modem"),
						knownvalue.StringExact("1-2"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact(phoneNumberUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_teltonika.test",
						tfjsonpath.New("unsafe_tls"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_teltonika.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccNotificationTeltonikaResourceConfigMinimal(
	name string, url string, username string, password string, phoneNumber string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_teltonika" "test" {
  name         = %[1]q
  is_active    = true
  url          = %[2]q
  username     = %[3]q
  password     = %[4]q
  phone_number = %[5]q
}
`, name, url, username, password, phoneNumber)
}

func testAccNotificationTeltonikaResourceConfig(
	name string, url string, username string, password string, modem string, phoneNumber string, unsafeTLS bool,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_teltonika" "test" {
  name         = %[1]q
  is_active    = true
  url          = %[2]q
  username     = %[3]q
  password     = %[4]q
  modem        = %[5]q
  phone_number = %[6]q
  unsafe_tls   = %[7]t
}
`, name, url, username, password, modem, phoneNumber, unsafeTLS)
}
