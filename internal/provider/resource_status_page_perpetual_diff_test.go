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

// TestAccStatusPageNoPerpetualDiff verifies that re-applying the same config
// with public_group_list (including explicit send_url=false and weight) does
// not produce a perpetual diff. This reproduces the remaining issue from #223
// where the server response omits optional fields, causing state to diverge
// from config.
func TestAccStatusPageNoPerpetualDiff(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-nodiff")
	title := "No Perpetual Diff Test"
	monitorName1 := acctest.RandomWithPrefix("test-mon1")
	monitorName2 := acctest.RandomWithPrefix("test-mon2")

	config := testAccStatusPageNoPerpetualDiffConfig(slug, title, monitorName1, monitorName2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             config,
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("public_group_list").AtSliceIndex(0).AtMapKey("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("public_group_list").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("Servers"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("public_group_list").AtSliceIndex(0).AtMapKey("weight"),
						knownvalue.Int64Exact(1),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("public_group_list").AtSliceIndex(1).AtMapKey("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("public_group_list").AtSliceIndex(1).AtMapKey("name"),
						knownvalue.StringExact("Services"),
					),
				},
			},
			{
				Config:             config,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:             config,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccStatusPageNoPerpetualDiffConfig(
	slug string,
	title string,
	monitorName1 string,
	monitorName2 string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "mon1" {
  name = %[3]q
  url  = "https://example.com"
}

resource "uptimekuma_monitor_http" "mon2" {
  name = %[4]q
  url  = "https://example.org"
}

resource "uptimekuma_status_page" "test" {
  slug      = %[1]q
  title     = %[2]q
  published = true

  public_group_list = [
    {
      name   = "Servers"
      weight = 1
      monitor_list = [
        {
          id       = uptimekuma_monitor_http.mon1.id
          send_url = false
        }
      ]
    },
    {
      name   = "Services"
      weight = 2
      monitor_list = [
        {
          id       = uptimekuma_monitor_http.mon2.id
          send_url = false
        }
      ]
    }
  ]
}
`, slug, title, monitorName1, monitorName2)
}
