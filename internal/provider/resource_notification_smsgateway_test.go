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

// TestAccNotificationSMSGatewayResource covers create, an update of every attribute and an import
// round-trip. The recipients list keeps its irregular whitespace to assert it is stored verbatim.
func TestAccNotificationSMSGatewayResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSGateway")
	nameUpdated := acctest.RandomWithPrefix("NotificationSMSGatewayUpdated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSGatewayResourceConfig(
					name,
					"http://localhost:8080",
					"api-key-1",
					"+15551234567",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact("http://localhost:8080"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact("api-key-1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("recipients"),
						knownvalue.StringExact("+15551234567"),
					),
				},
			},
			{
				Config: testAccNotificationSMSGatewayResourceConfig(
					nameUpdated,
					"https://sms.example.com/",
					"api-key-2",
					"+15551234567, +15559876543 ,",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact("https://sms.example.com/"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact("api-key-2"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsgateway.test",
						tfjsonpath.New("recipients"),
						knownvalue.StringExact("+15551234567, +15559876543 ,"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_smsgateway.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationSMSGatewayResourceInvalidConfig covers the validators and the required-argument
// errors. A blank recipients value is rejected because the server would send nothing and still
// report success.
func TestAccNotificationSMSGatewayResourceInvalidConfig(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSGatewayInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSGatewayResourceConfig(
					name, "http://localhost:8080", "api-key", "  ",
				),
				ExpectError: regexp.MustCompile(`must\s+not\s+be\s+empty`),
			},
			{
				Config: testAccNotificationSMSGatewayResourceConfig(
					name, "", "api-key", "+15551234567",
				),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length`),
			},
			{
				Config: testAccNotificationSMSGatewayResourceConfig(
					name, "http://localhost:8080", "", "+15551234567",
				),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length`),
			},
			{
				Config:      testAccNotificationSMSGatewayResourceMissingRecipientsConfig(name),
				ExpectError: regexp.MustCompile(`The argument "recipients" is required`),
			},
		},
	})
}

func testAccNotificationSMSGatewayResourceConfig(
	name string,
	url string,
	apiKey string,
	recipients string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsgateway" "test" {
  name       = %[1]q
  is_active  = true
  url        = %[2]q
  api_key    = %[3]q
  recipients = %[4]q
}
`, name, url, apiKey, recipients)
}

func testAccNotificationSMSGatewayResourceMissingRecipientsConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsgateway" "invalid" {
  name    = %[1]q
  url     = "http://localhost:8080"
  api_key = "api-key"
}
`, name)
}
