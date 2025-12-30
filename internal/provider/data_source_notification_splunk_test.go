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

func TestAccNotificationSplunkDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSplunk")
	restURL := "https://api.victorops.com/api/v2"
	severity := "critical"
	autoResolve := "resolve"
	integrationKey := "integration_key_12345"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSplunkDataSourceConfig(
					name,
					restURL,
					severity,
					autoResolve,
					integrationKey,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_splunk.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_splunk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSplunkDataSourceConfig(
	name string,
	restURL string,
	severity string,
	autoResolve string,
	integrationKey string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_splunk" "test" {
  name              = %[1]q
  is_active         = true
  rest_url          = %[2]q
  severity          = %[3]q
  auto_resolve      = %[4]q
  integration_key   = %[5]q
}

data "uptimekuma_notification_splunk" "test" {
  name = uptimekuma_notification_splunk.test.name
}
`, name, restURL, severity, autoResolve, integrationKey)
}
