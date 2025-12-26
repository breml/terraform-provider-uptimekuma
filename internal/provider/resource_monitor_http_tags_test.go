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

func TestAccMonitorHTTPResourceWithTags(t *testing.T) {
	monitorName := acctest.RandomWithPrefix("TestHTTPMonitorWithTags")
	tagName1 := acctest.RandomWithPrefix("TestTag1")
	tagName2 := acctest.RandomWithPrefix("TestTag2")
	url := "https://httpbin.org/status/200"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create monitor with one tag
				Config: testAccMonitorHTTPResourceConfigWithOneTag(monitorName, url, tagName1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(monitorName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					// Verify tags list has 1 element
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("tags"),
						knownvalue.ListSizeExact(1)),
					// Verify the tag has a non-zero tag_id
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test",
						tfjsonpath.New("tags").AtSliceIndex(0).AtMapKey("tag_id"),
						knownvalue.NotNull()),
					// Verify the tag value is null when not provided
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test",
						tfjsonpath.New("tags").AtSliceIndex(0).AtMapKey("value"),
						knownvalue.Null()),
				},
			},
			{
				// Update monitor to add a second tag
				Config: testAccMonitorHTTPResourceConfigWithTwoTags(monitorName, url, tagName1, tagName2),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify tags list now has 2 elements
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("tags"),
						knownvalue.ListSizeExact(2)),
				},
			},
			{
				// Update monitor to remove first tag (keep only second)
				Config: testAccMonitorHTTPResourceConfigWithSecondTagOnly(monitorName, url, tagName2),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify tags list has 1 element again
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("tags"),
						knownvalue.ListSizeExact(1)),
				},
			},
			{
				// Import and verify - this verifies tags with values are imported correctly
				ResourceName:      "uptimekuma_monitor_http.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMonitorHTTPResourceWithTagsImport(t *testing.T) {
	monitorName := acctest.RandomWithPrefix("TestHTTPMonitorImport")
	tagName := acctest.RandomWithPrefix("TestTag")
	url := "https://httpbin.org/status/200"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create monitor with tag (no value)
				Config: testAccMonitorHTTPResourceConfigWithOneTag(monitorName, url, tagName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("tags"),
						knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test",
						tfjsonpath.New("tags").AtSliceIndex(0).AtMapKey("value"),
						knownvalue.Null()),
				},
			},
			{
				// Import and verify null values are handled correctly
				ResourceName:      "uptimekuma_monitor_http.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorHTTPResourceConfigWithOneTag(monitorName, url, tagName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_tag" "test1" {
  name  = %[3]q
  color = "#00ff00"
}

resource "uptimekuma_monitor_http" "test" {
  name = %[1]q
  url  = %[2]q

  tags = [
    {
      tag_id = uptimekuma_tag.test1.id
    },
  ]
}
`, monitorName, url, tagName)
}

func testAccMonitorHTTPResourceConfigWithTwoTags(monitorName, url, tagName1, tagName2 string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_tag" "test1" {
  name  = %[3]q
  color = "#00ff00"
}

resource "uptimekuma_tag" "test2" {
  name  = %[4]q
  color = "#ff0000"
}

resource "uptimekuma_monitor_http" "test" {
  name = %[1]q
  url  = %[2]q

  tags = [
    {
      tag_id = uptimekuma_tag.test1.id
    },
    {
      tag_id = uptimekuma_tag.test2.id
      value  = "v2"
    },
  ]
}
`, monitorName, url, tagName1, tagName2)
}

func testAccMonitorHTTPResourceConfigWithSecondTagOnly(monitorName, url, tagName2 string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_tag" "test2" {
  name  = %[3]q
  color = "#ff0000"
}

resource "uptimekuma_monitor_http" "test" {
  name = %[1]q
  url  = %[2]q

  tags = [
    {
      tag_id = uptimekuma_tag.test2.id
      value  = "v2"
    },
  ]
}
`, monitorName, url, tagName2)
}
