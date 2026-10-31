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

// TestAccNotificationBearSMSResource covers the default shape, where sender_id stays unset and must
// therefore round-trip as null rather than as the empty string.
func TestAccNotificationBearSMSResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationBearSMS")
	nameUpdated := acctest.RandomWithPrefix("NotificationBearSMSUpdated")
	addr := "uptimekuma_notification_bearsms.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBearSMSResourceConfig(name, "user-1", "hash-key-1", "41790000001"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("is_default"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("apply_existing"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("username"), knownvalue.StringExact("user-1")),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("hash_key"), knownvalue.StringExact("hash-key-1")),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("phone_number"), knownvalue.StringExact("41790000001"),
					),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("sender_id"), knownvalue.Null()),
				},
			},
			{
				Config: testAccNotificationBearSMSResourceConfig(nameUpdated, "user-2", "hash-key-2", "9725000000"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("username"), knownvalue.StringExact("user-2")),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("hash_key"), knownvalue.StringExact("hash-key-2")),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("phone_number"), knownvalue.StringExact("9725000000"),
					),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("sender_id"), knownvalue.Null()),
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

// TestAccNotificationBearSMSResourceSenderID covers sender_id set, changed, imported and cleared
// again, which has to bring it back to null.
func TestAccNotificationBearSMSResourceSenderID(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationBearSMSSender")
	addr := "uptimekuma_notification_bearsms.sender"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBearSMSResourceSenderConfig(name, "UptimeKuma"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("sender_id"), knownvalue.StringExact("UptimeKuma")),
				},
			},
			{
				Config: testAccNotificationBearSMSResourceSenderConfig(name, "UptimeKuma1"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("sender_id"), knownvalue.StringExact("UptimeKuma1")),
				},
			},
			{
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationBearSMSResourceSenderClearedConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("sender_id"), knownvalue.Null()),
				},
			},
		},
	})
}

// TestAccNotificationBearSMSResourceSenderIDTooLong covers the sender_id length validator.
func TestAccNotificationBearSMSResourceSenderIDTooLong(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationBearSMSInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationBearSMSResourceSenderConfig(name, "TwelveChars!"),
				ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Length.*sender_id`),
			},
		},
	})
}

func testAccNotificationBearSMSResourceConfig(
	name string, username string, hashKey string, phoneNumber string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bearsms" "test" {
  name         = %[1]q
  is_active    = true
  username     = %[2]q
  hash_key     = %[3]q
  phone_number = %[4]q
}
`, name, username, hashKey, phoneNumber)
}

func testAccNotificationBearSMSResourceSenderConfig(name string, senderID string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bearsms" "sender" {
  name         = %[1]q
  is_active    = true
  username     = "bearsms-user"
  hash_key     = "bearsms-hash-key"
  sender_id    = %[2]q
  phone_number = "41790000000"
}
`, name, senderID)
}

func testAccNotificationBearSMSResourceSenderClearedConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bearsms" "sender" {
  name         = %[1]q
  is_active    = true
  username     = "bearsms-user"
  hash_key     = "bearsms-hash-key"
  phone_number = "41790000000"
}
`, name)
}
