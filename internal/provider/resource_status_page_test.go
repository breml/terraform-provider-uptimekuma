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

func TestAccStatusPageResource(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-status")
	title := "Test Status Page"
	titleUpdated := "Updated Test Status Page"
	description := "Test status page description"
	descriptionUpdated := "Updated test status page description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccStatusPageResourceConfig(slug, title, description, true),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("slug"), knownvalue.StringExact(slug)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("title"), knownvalue.StringExact(title)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("published"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccStatusPageResourceConfig(slug, titleUpdated, descriptionUpdated, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("slug"), knownvalue.StringExact(slug)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("title"), knownvalue.StringExact(titleUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("description"), knownvalue.StringExact(descriptionUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("published"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func testAccStatusPageResourceConfig(slug, title, description string, published bool) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_status_page" "test" {
  slug        = %[1]q
  title       = %[2]q
  description = %[3]q
  published   = %[4]t
}
`, slug, title, description, published)
}

func TestAccStatusPageResourceMinimal(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-minimal")
	title := "Minimal Status Page"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageResourceConfigMinimal(slug, title),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("slug"), knownvalue.StringExact(slug)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("title"), knownvalue.StringExact(title)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("published"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("show_tags"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("show_powered_by"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("show_certificate_expiry"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func testAccStatusPageResourceConfigMinimal(slug, title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_status_page" "test" {
  slug  = %[1]q
  title = %[2]q
}
`, slug, title)
}

func TestAccStatusPageResourceWithAllOptions(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-full")
	title := "Full Status Page"
	description := "Full test status page"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageResourceConfigWithAllOptions(slug, title, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("slug"), knownvalue.StringExact(slug)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("title"), knownvalue.StringExact(title)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("theme"), knownvalue.StringExact("light")),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("show_tags"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("show_powered_by"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("show_certificate_expiry"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func testAccStatusPageResourceConfigWithAllOptions(slug, title, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_status_page" "test" {
  slug                    = %[1]q
  title                   = %[2]q
  description             = %[3]q
  theme                   = "light"
  published               = true
  show_tags               = true
  show_powered_by         = false
  show_certificate_expiry = true
  footer_text             = "© 2024 Test Company"
  custom_css              = "body { background: #f0f0f0; }"
}
`, slug, title, description)
}

func TestAccStatusPageResourceWithMonitors(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-monitors")
	title := "Status Page with Monitors"
	monitorName := acctest.RandomWithPrefix("test-monitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageResourceConfigWithMonitors(slug, title, monitorName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("slug"), knownvalue.StringExact(slug)),
					statecheck.ExpectKnownValue("uptimekuma_status_page.test", tfjsonpath.New("title"), knownvalue.StringExact(title)),
				},
			},
		},
	})
}

func testAccStatusPageResourceConfigWithMonitors(slug, title, monitorName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name = %[3]q
  url  = "https://example.com"
}

resource "uptimekuma_status_page" "test" {
  slug      = %[1]q
  title     = %[2]q
  published = true

  public_group_list = [
    {
      name   = "Production Services"
      weight = 1
      monitor_list = [
        {
          id       = uptimekuma_monitor_http.test.id
          send_url = false
        }
      ]
    }
  ]
}
`, slug, title, monitorName)
}
