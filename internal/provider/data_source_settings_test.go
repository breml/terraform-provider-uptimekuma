package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSettingsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingsDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_settings.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("settings"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_settings.test",
						tfjsonpath.New("server_timezone"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_settings.test",
						tfjsonpath.New("keep_data_period_days"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_settings.test",
						tfjsonpath.New("entry_page"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccSettingsDataSourceConfig() string {
	return providerConfig() + `
data "uptimekuma_settings" "test" {}
`
}
