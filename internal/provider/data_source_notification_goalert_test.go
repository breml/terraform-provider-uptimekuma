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

func TestAccNotificationGoAlertDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGoAlert")
	baseURL := "https://goalert.example.com"
	token := "test-token-123456789"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGoAlertDataSourceConfig(
					name,
					baseURL,
					token,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_goalert.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_goalert.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationGoAlertDataSourceConfig(
	name string, baseURL string, token string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_goalert" "test" {
  name     = %[1]q
  is_active = true
  base_url = %[2]q
  token    = %[3]q
}

data "uptimekuma_notification_goalert" "test" {
  name = uptimekuma_notification_goalert.test.name
}
`, name, baseURL, token)
}
