package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
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

	resource.ParallelTest(t, resource.TestCase{
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

func TestAccNotificationSMTPResourceAdditionalHeaders(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMTPHeaders")
	headers := `{"X-Custom-Header":"Additional Header"}`
	headersUpdated := `{"X-Custom-Header":"Changed Header","X-Environment":"production"}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMTPAdditionalHeadersConfig(name, headers),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("additional_headers"),
						knownvalue.StringExact(headers),
					),
				},
			},
			{
				Config: testAccNotificationSMTPAdditionalHeadersConfig(name, headersUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("additional_headers"),
						knownvalue.StringExact(headersUpdated),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_smtp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Removing the attribute must clear it on the server rather than leave the previous
			// value behind, which would show up as a plan that never converges.
			{
				Config: testAccNotificationSMTPAdditionalHeadersConfig(name, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smtp.test",
						tfjsonpath.New("additional_headers"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

// testAccNotificationSMTPAdditionalHeadersConfig renders a minimal SMTP notification. An empty
// headers argument omits additional_headers entirely, so the attribute is null rather than empty.
func testAccNotificationSMTPAdditionalHeadersConfig(name string, headers string) string {
	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smtp" "test" {
  name = %[1]q
  host = "smtp.example.com"
  from = "uptime-kuma@example.com"
  to   = "admin@example.com"
`, name)

	if headers != "" {
		config += fmt.Sprintf("  additional_headers = %q\n", headers)
	}

	config += "}\n"

	return config
}

func TestAccNotificationSMTPResourcePasswordWO(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMTPPasswordWO")
	kumaClient := testAccOutOfBandClient(t)
	addr := "uptimekuma_notification_smtp.test"

	// Neither attribute may carry the password in state, whatever the version.
	passwordAbsentFromState := func(version string) []statecheck.StateCheck {
		return []statecheck.StateCheck{
			statecheck.ExpectKnownValue(addr, tfjsonpath.New("password"), knownvalue.Null()),
			statecheck.ExpectKnownValue(addr, tfjsonpath.New("password_wo"), knownvalue.Null()),
			statecheck.ExpectKnownValue(addr, tfjsonpath.New("password_wo_version"), knownvalue.StringExact(version)),
		}
	}

	plainPasswordInState := func(password string) []statecheck.StateCheck {
		return []statecheck.StateCheck{
			statecheck.ExpectKnownValue(addr, tfjsonpath.New("password"), knownvalue.StringExact(password)),
			statecheck.ExpectKnownValue(addr, tfjsonpath.New("password_wo_version"), knownvalue.Null()),
		}
	}

	expectUpdate := resource.ConfigPlanChecks{
		PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		Steps: []resource.TestStep{
			// password and password_wo are mutually exclusive. The full detail is matched, because
			// every attribute combination validator shares the same summary.
			{
				Config: testAccNotificationSMTPPasswordConfig(name, `
  password            = "plain"
  password_wo         = "write-only"
  password_wo_version = "1"`),
				ExpectError: regexp.MustCompile(
					`Attribute\s+"password_wo"\s+cannot\s+be\s+specified\s+when\s+"password"\s+is\s+specified`,
				),
			},
			// password_wo without a version could never be rotated.
			{
				Config: testAccNotificationSMTPPasswordConfig(name, `
  password_wo = "write-only"`),
				ExpectError: regexp.MustCompile(
					`Attribute\s+"password_wo_version"\s+must\s+be\s+specified\s+when\s+"password_wo"\s+is\s+specified`,
				),
			},
			// A version without password_wo would stop Read from refreshing password.
			{
				Config: testAccNotificationSMTPPasswordConfig(name, `
  password_wo_version = "1"`),
				ExpectError: regexp.MustCompile(
					`Attribute\s+"password_wo"\s+must\s+be\s+specified\s+when\s+"password_wo_version"\s+is\s+specified`,
				),
			},
			{
				Config:            testAccNotificationSMTPPasswordConfig(name, `password = "plain-secret"`),
				ConfigStateChecks: plainPasswordInState("plain-secret"),
				Check:             testAccCheckSMTPServerPassword(t, kumaClient, addr, "plain-secret"),
			},
			// Moving from password to password_wo sends the write-only value and drops password
			// from state.
			{
				Config:            testAccNotificationSMTPPasswordWOConfig(name, "first-secret", "1"),
				ConfigPlanChecks:  expectUpdate,
				ConfigStateChecks: passwordAbsentFromState("1"),
				Check:             testAccCheckSMTPServerPassword(t, kumaClient, addr, "first-secret"),
			},
			// A changed write-only value alone is invisible to the plan: nothing is sent.
			{
				Config: testAccNotificationSMTPPasswordWOConfig(name, "second-secret", "1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: testAccCheckSMTPServerPassword(t, kumaClient, addr, "first-secret"),
			},
			// Bumping the version sends the current write-only value.
			{
				Config:            testAccNotificationSMTPPasswordWOConfig(name, "second-secret", "2"),
				ConfigPlanChecks:  expectUpdate,
				ConfigStateChecks: passwordAbsentFromState("2"),
				Check:             testAccCheckSMTPServerPassword(t, kumaClient, addr, "second-secret"),
			},
			// An import has no password_wo_version to go by, so Read copies the server's password
			// into state, as the password_wo description documents. The next apply removes it.
			{
				ResourceName:     addr,
				ImportState:      true,
				ImportStateCheck: testAccCheckSMTPImportedPassword("second-secret"),
			},
			// Moving back to password stores it in state again.
			{
				Config:            testAccNotificationSMTPPasswordConfig(name, `password = "plain-again"`),
				ConfigPlanChecks:  expectUpdate,
				ConfigStateChecks: plainPasswordInState("plain-again"),
				Check:             testAccCheckSMTPServerPassword(t, kumaClient, addr, "plain-again"),
			},
		},
	})
}

// testAccCheckSMTPServerPassword asserts the password Uptime Kuma stores for the notification,
// which is the only place a write-only value can be observed.
func testAccCheckSMTPServerPassword(
	t *testing.T,
	kumaClient *kuma.Client,
	resourceAddr string,
	want string,
) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		id, err := testAccStateID(s, resourceAddr)
		if err != nil {
			return err
		}

		got, err := smtpServerSecrets(t.Context(), kumaClient, id)
		if err != nil {
			return err
		}

		if got["password"] != want {
			return fmt.Errorf("server password for notification %d is %q, want %q", id, got["password"], want)
		}

		return nil
	}
}

// smtpServerSecrets reads the password Uptime Kuma stores for an SMTP notification.
func smtpServerSecrets(ctx context.Context, kumaClient *kuma.Client, id int64) (map[string]string, error) {
	return notificationSecrets(func(n *notification.SMTP) map[string]string {
		return map[string]string{"password": n.Password}
	})(ctx, kumaClient, id)
}

// testAccCheckSMTPImportedPassword asserts the state an import produces: password holds the
// server's password and password_wo_version is unset.
func testAccCheckSMTPImportedPassword(want string) resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) != 1 {
			return fmt.Errorf("import produced %d resources, want 1", len(states))
		}

		attrs := states[0].Attributes

		if attrs["password"] != want {
			return fmt.Errorf("imported password is %q, want %q", attrs["password"], want)
		}

		if version, ok := attrs["password_wo_version"]; ok && version != "" {
			return fmt.Errorf("imported password_wo_version is %q, want unset", version)
		}

		return nil
	}
}

// testAccNotificationSMTPPasswordWOConfig renders an SMTP notification authenticated with a
// write-only password.
func testAccNotificationSMTPPasswordWOConfig(name string, password string, version string) string {
	return testAccNotificationSMTPPasswordConfig(name, fmt.Sprintf(`
  password_wo         = %q
  password_wo_version = %q`, password, version))
}

// testAccNotificationSMTPPasswordConfig renders an SMTP notification with a username, whose
// password is given by passwordAttrs, the HCL lines of the password attributes.
func testAccNotificationSMTPPasswordConfig(name string, passwordAttrs string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smtp" "test" {
  name     = %[1]q
  host     = "smtp.example.com"
  from     = "uptime-kuma@example.com"
  to       = "admin@example.com"
  username = "smtp-user"
  %[2]s
}
`, name, passwordAttrs)
}
