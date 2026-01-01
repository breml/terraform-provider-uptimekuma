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

func TestAccNotificationAlertaDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationAlerta")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAlertaDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_alerta.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationAlertaDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_alerta.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationAlertaDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_alerta" "test" {
  name         = %[1]q
  is_active    = true
  api_endpoint = "https://alerta.example.com"
  api_key      = "test-api-key-12345"
}

data "uptimekuma_notification_alerta" "test" {
  name = uptimekuma_notification_alerta.test.name
}
`, name)
}

func testAccNotificationAlertaDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_alerta" "test" {
  name         = %[1]q
  is_active    = true
  api_endpoint = "https://alerta.example.com"
  api_key      = "test-api-key-12345"
}

data "uptimekuma_notification_alerta" "test" {
  id = uptimekuma_notification_alerta.test.id
}
`, name)
}
