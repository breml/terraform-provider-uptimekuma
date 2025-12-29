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

func TestAccNotificationSMTPResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMTP")
	nameUpdated := acctest.RandomWithPrefix("NotificationSMTPUpdated")
	host := "smtp.example.com"
	hostUpdated := "smtp.updated.com"
	from := "uptime-kuma@example.com"
	fromUpdated := "monitoring@example.com"
	to := "admin@example.com"
	toUpdated := "alerts@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMTPResourceConfig(
					name,
					host,
					587,
					true,
					false,
					from,
					to,
					"",
					"",
					"",
					false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("host"),
						knownvalue.StringExact(host),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(587),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("secure"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("ignore_tls_error"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("from"),
						knownvalue.StringExact(from),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("to"),
						knownvalue.StringExact(to),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("html_body"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationSMTPResourceConfig(
					nameUpdated,
					hostUpdated,
					25,
					false,
					true,
					fromUpdated,
					toUpdated,
					"supervisor@example.com",
					"archive@example.com",
					"Updated Alert",
					true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("host"),
						knownvalue.StringExact(hostUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(25),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("secure"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("ignore_tls_error"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("from"),
						knownvalue.StringExact(fromUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("to"),
						knownvalue.StringExact(toUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("cc"),
						knownvalue.StringExact("supervisor@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("bcc"),
						knownvalue.StringExact("archive@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("custom_subject"),
						knownvalue.StringExact("Updated Alert"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("html_body"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccNotificationSMTPResourceConfig(
	name string,
	host string,
	port int64,
	secure bool,
	ignoreTLSError bool,
	from string,
	to string,
	cc string,
	bcc string,
	customSubject string,
	htmlBody bool,
) string {
	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smtp" "test" {
  name              = %[1]q
  is_active         = true
  host              = %[2]q
  port              = %[3]d
  secure            = %[4]t
  ignore_tls_error  = %[5]t
  from              = %[6]q
  to                = %[7]q
  html_body         = %[11]t
`, name, host, port, secure, ignoreTLSError, from, to, cc, bcc, customSubject, htmlBody)

	if cc != "" {
		config += fmt.Sprintf("  cc = %q\n", cc)
	}

	if bcc != "" {
		config += fmt.Sprintf("  bcc = %q\n", bcc)
	}

	if customSubject != "" {
		config += fmt.Sprintf("  custom_subject = %q\n", customSubject)
	}

	config += "}\n"
	return config
}
