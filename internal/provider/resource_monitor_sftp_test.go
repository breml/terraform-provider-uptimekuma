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

// testSFTPPrivateKey is a placeholder in the shape of an OpenSSH private key.
// The acceptance test container has no SFTP server to log in to, so the key is
// never used; it only has to round-trip, newlines included.
const testSFTPPrivateKey = "-----BEGIN OPENSSH PRIVATE KEY-----\n" +
	"b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\n" +
	"-----END OPENSSH PRIVATE KEY-----\n"

// The SFTP tests only assert that the configuration round-trips: the
// acceptance test container has no SFTP server, so the check itself never
// goes up.
func TestAccMonitorSFTPResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSFTPMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestSFTPMonitorUpdated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSFTPResourceConfigPassword(
					name, "sftp.example.com", 2222, "backup", "s3cret", "/upload", "Test SFTP monitor",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("sftp.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(2222),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Test SFTP monitor"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(10),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_username"),
						knownvalue.StringExact("backup"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_auth_method"),
						knownvalue.StringExact("password"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_password"),
						knownvalue.StringExact("s3cret"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_private_key"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("sftp_path"),
						knownvalue.StringExact("/upload"),
					),
				},
			},
			{
				// Switching to key authentication must clear the password the
				// previous step stored, rather than leave it behind on the server.
				Config: testAccMonitorSFTPResourceConfigPrivateKey(
					nameUpdated, "sftp2.example.com", "deploy", testSFTPPrivateKey, "phrase",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("sftp2.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_username"),
						knownvalue.StringExact("deploy"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_auth_method"),
						knownvalue.StringExact("privateKey"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_private_key"),
						knownvalue.StringExact(testSFTPPrivateKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_passphrase"),
						knownvalue.StringExact("phrase"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_password"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("port"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("sftp_path"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("description"),
						knownvalue.Null(),
					),
				},
			},
			{
				// Switching back to password authentication must clear the key
				// material, which is worse to leave behind than a password.
				Config: testAccMonitorSFTPResourceConfigPassword(
					name, "sftp.example.com", 2222, "backup", "s3cret", "/upload", "Test SFTP monitor",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_auth_method"),
						knownvalue.StringExact("password"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_password"),
						knownvalue.StringExact("s3cret"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_private_key"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_passphrase"),
						knownvalue.Null(),
					),
				},
			},
			{
				// The server returns the credentials to the owning user, so an
				// import needs no ignore list.
				ResourceName:      "uptimekuma_monitor_sftp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorSFTPResourceConfigPassword(
	name string,
	hostname string,
	port int64,
	username string,
	password string,
	sftpPath string,
	description string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sftp" "test" {
  name         = %[1]q
  description  = %[7]q
  hostname     = %[2]q
  port         = %[3]d
  ssh_username = %[4]q
  ssh_password = %[5]q
  sftp_path    = %[6]q
  active       = true
}
`, name, hostname, port, username, password, sftpPath, description)
}

func testAccMonitorSFTPResourceConfigPrivateKey(
	name string,
	hostname string,
	username string,
	privateKey string,
	passphrase string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sftp" "test" {
  name            = %[1]q
  hostname        = %[2]q
  ssh_username    = %[3]q
  ssh_auth_method = "privateKey"
  ssh_private_key = %[4]q
  ssh_passphrase  = %[5]q
  active          = true
}
`, name, hostname, username, privateKey, passphrase)
}

// TestAccMonitorSFTPResourceMinimal verifies that the optional fields stay null
// when they are not configured. The server stores a nil pointer as SQL NULL and
// reads it back as null rather than as the value the check falls back to, so
// modelling the fallbacks as Terraform defaults would cause a perpetual diff.
// The timeout and the auth method are the exceptions: both always carry a
// value on the wire.
func TestAccMonitorSFTPResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSFTPMonitorMinimal")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSFTPResourceConfigMinimal(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("port"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_password"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_private_key"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_passphrase"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("sftp_path"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(10),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("ssh_auth_method"),
						knownvalue.StringExact("password"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_sftp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorSFTPResourceConfigMinimal(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sftp" "test" {
  name         = %[1]q
  hostname     = "sftp.example.com"
  ssh_username = "backup"
}
`, name)
}

// TestAccMonitorSFTPResourceFractionalTimeout verifies that a fractional
// timeout round-trips unchanged and that removing it returns to the default.
func TestAccMonitorSFTPResourceFractionalTimeout(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSFTPMonitorFractionalTimeout")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSFTPResourceConfigWithAttribute(name, "timeout", "2.5"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(2.5),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_sftp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitorSFTPResourceConfigMinimal(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(10),
					),
				},
			},
		},
	})
}

// TestAccMonitorSFTPResourceWithAllOptions covers the base monitor fields, which
// every monitor type copies into its own build and populate helpers rather than
// sharing, so a slip here would silently drop a monitor out of its group or
// leave it without notifications.
func TestAccMonitorSFTPResourceWithAllOptions(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestSFTPMonitorGroup")
	notificationName := acctest.RandomWithPrefix("TestSFTPMonitorNotification")
	tagName := acctest.RandomWithPrefix("TestSFTPMonitorTag")
	name := acctest.RandomWithPrefix("TestSFTPMonitorFull")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSFTPResourceConfigWithAllOptions(
					groupName, notificationName, tagName, name,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("notification_ids"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("tags"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("tags").AtSliceIndex(0).AtMapKey("value"),
						knownvalue.StringExact("production"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(10),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sftp.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_sftp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorSFTPResourceConfigWithAllOptions(
	groupName string,
	notificationName string,
	tagName string,
	name string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_notification_webhook" "test" {
  name        = %[2]q
  webhook_url = "https://example.com/webhook"
  is_active   = true
}

resource "uptimekuma_tag" "test" {
  name  = %[3]q
  color = "#00ff00"
}

resource "uptimekuma_monitor_sftp" "test" {
  name             = %[4]q
  hostname         = "sftp.example.com"
  ssh_username     = "backup"
  ssh_password     = "s3cret"
  interval         = 300
  retry_interval   = 120
  resend_interval  = 10
  max_retries      = 5
  upside_down      = true
  active           = true
  parent           = uptimekuma_monitor_group.test.id
  notification_ids = [uptimekuma_notification_webhook.test.id]

  tags = [
    {
      tag_id = uptimekuma_tag.test.id
      value  = "production"
    },
  ]
}
`, groupName, notificationName, tagName, name)
}

// TestAccMonitorSFTPResourceValidators pins the plan-time checks, which catch
// what the client would otherwise reject at apply time or the server would
// silently reinterpret.
func TestAccMonitorSFTPResourceValidators(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSFTPMonitorValidators")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMonitorSFTPResourceConfigWithAttribute(name, "hostname", `" "`),
				ExpectError: regexp.MustCompile(`must\s+not\s+be\s+empty`),
			},
			{
				Config:      testAccMonitorSFTPResourceConfigWithAttribute(name, "ssh_username", `""`),
				ExpectError: regexp.MustCompile(`must\s+not\s+be\s+empty`),
			},
			{
				Config:      testAccMonitorSFTPResourceConfigWithAttribute(name, "port", "0"),
				ExpectError: regexp.MustCompile(`must be between 1 and 65535`),
			},
			{
				Config:      testAccMonitorSFTPResourceConfigWithAttribute(name, "timeout", "0"),
				ExpectError: regexp.MustCompile(`must be at least 0\.10{5}`),
			},
			{
				Config:      testAccMonitorSFTPResourceConfigWithAttribute(name, "ssh_auth_method", `"publickey"`),
				ExpectError: regexp.MustCompile(`value must be one of`),
			},
			{
				Config:      testAccMonitorSFTPResourceConfigWithAttribute(name, "ssh_auth_method", `"privateKey"`),
				ExpectError: regexp.MustCompile(`"ssh_private_key"\s+must\s+be\s+set`),
			},
			{
				Config:      testAccMonitorSFTPResourceConfigWithAttribute(name, "ssh_passphrase", `"phrase"`),
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
			{
				Config:      testAccMonitorSFTPResourceConfigWithAttribute(name, "ssh_private_key", `"key"`),
				ExpectError: regexp.MustCompile(`"ssh_private_key"\s+is\s+ignored`),
			},
			{
				Config:      testAccMonitorSFTPResourceConfigWithAttribute(name, "ssh_password", `" "`),
				ExpectError: regexp.MustCompile(`must\s+not\s+be\s+empty`),
			},
		},
	})
}

// testAccMonitorSFTPResourceConfigWithAttribute renders a minimal monitor with
// one attribute added or, for hostname and ssh_username, overridden.
func testAccMonitorSFTPResourceConfigWithAttribute(name string, attribute string, value string) string {
	attrs := map[string]string{
		"hostname":     `"sftp.example.com"`,
		"ssh_username": `"backup"`,
	}
	attrs[attribute] = value

	extra := ""
	if attribute != "hostname" && attribute != "ssh_username" {
		extra = fmt.Sprintf("  %s = %s\n", attribute, value)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sftp" "test" {
  name         = %[1]q
  hostname     = %[2]s
  ssh_username = %[3]s
%[4]s}
`, name, attrs["hostname"], attrs["ssh_username"], extra)
}
