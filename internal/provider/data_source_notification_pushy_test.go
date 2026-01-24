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

func TestAccNotificationPushyDataSourceByID(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPushyDSID")
	apiKey := "test-api-key-ds-id-123456789"
	token := "test-device-token-ds-id-abc123"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPushyDataSourceByIDConfig(name, apiKey, token),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pushy.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationPushyDataSourceByIDConfig(
	name string, apiKey string, token string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pushy" "test" {
  name     = %[1]q
  is_active = true
  api_key  = %[2]q
  token    = %[3]q
}

data "uptimekuma_notification_pushy" "test" {
  id = uptimekuma_notification_pushy.test.id
}
`, name, apiKey, token)
}
