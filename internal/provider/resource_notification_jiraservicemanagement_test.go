package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationJiraServiceManagementResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationJiraServiceManagement")
	nameUpdated := acctest.RandomWithPrefix("NotificationJiraServiceManagementUpdated")
	apiToken := "test-api-token-123"
	apiTokenUpdated := "test-api-token-456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationJiraServiceManagementResourceConfig(
					name,
					"cloud-id-123",
					"user@example.com",
					apiToken,
					1,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("cloud_id"),
						knownvalue.StringExact("cloud-id-123"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("email"),
						knownvalue.StringExact("user@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("api_token"),
						knownvalue.StringExact(apiToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("priority"),
						knownvalue.Int64Exact(1),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationJiraServiceManagementResourceConfig(
					nameUpdated,
					"cloud-id-456",
					"updated@example.com",
					apiTokenUpdated,
					3,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("cloud_id"),
						knownvalue.StringExact("cloud-id-456"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("email"),
						knownvalue.StringExact("updated@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("api_token"),
						knownvalue.StringExact(apiTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("priority"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_jiraservicemanagement.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_jiraservicemanagement.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationJiraServiceManagementImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_token"},
			},
		},
	})
}

func testAccNotificationJiraServiceManagementResourceConfig(
	name string, cloudID string, email string, apiToken string, priority int64,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_jiraservicemanagement" "test" {
  name      = %[1]q
  is_active = true
  cloud_id  = %[2]q
  email     = %[3]q
  api_token = %[4]q
  priority  = %[5]d
}
`, name, cloudID, email, apiToken, priority)
}

func testAccNotificationJiraServiceManagementImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_jiraservicemanagement.test"]
	return rs.Primary.Attributes["id"], nil
}
