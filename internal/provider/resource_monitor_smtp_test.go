package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccMonitorSMTPResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSMTPMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestSMTPMonitorUpdated")
	description := "Test SMTP monitor description"
	descriptionUpdated := "Updated test SMTP monitor description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSMTPResourceConfig(
					name,
					description,
					"smtp.example.com",
					587,
					"starttls",
					true,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("smtp.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(587),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("smtp_security"),
						knownvalue.StringExact("starttls"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorSMTPResourceConfig(
					nameUpdated,
					descriptionUpdated,
					"mail.example.com",
					465,
					"secure",
					false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(descriptionUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("mail.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(465),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("smtp_security"),
						knownvalue.StringExact("secure"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_smtp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorSMTPResourceConfig(
	name string,
	description string,
	hostname string,
	port int64,
	smtpSecurity string,
	domainExpiry bool,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_smtp" "test" {
  name                        = %[1]q
  description                 = %[2]q
  hostname                    = %[3]q
  port                        = %[4]d
  smtp_security               = %[5]q
  active                      = true
  domain_expiry_notification  = %[6]t
}
`, name, description, hostname, port, smtpSecurity, domainExpiry)
}

func TestAccMonitorSMTPResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSMTPMonitorMinimal")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSMTPResourceConfigMinimal(name, "smtp.example.com"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("smtp.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(587),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("smtp_security"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorSMTPResourceConfigMinimal(name string, hostname string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_smtp" "test" {
  name     = %[1]q
  hostname = %[2]q
}
`, name, hostname)
}

func TestAccMonitorSMTPResourceWithAllOptions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSMTPMonitorFull")
	description := "Full test SMTP monitor"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSMTPResourceConfigWithAllOptions(name, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("smtp.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(465),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("smtp_security"),
						knownvalue.StringExact("secure"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(90),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccMonitorSMTPResourceConfigWithAllOptions(name string, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_smtp" "test" {
  name            = %[1]q
  description     = %[2]q
  hostname        = "smtp.example.com"
  port            = 465
  smtp_security   = "secure"
  interval        = 120
  retry_interval  = 90
  resend_interval = 0
  max_retries     = 5
  upside_down     = false
  active          = false
}
`, name, description)
}

func TestAccMonitorSMTPResourceSecurityModes(t *testing.T) {
	securityModes := []string{"secure", "starttls", "nostarttls"}

	for _, mode := range securityModes {
		t.Run(mode, func(t *testing.T) {
			name := acctest.RandomWithPrefix(fmt.Sprintf("TestSMTP%s", mode))

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccMonitorSMTPResourceConfigSecurityMode(name, "smtp.example.com", mode),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(
								"uptimekuma_monitor_smtp.test",
								tfjsonpath.New("name"),
								knownvalue.StringExact(name),
							),
							statecheck.ExpectKnownValue(
								"uptimekuma_monitor_smtp.test",
								tfjsonpath.New("hostname"),
								knownvalue.StringExact("smtp.example.com"),
							),
							statecheck.ExpectKnownValue(
								"uptimekuma_monitor_smtp.test",
								tfjsonpath.New("smtp_security"),
								knownvalue.StringExact(mode),
							),
						},
					},
					{
						RefreshState:       true,
						ExpectNonEmptyPlan: false,
					},
					{
						ResourceName:      "uptimekuma_monitor_smtp.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func testAccMonitorSMTPResourceConfigSecurityMode(name string, hostname string, securityMode string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_smtp" "test" {
  name          = %[1]q
  hostname      = %[2]q
  smtp_security = %[3]q
}
`, name, hostname, securityMode)
}

// TestAccMonitorSMTPResourceSecure covers issue #470: an SMTPS monitor can be
// described, survives an import without a diff, and is cleared by removing the
// attribute.
func TestAccMonitorSMTPResourceSecure(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSMTPMonitorSecure")
	config := testAccMonitorSMTPResourceConfigSecurityMode(name, "smtp.example.com", "secure")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("smtp_security"),
						knownvalue.StringExact("secure"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_smtp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: testAccMonitorSMTPResourceConfigMinimal(name, "smtp.example.com"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_smtp.test",
						tfjsonpath.New("smtp_security"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

func TestAccMonitorSMTPResourceLegacySecurityModesRejected(t *testing.T) {
	for _, mode := range []string{"None", "STARTTLS", "TLS"} {
		t.Run(mode, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testAccMonitorSMTPResourceConfigSecurityMode("unused", "smtp.example.com", mode),
						PlanOnly:    true,
						ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
					},
				},
			})
		})
	}
}
