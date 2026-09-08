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

const notificationNtfyResourceName = "uptimekuma_notification_ntfy.test"

func TestAccNotificationNtfyResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNtfy")
	nameUpdated := acctest.RandomWithPrefix("NotificationNtfyUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNotificationNtfyResourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("authentication_method"),
						knownvalue.StringExact("none"),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("server_url"),
						knownvalue.StringExact("https://ntfy.sh"),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("priority"),
						knownvalue.Int32Exact(5),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("topic"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("access_token"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("username"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("password"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("icon"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("priority_down"),
						knownvalue.Null(),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccNotificationNtfyResourceConfig(nameUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("authentication_method"),
						knownvalue.StringExact("none"),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("server_url"),
						knownvalue.StringExact("https://ntfy.sh"),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("priority"),
						knownvalue.Int32Exact(5),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("topic"),
						knownvalue.StringExact(nameUpdated),
					),
				},
			},
			// Import testing
			{
				ResourceName:      notificationNtfyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccNotificationNtfyResourceAccessToken is the regression test for
// https://github.com/breml/terraform-provider-uptimekuma/issues/389: the access
// token has to reach Uptime Kuma on create and must survive an update that only
// touches an unrelated attribute.
func TestAccNotificationNtfyResourceAccessToken(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNtfyToken")
	topicUpdated := acctest.RandomWithPrefix("topic")
	accessToken := "tk_abcdefghijklmnopqrstuvwxyz"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNtfyResourceConfigAccessToken(name, name, accessToken),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("authentication_method"),
						knownvalue.StringExact("accessToken"),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessToken),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("icon"),
						knownvalue.StringExact("https://example.com/icon.png"),
					),
				},
			},
			// Change an unrelated attribute: the access token must not be nulled.
			{
				Config: testAccNotificationNtfyResourceConfigAccessToken(name, topicUpdated, accessToken),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("topic"),
						knownvalue.StringExact(topicUpdated),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessToken),
					),
				},
			},
			{
				ResourceName:      notificationNtfyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNotificationNtfyResourceUsernamePassword(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNtfyBasic")
	topicUpdated := acctest.RandomWithPrefix("topic")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNtfyResourceConfigUsernamePassword(name, name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("authentication_method"),
						knownvalue.StringExact("usernamePassword"),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("username"),
						knownvalue.StringExact("ntfy-user"),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("password"),
						knownvalue.StringExact("ntfy-secret"),
					),
				},
			},
			// Change an unrelated attribute: the credentials must not be nulled.
			{
				Config: testAccNotificationNtfyResourceConfigUsernamePassword(name, topicUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("topic"),
						knownvalue.StringExact(topicUpdated),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("username"),
						knownvalue.StringExact("ntfy-user"),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("password"),
						knownvalue.StringExact("ntfy-secret"),
					),
				},
			},
			// Import verification is what proves the credentials reached Uptime Kuma: the state
			// checks above only echo back what Create and Update wrote from the plan.
			{
				ResourceName:      notificationNtfyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNotificationNtfyResourceTemplate(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNtfyTemplate")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNtfyResourceConfigTemplate(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("priority_down"),
						knownvalue.Int32Exact(1),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("call"),
						knownvalue.StringExact("+41791234567"),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("use_template"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("custom_title"),
						knownvalue.StringExact("{{ name }} is {{ status }}"),
					),
					statecheck.ExpectKnownValue(
						notificationNtfyResourceName,
						tfjsonpath.New("custom_message"),
						knownvalue.StringExact("{{ msg }}"),
					),
				},
			},
			{
				ResourceName:      notificationNtfyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNotificationNtfyResourceMissingAccessToken(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNtfyValidation")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNtfyResourceConfigAuthOnly(name, "accessToken"),
				ExpectError: regexp.MustCompile(
					`Missing Attribute for ntfy Authentication Method`,
				),
			},
		},
	})
}

func TestAccNotificationNtfyResourceMissingUsernamePassword(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNtfyValidation")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNtfyResourceConfigAuthOnly(name, "usernamePassword"),
				ExpectError: regexp.MustCompile(
					`Missing Attribute for ntfy Authentication Method`,
				),
			},
		},
	})
}

func TestAccNotificationNtfyResourceConflictingCredentials(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNtfyValidation")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNtfyResourceConfigConflicting(name),
				ExpectError: regexp.MustCompile(
					`Invalid Attribute for ntfy Authentication Method`,
				),
			},
		},
	})
}

// TestAccNotificationNtfyResourceCredentialsWithoutMethod covers the configuration that hits
// https://github.com/breml/terraform-provider-uptimekuma/issues/389 without any explicit
// authentication_method: the schema default "none" applies, so the credentials would be sent
// to Uptime Kuma and ignored there.
func TestAccNotificationNtfyResourceCredentialsWithoutMethod(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNtfyValidation")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNtfyResourceConfigCredentialsWithoutMethod(name),
				ExpectError: regexp.MustCompile(
					`Invalid Attribute for ntfy Authentication Method`,
				),
			},
		},
	})
}

func testAccNotificationNtfyResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ntfy" "test" {
  name      = %[1]q
  is_active = true

  authentication_method = "none"
  server_url            = "https://ntfy.sh"
  priority              = 5
  topic                 = %[1]q
}
`, name)
}

func testAccNotificationNtfyResourceConfigAccessToken(name string, topic string, accessToken string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ntfy" "test" {
  name      = %[1]q
  is_active = true

  authentication_method = "accessToken"
  access_token          = %[3]q
  server_url            = "https://ntfy.sh"
  priority              = 5
  topic                 = %[2]q
  icon                  = "https://example.com/icon.png"
}
`, name, topic, accessToken)
}

func testAccNotificationNtfyResourceConfigUsernamePassword(name string, topic string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ntfy" "test" {
  name      = %[1]q
  is_active = true

  authentication_method = "usernamePassword"
  username              = "ntfy-user"
  password              = "ntfy-secret"
  server_url            = "https://ntfy.sh"
  priority              = 5
  topic                 = %[2]q
}
`, name, topic)
}

func testAccNotificationNtfyResourceConfigTemplate(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ntfy" "test" {
  name      = %[1]q
  is_active = true

  server_url     = "https://ntfy.sh"
  topic          = %[1]q
  priority       = 4
  priority_down  = 1
  call           = "+41791234567"
  use_template   = true
  custom_title   = "{{ name }} is {{ status }}"
  custom_message = "{{ msg }}"
}
`, name)
}

func testAccNotificationNtfyResourceConfigAuthOnly(name string, authenticationMethod string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ntfy" "test" {
  name = %[1]q

  authentication_method = %[2]q
  topic                 = %[1]q
}
`, name, authenticationMethod)
}

func testAccNotificationNtfyResourceConfigCredentialsWithoutMethod(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ntfy" "test" {
  name = %[1]q

  username = "ntfy-user"
  password = "ntfy-secret"
  topic    = %[1]q
}
`, name)
}

func testAccNotificationNtfyResourceConfigConflicting(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ntfy" "test" {
  name = %[1]q

  authentication_method = "accessToken"
  access_token          = "tk_abcdefghijklmnopqrstuvwxyz"
  username              = "ntfy-user"
  topic                 = %[1]q
}
`, name)
}
