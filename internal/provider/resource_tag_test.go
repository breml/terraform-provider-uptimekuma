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

func TestAccTagResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestTag")
	nameUpdated := acctest.RandomWithPrefix("TestTagUpdated")
	color := "#3498db"
	colorUpdated := "#2ecc71"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccTagResourceConfig(name, color),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_tag.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_tag.test",
						tfjsonpath.New("color"),
						knownvalue.StringExact(color),
					),
				},
			},
			{
				Config:             testAccTagResourceConfig(nameUpdated, colorUpdated),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_tag.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_tag.test",
						tfjsonpath.New("color"),
						knownvalue.StringExact(colorUpdated),
					),
				},
			},
		},
	})
}

func testAccTagResourceConfig(name string, color string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_tag" "test" {
  name  = %[1]q
  color = %[2]q
}
`, name, color)
}

func TestAccTagResourceDelete(t *testing.T) {
	name := acctest.RandomWithPrefix("TestTagDelete")
	color := "#9b59b6"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagResourceConfig(name, color),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_tag.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config:             testAccTagResourceConfigEmpty(),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccTagResourceConfigEmpty() string {
	return providerConfig()
}
