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

// TestAccNotificationPlivoResource covers the SMS shape, where message_type and answer_url stay
// unset and must therefore round-trip as null rather than as the empty string.
func TestAccNotificationPlivoResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPlivo")
	nameUpdated := acctest.RandomWithPrefix("NotificationPlivoUpdated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPlivoResourceConfig(
					name,
					"MAXXXXXXXXXXXXXXXXXX",
					"authtoken0000000000000000000000000000000",
					"+15550001111",
					"+15550002222",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("auth_id"),
						knownvalue.StringExact("MAXXXXXXXXXXXXXXXXXX"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact("authtoken0000000000000000000000000000000"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("from_number"),
						knownvalue.StringExact("+15550001111"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact("+15550002222"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("message_type"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("answer_url"),
						knownvalue.Null(),
					),
				},
			},
			{
				Config: testAccNotificationPlivoResourceConfig(
					nameUpdated,
					"MAYYYYYYYYYYYYYYYYYY",
					"authtoken1111111111111111111111111111111",
					"+15550003333",
					"+15550004444",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("auth_id"),
						knownvalue.StringExact("MAYYYYYYYYYYYYYYYYYY"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact("authtoken1111111111111111111111111111111"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("from_number"),
						knownvalue.StringExact("+15550003333"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact("+15550004444"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("message_type"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.test",
						tfjsonpath.New("answer_url"),
						knownvalue.Null(),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_plivo.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationPlivoResourceCall covers the call shape, where answer_url is mandatory, and
// the switch back to an explicit SMS, which has to clear answer_url again.
func TestAccNotificationPlivoResourceCall(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPlivoCall")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPlivoResourceCallConfig(name, "https://example.com/answer"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.call",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.call",
						tfjsonpath.New("message_type"),
						knownvalue.StringExact("call"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.call",
						tfjsonpath.New("answer_url"),
						knownvalue.StringExact("https://example.com/answer"),
					),
				},
			},
			{
				Config: testAccNotificationPlivoResourceCallConfig(name, "https://example.com/other-answer"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.call",
						tfjsonpath.New("answer_url"),
						knownvalue.StringExact("https://example.com/other-answer"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_plivo.call",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationPlivoResourceSMSConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.call",
						tfjsonpath.New("message_type"),
						knownvalue.StringExact("sms"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_plivo.call",
						tfjsonpath.New("answer_url"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

// TestAccNotificationPlivoResourceInvalidConfig covers the configuration validation that
// stringvalidator cannot express, because both halves depend on the value of message_type.
func TestAccNotificationPlivoResourceInvalidConfig(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPlivoInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNotificationPlivoResourceCallWithoutAnswerURLConfig(name),
				ExpectError: regexp.MustCompile(`Missing Attribute Configuration`),
			},
			{
				Config:      testAccNotificationPlivoResourceSMSWithAnswerURLConfig(name),
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
			{
				Config:      testAccNotificationPlivoResourceInvalidMessageTypeConfig(name),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
			},
			{
				Config:      testAccNotificationPlivoResourceInvalidAnswerURLConfig(name),
				ExpectError: regexp.MustCompile(`Invalid URL`),
			},
		},
	})
}

func testAccNotificationPlivoResourceConfig(
	name string, authID string, authToken string, fromNumber string, toNumber string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_plivo" "test" {
  name        = %[1]q
  is_active   = true
  auth_id     = %[2]q
  auth_token  = %[3]q
  from_number = %[4]q
  to_number   = %[5]q
}
`, name, authID, authToken, fromNumber, toNumber)
}

func testAccNotificationPlivoResourceCallConfig(name string, answerURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_plivo" "call" {
  name         = %[1]q
  is_active    = true
  auth_id      = "MAXXXXXXXXXXXXXXXXXX"
  auth_token   = "authtoken0000000000000000000000000000000"
  from_number  = "+15550001111"
  to_number    = "+15550002222"
  message_type = "call"
  answer_url   = %[2]q
}
`, name, answerURL)
}

func testAccNotificationPlivoResourceSMSConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_plivo" "call" {
  name         = %[1]q
  is_active    = true
  auth_id      = "MAXXXXXXXXXXXXXXXXXX"
  auth_token   = "authtoken0000000000000000000000000000000"
  from_number  = "+15550001111"
  to_number    = "+15550002222"
  message_type = "sms"
}
`, name)
}

func testAccNotificationPlivoResourceCallWithoutAnswerURLConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_plivo" "invalid" {
  name         = %[1]q
  auth_id      = "MAXXXXXXXXXXXXXXXXXX"
  auth_token   = "authtoken0000000000000000000000000000000"
  from_number  = "+15550001111"
  to_number    = "+15550002222"
  message_type = "call"
}
`, name)
}

func testAccNotificationPlivoResourceSMSWithAnswerURLConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_plivo" "invalid" {
  name        = %[1]q
  auth_id     = "MAXXXXXXXXXXXXXXXXXX"
  auth_token  = "authtoken0000000000000000000000000000000"
  from_number = "+15550001111"
  to_number   = "+15550002222"
  answer_url  = "https://example.com/answer"
}
`, name)
}

func testAccNotificationPlivoResourceInvalidMessageTypeConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_plivo" "invalid" {
  name         = %[1]q
  auth_id      = "MAXXXXXXXXXXXXXXXXXX"
  auth_token   = "authtoken0000000000000000000000000000000"
  from_number  = "+15550001111"
  to_number    = "+15550002222"
  message_type = "fax"
}
`, name)
}

func testAccNotificationPlivoResourceInvalidAnswerURLConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_plivo" "invalid" {
  name         = %[1]q
  auth_id      = "MAXXXXXXXXXXXXXXXXXX"
  auth_token   = "authtoken0000000000000000000000000000000"
  from_number  = "+15550001111"
  to_number    = "+15550002222"
  message_type = "call"
  answer_url   = "not-a-url"
}
`, name)
}
