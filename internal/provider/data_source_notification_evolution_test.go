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

func TestAccNotificationEvolutionDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationEvolution")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationEvolutionDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_evolution.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationEvolutionDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_evolution.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationEvolutionDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_evolution" "test" {
  name          = %[1]q
  is_active     = true
  api_url       = "https://api.evolution.example.com"
  instance_name = "testinstance"
  auth_token    = "testAuthToken123"
  recipient     = "+551198765432"
}

data "uptimekuma_notification_evolution" "test" {
  name = uptimekuma_notification_evolution.test.name
}
`, name)
}

func testAccNotificationEvolutionDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_evolution" "test" {
  name          = %[1]q
  is_active     = true
  api_url       = "https://api.evolution.example.com"
  instance_name = "testinstance"
  auth_token    = "testAuthToken123"
  recipient     = "+551198765432"
}

data "uptimekuma_notification_evolution" "test" {
  id = uptimekuma_notification_evolution.test.id
}
`, name)
}
