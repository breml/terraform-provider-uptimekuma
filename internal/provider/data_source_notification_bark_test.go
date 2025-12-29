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

func TestAccNotificationBarkDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationBark")
	endpointURL := "https://api.bark.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBarkDataSourceConfig(
					name,
					endpointURL,
					"test-group",
					"default",
					"v1",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_bark.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_bark.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccNotificationBarkDataSourceConfig(
	name string,
	endpoint string,
	group string,
	sound string,
	apiVersion string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bark" "test" {
  name        = %[1]q
  is_active   = true
  endpoint    = %[2]q
  group       = %[3]q
  sound       = %[4]q
  api_version = %[5]q
}

data "uptimekuma_notification_bark" "test" {
  name = uptimekuma_notification_bark.test.name
}
`, name, endpoint, group, sound, apiVersion)
}
