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

// TestAccNotificationClickUpResource covers create, an update of every ClickUp-specific field and
// an import round-trip. disable_url walks through unset, true, false and unset again: the server
// stores it only when sent, so an unset value must stay null rather than read back as false.
func TestAccNotificationClickUpResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationClickUp")
	nameUpdated := acctest.RandomWithPrefix("NotificationClickUpUpdated")

	initial := clickUpTestConfig{
		name:        name,
		token:       "pk_12345678_TESTTOKEN",
		workspaceID: "9012345678",
		channelID:   "abc12-345",
	}

	updated := clickUpTestConfig{
		name:        nameUpdated,
		token:       "pk_87654321_OTHERTOKEN",
		workspaceID: "9087654321",
		channelID:   "xyz98-765",
		disableURL:  "true",
	}

	disabledFalse := updated
	disabledFalse.disableURL = "false"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationClickUpResourceConfig(initial),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("token"),
						knownvalue.StringExact("pk_12345678_TESTTOKEN"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("workspace_id"),
						knownvalue.StringExact("9012345678"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("channel_id"),
						knownvalue.StringExact("abc12-345"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("disable_url"),
						knownvalue.Null(),
					),
				},
			},
			{
				Config: testAccNotificationClickUpResourceConfig(updated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("token"),
						knownvalue.StringExact("pk_87654321_OTHERTOKEN"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("workspace_id"),
						knownvalue.StringExact("9087654321"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("channel_id"),
						knownvalue.StringExact("xyz98-765"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("disable_url"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_clickup.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationClickUpResourceConfig(disabledFalse),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("disable_url"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_clickup.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationClickUpResourceConfig(initial),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clickup.test",
						tfjsonpath.New("disable_url"),
						knownvalue.Null(),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_clickup.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccNotificationClickUpResourceInvalidConfig covers the length validators and the
// required-argument error.
func TestAccNotificationClickUpResourceInvalidConfig(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationClickUpInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationClickUpResourceConfig(clickUpTestConfig{
					name:        name,
					token:       "",
					workspaceID: "9012345678",
					channelID:   "abc12-345",
				}),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length`),
			},
			{
				Config: testAccNotificationClickUpResourceConfig(clickUpTestConfig{
					name:        name,
					token:       "pk_12345678_TESTTOKEN",
					workspaceID: "9012345678",
					channelID:   "",
				}),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length`),
			},
			{
				Config: testAccNotificationClickUpResourceConfig(clickUpTestConfig{
					name:        name,
					token:       "pk_12345678_TESTTOKEN",
					workspaceID: "",
					channelID:   "abc12-345",
				}),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length`),
			},
			{
				Config:      testAccNotificationClickUpResourceMissingWorkspaceConfig(name),
				ExpectError: regexp.MustCompile(`The argument "workspace_id" is required`),
			},
		},
	})
}

// clickUpTestConfig holds the attribute values of a ClickUp test resource. An empty disableURL
// leaves the attribute out of the configuration.
type clickUpTestConfig struct {
	name        string
	token       string
	workspaceID string
	channelID   string
	disableURL  string
}

func testAccNotificationClickUpResourceConfig(cfg clickUpTestConfig) string {
	disableURL := ""
	if cfg.disableURL != "" {
		disableURL = "  disable_url  = " + cfg.disableURL + "\n"
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_clickup" "test" {
  name         = %[1]q
  token        = %[2]q
  workspace_id = %[3]q
  channel_id   = %[4]q
%[5]s}
`, cfg.name, cfg.token, cfg.workspaceID, cfg.channelID, disableURL)
}

func testAccNotificationClickUpResourceMissingWorkspaceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_clickup" "invalid" {
  name       = %[1]q
  token      = "pk_12345678_TESTTOKEN"
  channel_id = "abc12-345"
}
`, name)
}
