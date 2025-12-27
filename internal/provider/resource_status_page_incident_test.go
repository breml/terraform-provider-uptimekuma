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

func TestAccStatusPageIncidentResource(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-incident")
	statusPageTitle := "Test Status Page"
	incidentTitle := "Test Incident"
	incidentContent := "This is a test incident"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageIncidentResourceConfig(
					slug,
					statusPageTitle,
					incidentTitle,
					incidentContent,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(incidentTitle),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("content"),
						knownvalue.StringExact(incidentContent),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("pin"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccStatusPageIncidentResourceConfig(
	slug string,
	statusPageTitle string,
	incidentTitle string,
	incidentContent string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_status_page" "test" {
  slug      = %[1]q
  title     = %[2]q
  published = true
}

resource "uptimekuma_status_page_incident" "test" {
  status_page_slug = uptimekuma_status_page.test.slug
  title            = %[3]q
  content          = %[4]q
}
`, slug, statusPageTitle, incidentTitle, incidentContent)
}

func TestAccStatusPageIncidentResourceWithStyle(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-incident-style")
	statusPageTitle := "Test Status Page"
	incidentTitle := "Maintenance Window"
	incidentContent := "Scheduled maintenance"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageIncidentResourceConfigWithStyle(
					slug,
					statusPageTitle,
					incidentTitle,
					incidentContent,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(incidentTitle),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("content"),
						knownvalue.StringExact(incidentContent),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("style"),
						knownvalue.StringExact("info"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("pin"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccStatusPageIncidentResourceConfigWithStyle(
	slug string, statusPageTitle string, incidentTitle string, incidentContent string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_status_page" "test" {
  slug      = %[1]q
  title     = %[2]q
  published = true
}

resource "uptimekuma_status_page_incident" "test" {
  status_page_slug = uptimekuma_status_page.test.slug
  title            = %[3]q
  content          = %[4]q
  style            = "info"
  pin              = true
}
`, slug, statusPageTitle, incidentTitle, incidentContent)
}

func TestAccStatusPageIncidentResourceUpdate(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-incident-update")
	statusPageTitle := "Test Status Page"
	incidentTitle := "Initial Incident"
	incidentTitleUpdated := "Updated Incident"
	incidentContent := "Initial content"
	incidentContentUpdated := "Updated content"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageIncidentResourceConfigWithStyle(
					slug,
					statusPageTitle,
					incidentTitle,
					incidentContent,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(incidentTitle),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("content"),
						knownvalue.StringExact(incidentContent),
					),
				},
			},
			{
				Config: testAccStatusPageIncidentResourceConfigWithStyle(
					slug,
					statusPageTitle,
					incidentTitleUpdated,
					incidentContentUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(incidentTitleUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page_incident.test",
						tfjsonpath.New("content"),
						knownvalue.StringExact(incidentContentUpdated),
					),
				},
			},
		},
	})
}
