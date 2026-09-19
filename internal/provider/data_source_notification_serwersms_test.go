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

func TestAccNotificationSerwersmsDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSerwersms")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSerwersmsDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_serwersms.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_serwersms.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSerwersmsDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_serwersms" "test" {
  name         = %[1]q
  is_active    = true
  username     = "testuser"
  password     = "testpass123"
  phone_number = "+48501234567"
  sender_name  = "TestSender"
}

data "uptimekuma_notification_serwersms" "by_name" {
  name = uptimekuma_notification_serwersms.test.name
}

data "uptimekuma_notification_serwersms" "by_id" {
  id = uptimekuma_notification_serwersms.test.id
}
`, name)
}
