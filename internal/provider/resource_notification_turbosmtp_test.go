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

// TestAccNotificationTurboSMTPResource covers the default shape, where region and the optional
// fields stay unset and must therefore round-trip as null rather than as the empty string.
func TestAccNotificationTurboSMTPResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTurboSMTP")
	nameUpdated := acctest.RandomWithPrefix("NotificationTurboSMTPUpdated")
	addr := "uptimekuma_notification_turbosmtp.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTurboSMTPResourceConfig(
					name,
					"consumer-key-1",
					"consumer-secret-1",
					"alerts@example.com",
					"oncall@example.com",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("is_default"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("apply_existing"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("consumer_key"), knownvalue.StringExact("consumer-key-1"),
					),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("consumer_secret"), knownvalue.StringExact("consumer-secret-1"),
					),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("from_email"), knownvalue.StringExact("alerts@example.com"),
					),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("to_email"), knownvalue.StringExact("oncall@example.com"),
					),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("region"), knownvalue.Null()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("cc_email"), knownvalue.Null()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("bcc_email"), knownvalue.Null()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("subject"), knownvalue.Null()),
				},
			},
			{
				Config: testAccNotificationTurboSMTPResourceConfig(
					nameUpdated,
					"consumer-key-2",
					"consumer-secret-2",
					"monitoring@example.com",
					"oncall@example.com,ops@example.com",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("consumer_key"), knownvalue.StringExact("consumer-key-2"),
					),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("consumer_secret"), knownvalue.StringExact("consumer-secret-2"),
					),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("from_email"), knownvalue.StringExact("monitoring@example.com"),
					),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("to_email"), knownvalue.StringExact("oncall@example.com,ops@example.com"),
					),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("region"), knownvalue.Null()),
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

// TestAccNotificationTurboSMTPResourceEU covers the EU region with every optional field set, and
// clearing them again, which has to bring each back to null.
func TestAccNotificationTurboSMTPResourceEU(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTurboSMTPEU")
	addr := "uptimekuma_notification_turbosmtp.eu"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTurboSMTPResourceEUConfig(
					name, "cc@example.com", "bcc@example.com", "Uptime Kuma alert",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("region"), knownvalue.StringExact("eu")),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("cc_email"), knownvalue.StringExact("cc@example.com"),
					),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("bcc_email"), knownvalue.StringExact("bcc@example.com"),
					),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("subject"), knownvalue.StringExact("Uptime Kuma alert"),
					),
				},
			},
			{
				Config: testAccNotificationTurboSMTPResourceEUConfig(
					name, "cc1@example.com, cc2@example.com", "bcc2@example.com", "Service down",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("cc_email"), knownvalue.StringExact("cc1@example.com, cc2@example.com"),
					),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("bcc_email"), knownvalue.StringExact("bcc2@example.com"),
					),
					statecheck.ExpectKnownValue(
						addr, tfjsonpath.New("subject"), knownvalue.StringExact("Service down"),
					),
				},
			},
			{
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationTurboSMTPResourceClearedConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("region"), knownvalue.Null()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("cc_email"), knownvalue.Null()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("bcc_email"), knownvalue.Null()),
					statecheck.ExpectKnownValue(addr, tfjsonpath.New("subject"), knownvalue.Null()),
				},
			},
		},
	})
}

// TestAccNotificationTurboSMTPResourceInvalidRegion covers the region validator.
func TestAccNotificationTurboSMTPResourceInvalidRegion(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTurboSMTPInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationTurboSMTPResourceInvalidRegionConfig(name),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
			},
		},
	})
}

func testAccNotificationTurboSMTPResourceConfig(
	name string, consumerKey string, consumerSecret string, fromEmail string, toEmail string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_turbosmtp" "test" {
  name            = %[1]q
  consumer_key    = %[2]q
  consumer_secret = %[3]q
  from_email      = %[4]q
  to_email        = %[5]q
}
`, name, consumerKey, consumerSecret, fromEmail, toEmail)
}

func testAccNotificationTurboSMTPResourceEUConfig(
	name string, ccEmail string, bccEmail string, subject string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_turbosmtp" "eu" {
  name            = %[1]q
  consumer_key    = "consumer-key-eu"
  consumer_secret = "consumer-secret-eu"
  region          = "eu"
  from_email      = "alerts@example.eu"
  to_email        = "oncall@example.eu"
  cc_email        = %[2]q
  bcc_email       = %[3]q
  subject         = %[4]q
}
`, name, ccEmail, bccEmail, subject)
}

func testAccNotificationTurboSMTPResourceClearedConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_turbosmtp" "eu" {
  name            = %[1]q
  consumer_key    = "consumer-key-eu"
  consumer_secret = "consumer-secret-eu"
  from_email      = "alerts@example.eu"
  to_email        = "oncall@example.eu"
}
`, name)
}

func testAccNotificationTurboSMTPResourceInvalidRegionConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_turbosmtp" "invalid" {
  name            = %[1]q
  consumer_key    = "consumer-key"
  consumer_secret = "consumer-secret"
  region          = "asia"
  from_email      = "alerts@example.com"
  to_email        = "oncall@example.com"
}
`, name)
}
