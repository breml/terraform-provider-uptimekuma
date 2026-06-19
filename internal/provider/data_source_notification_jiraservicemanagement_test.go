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

func TestAccNotificationJiraServiceManagementDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationJiraServiceManagement")
	apiToken := "test-api-token-123"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationJiraServiceManagementDataSourceConfig(name, apiToken),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_jiraservicemanagement.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_jiraservicemanagement.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_jiraservicemanagement.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationJiraServiceManagementDataSourceConfig(name string, apiToken string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_jiraservicemanagement" "test" {
  name      = %[1]q
  is_active = true
  cloud_id  = "cloud-id-123"
  email     = "user@example.com"
  api_token = %[2]q
  priority  = 1
}

data "uptimekuma_notification_jiraservicemanagement" "by_name" {
  name = uptimekuma_notification_jiraservicemanagement.test.name
}

data "uptimekuma_notification_jiraservicemanagement" "by_id" {
  id = uptimekuma_notification_jiraservicemanagement.test.id
}
`, name, apiToken)
}
