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

func TestAccStatusPageDataSource(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-status")
	title := "Test Status Page"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageDataSourceConfig(slug, title),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_status_page.test",
						tfjsonpath.New("slug"),
						knownvalue.StringExact(slug),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_status_page.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
				},
			},
		},
	})
}

func testAccStatusPageDataSourceConfig(slug string, title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_status_page" "test" {
  slug        = %[1]q
  title       = %[2]q
  description = "Test status page"
  published   = true
}

data "uptimekuma_status_page" "test" {
  slug = uptimekuma_status_page.test.slug
}
`, slug, title)
}
