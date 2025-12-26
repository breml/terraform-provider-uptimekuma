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

func TestAccProxyDataSourceByID(t *testing.T) {
	name := acctest.RandomWithPrefix("TestProxy")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_proxy.test",
						tfjsonpath.New("host"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccProxyDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_proxy" "test" {
  protocol = "http"
  host     = %[1]q
  port     = 8080
  active   = true
}

data "uptimekuma_proxy" "test" {
  id = uptimekuma_proxy.test.id
}
`, name)
}
