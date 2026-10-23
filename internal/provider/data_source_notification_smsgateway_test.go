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

func TestAccNotificationSMSGatewayDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationSMSGateway")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSGatewayDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smsgateway.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smsgateway.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smsgateway.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSMSGatewayDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsgateway" "test" {
  name       = %[1]q
  is_active  = true
  url        = "http://localhost:8080"
  api_key    = "test-api-key"
  recipients = "+15551234567"
}

data "uptimekuma_notification_smsgateway" "by_name" {
  name = uptimekuma_notification_smsgateway.test.name
}

data "uptimekuma_notification_smsgateway" "by_id" {
  id = uptimekuma_notification_smsgateway.test.id
}
`, name)
}
