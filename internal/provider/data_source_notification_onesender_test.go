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

func TestAccNotificationOnesenderDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationOnesender")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOnesenderDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_onesender.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationOnesenderDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_onesender.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationOnesenderDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_onesender" "test" {
  name          = %[1]q
  is_active     = true
  url           = "https://onesender.example.com/api/v1/send"
  token         = "test-token-abc123"
  receiver      = "+6281234567890"
  type_receiver = "private"
}

data "uptimekuma_notification_onesender" "test" {
  name = uptimekuma_notification_onesender.test.name
}
`, name)
}

func testAccNotificationOnesenderDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_onesender" "test" {
  name          = %[1]q
  is_active     = true
  url           = "https://onesender.example.com/api/v1/send"
  token         = "test-token-abc123"
  receiver      = "+6281234567890"
  type_receiver = "private"
}

data "uptimekuma_notification_onesender" "test" {
  id = uptimekuma_notification_onesender.test.id
}
`, name)
}
