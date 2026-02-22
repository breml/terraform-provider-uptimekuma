package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// testAccNewKumaClient creates a new kuma client for use in acceptance tests.
// This is used to perform out-of-band operations (e.g. deleting resources)
// to simulate external modifications.
func testAccNewKumaClient(t *testing.T) *kuma.Client {
	t.Helper()

	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skip("TF_ACC=1 not set")
	}

	kumaClient, err := kuma.New(t.Context(), endpoint, username, password)
	if err != nil {
		t.Fatalf("failed to create kuma client: %v", err)
	}

	t.Cleanup(func() {
		_ = kumaClient.Disconnect()
	})

	return kumaClient
}

// testAccDeleteMonitorExternally deletes a monitor via the kuma API, simulating
// an external deletion outside of Terraform.
func testAccDeleteMonitorExternally(
	t *testing.T,
	kumaClient *kuma.Client,
	resourceAddr string,
) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceAddr)
		}

		id, err := strconv.ParseInt(rs.Primary.Attributes["id"], 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse monitor id: %w", err)
		}

		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		deleteErr := kumaClient.DeleteMonitor(ctx, id)
		if deleteErr != nil {
			return fmt.Errorf("failed to delete monitor externally: %w", deleteErr)
		}

		return nil
	}
}

// TestAccMonitorHTTPResource_disappears verifies that when an HTTP monitor is
// deleted externally (outside Terraform), the provider removes it from state
// and plans to recreate it.
func TestAccMonitorHTTPResource_disappears(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPDisappears")
	url := "https://httpbin.org/status/200"
	kumaClient := testAccNewKumaClient(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorHTTPResourceConfig(name, url, "GET", 60, 48),
				Check:              testAccDeleteMonitorExternally(t, kumaClient, "uptimekuma_monitor_http.test"),
				ExpectNonEmptyPlan: true,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"uptimekuma_monitor_http.test",
							plancheck.ResourceActionCreate,
						),
					},
				},
			},
		},
	})
}

// TestAccMonitorPingResource_disappears verifies that when a Ping monitor is
// deleted externally, the provider removes it from state and plans to recreate it.
func TestAccMonitorPingResource_disappears(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPingDisappears")
	kumaClient := testAccNewKumaClient(t)

	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name     = %[1]q
  hostname = "8.8.8.8"
}
`, name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             config,
				Check:              testAccDeleteMonitorExternally(t, kumaClient, "uptimekuma_monitor_ping.test"),
				ExpectNonEmptyPlan: true,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"uptimekuma_monitor_ping.test",
							plancheck.ResourceActionCreate,
						),
					},
				},
			},
		},
	})
}

// TestAccMonitorGroupResource_disappears verifies that when a monitor group is
// deleted externally, the provider removes it from state and plans to recreate it.
func TestAccMonitorGroupResource_disappears(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGroupDisappears")
	kumaClient := testAccNewKumaClient(t)

	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}
`, name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             config,
				Check:              testAccDeleteMonitorExternally(t, kumaClient, "uptimekuma_monitor_group.test"),
				ExpectNonEmptyPlan: true,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"uptimekuma_monitor_group.test",
							plancheck.ResourceActionCreate,
						),
					},
				},
			},
		},
	})
}

// TestAccTagResource_disappears verifies that when a tag is deleted externally,
// the provider removes it from state and plans to recreate it.
func TestAccTagResource_disappears(t *testing.T) {
	name := acctest.RandomWithPrefix("TestTagDisappears")
	color := "#3498db"
	kumaClient := testAccNewKumaClient(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccTagResourceConfig(name, color),
				ExpectNonEmptyPlan: true,
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources["uptimekuma_tag.test"]
					if !ok {
						return errors.New("resource uptimekuma_tag.test not found in state")
					}

					id, err := strconv.ParseInt(rs.Primary.Attributes["id"], 10, 64)
					if err != nil {
						return fmt.Errorf("failed to parse tag id: %w", err)
					}

					ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
					defer cancel()

					deleteErr := kumaClient.DeleteTag(ctx, id)
					if deleteErr != nil {
						return fmt.Errorf("failed to delete tag externally: %w", deleteErr)
					}

					return nil
				},
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"uptimekuma_tag.test",
							plancheck.ResourceActionCreate,
						),
					},
				},
			},
		},
	})
}

// TestAccStatusPageResource_disappears verifies that when a status page is
// deleted externally, the provider removes it from state and plans to recreate it.
func TestAccStatusPageResource_disappears(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-disappears")
	title := "Disappears Test Status Page"
	kumaClient := testAccNewKumaClient(t)

	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_status_page" "test" {
  slug  = %[1]q
  title = %[2]q
}
`, slug, title)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             config,
				ExpectNonEmptyPlan: true,
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources["uptimekuma_status_page.test"]
					if !ok {
						return errors.New("resource uptimekuma_status_page.test not found in state")
					}

					slugVal := rs.Primary.Attributes["slug"]

					ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
					defer cancel()

					err := kumaClient.DeleteStatusPage(ctx, slugVal)
					if err != nil {
						return fmt.Errorf("failed to delete status page externally: %w", err)
					}

					return nil
				},
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"uptimekuma_status_page.test",
							plancheck.ResourceActionCreate,
						),
					},
				},
			},
		},
	})
}
