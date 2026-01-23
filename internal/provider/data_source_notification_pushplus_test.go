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

func TestAccNotificationPushPlusDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPushPlus")
	sendKey := "test-send-key-datasource-123"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPushPlusDataSourceConfig(name, sendKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pushplus.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationPushPlusDataSourceConfig(name string, sendKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pushplus" "test" {
  name     = %[1]q
  is_active = true
  send_key = %[2]q
}

data "uptimekuma_notification_pushplus" "test" {
  name = uptimekuma_notification_pushplus.test.name
}
`, name, sendKey)
}
