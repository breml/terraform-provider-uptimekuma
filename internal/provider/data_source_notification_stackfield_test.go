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

func TestAccNotificationStackfieldDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationStackfieldDS")
	webhookURL := "https://api.stackfield.com/hooks/XXXXXXXXXXXXXXXXXXXXXXXX"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationStackfieldDataSourceConfig(name, webhookURL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_stackfield.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_stackfield.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccNotificationStackfieldDataSourceConfig(name string, webhookURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_stackfield" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
}

data "uptimekuma_notification_stackfield" "test" {
  name = uptimekuma_notification_stackfield.test.name
}
`, name, webhookURL)
}
