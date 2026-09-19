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

func TestAccNotificationClicksendSmsDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationClicksendSms")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationClicksendSmsDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_clicksendsms.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_clicksendsms.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationClicksendSmsDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_clicksendsms" "test" {
  name        = %[1]q
  is_active   = true
  login       = "test@example.com"
  password    = "testApiKey123"
  to_number   = "+61412345678"
  sender_name = "TestSender"
}

data "uptimekuma_notification_clicksendsms" "by_name" {
  name = uptimekuma_notification_clicksendsms.test.name
}

data "uptimekuma_notification_clicksendsms" "by_id" {
  id = uptimekuma_notification_clicksendsms.test.id
}
`, name)
}
