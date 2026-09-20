package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
	"github.com/breml/go-uptime-kuma-client/notification"
)

// testAccOutOfBandClient returns the dedicated out-of-band kuma client for use
// in acceptance tests. This client is separate from the provider's pooled
// connection, created once in TestMain, to genuinely simulate external
// modifications while avoiding Uptime Kuma's login rate limiting.
func testAccOutOfBandClient(t *testing.T) *kuma.Client {
	t.Helper()

	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skip("TF_ACC=1 not set")
	}

	if outOfBandClient == nil {
		t.Fatal("out-of-band client not initialized — TestMain did not run with TF_ACC=1")
	}

	return outOfBandClient
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
	kumaClient := testAccOutOfBandClient(t)

	resource.ParallelTest(t, resource.TestCase{
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
	kumaClient := testAccOutOfBandClient(t)

	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name     = %[1]q
  hostname = "8.8.8.8"
}
`, name)

	resource.ParallelTest(t, resource.TestCase{
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
	kumaClient := testAccOutOfBandClient(t)

	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}
`, name)

	resource.ParallelTest(t, resource.TestCase{
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
	kumaClient := testAccOutOfBandClient(t)

	resource.ParallelTest(t, resource.TestCase{
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

// testAccChangeMonitorTypeToHTTP changes a monitor's type to HTTP via the kuma
// API, simulating an external type change outside of Terraform.
func testAccChangeMonitorTypeToHTTP(
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

		// Get current monitor to copy base fields (ID, Name, Interval, etc.)
		var pingMonitor monitor.Ping
		getErr := kumaClient.GetMonitorAs(ctx, id, &pingMonitor)
		if getErr != nil {
			return fmt.Errorf("failed to get monitor: %w", getErr)
		}

		// Build an HTTP monitor with the same base fields but a different type.
		// HTTP.MarshalJSON hardcodes type="http", so UpdateMonitor changes the
		// server-side type regardless of the base's internalType.
		httpMonitor := monitor.HTTP{
			Base: pingMonitor.Base,
		}
		httpMonitor.URL = "https://httpbin.org/status/200"
		// Server requires accepted_statuscodes to be an array, not null.
		httpMonitor.AcceptedStatusCodes = []string{"200-299"}

		updateErr := kumaClient.UpdateMonitor(ctx, &httpMonitor)
		if updateErr != nil {
			return fmt.Errorf("failed to change monitor type to http: %w", updateErr)
		}

		return nil
	}
}

// TestAccMonitorPingResource_typeDrift verifies that when a ping monitor's type
// is changed externally in Uptime Kuma, the provider detects the drift and
// removes the resource from state, triggering a re-create on the next plan.
func TestAccMonitorPingResource_typeDrift(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPingTypeDrift")
	kumaClient := testAccOutOfBandClient(t)

	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name     = %[1]q
  hostname = "8.8.8.8"
}
`, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             config,
				Check:              testAccChangeMonitorTypeToHTTP(t, kumaClient, "uptimekuma_monitor_ping.test"),
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

// TestAccStatusPageResource_disappears verifies that when a status page is
// deleted externally, the provider removes it from state and plans to recreate it.
func TestAccStatusPageResource_disappears(t *testing.T) {
	slug := acctest.RandomWithPrefix("test-disappears")
	title := "Disappears Test Status Page"
	kumaClient := testAccOutOfBandClient(t)

	config := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_status_page" "test" {
  slug  = %[1]q
  title = %[2]q
}
`, slug, title)

	resource.ParallelTest(t, resource.TestCase{
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

// testAccDeleteNotificationExternally deletes a notification via the kuma API,
// simulating an external deletion outside of Terraform.
func testAccDeleteNotificationExternally(
	t *testing.T,
	kumaClient *kuma.Client,
	resourceAddr string,
) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		id, err := testAccResourceID(s, resourceAddr)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		deleteErr := kumaClient.DeleteNotification(ctx, id)
		if deleteErr != nil {
			return fmt.Errorf("failed to delete notification externally: %w", deleteErr)
		}

		return nil
	}
}

// testAccChangeNotificationTypeToGotify rewrites a notification as a Gotify one,
// keeping its ID and name, simulating a type change outside of Terraform.
func testAccChangeNotificationTypeToGotify(
	t *testing.T,
	kumaClient *kuma.Client,
	resourceAddr string,
) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		id, err := testAccResourceID(s, resourceAddr)
		if err != nil {
			return err
		}

		rs := s.RootModule().Resources[resourceAddr]

		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		gotify := notification.Gotify{
			Base: notification.Base{
				ID:       id,
				Name:     rs.Primary.Attributes["name"],
				IsActive: true,
			},
			GotifyDetails: notification.GotifyDetails{
				ServerURL:        "https://gotify.example.com",
				ApplicationToken: "token",
			},
		}

		updateErr := kumaClient.UpdateNotification(ctx, gotify)
		if updateErr != nil {
			return fmt.Errorf("failed to change notification type to gotify: %w", updateErr)
		}

		return nil
	}
}

// testAccResourceID reads the numeric id attribute of a resource from state.
func testAccResourceID(s *terraform.State, resourceAddr string) (int64, error) {
	rs, ok := s.RootModule().Resources[resourceAddr]
	if !ok {
		return 0, fmt.Errorf("resource %s not found in state", resourceAddr)
	}

	id, err := strconv.ParseInt(rs.Primary.Attributes["id"], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse id of %s: %w", resourceAddr, err)
	}

	return id, nil
}

// TestAccNotificationWebhookResource_disappears verifies that a notification
// deleted externally is removed from state and planned for re-creation. Every
// typed notification resource shares this read path, so it stands in for all of
// them.
func TestAccNotificationWebhookResource_disappears(t *testing.T) {
	name := acctest.RandomWithPrefix("TestWebhookDisappears")
	kumaClient := testAccOutOfBandClient(t)

	config := testAccNotificationWebhookResourceConfig(name, "https://example.com/hook", "json")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             config,
				Check:              testAccDeleteNotificationExternally(t, kumaClient, "uptimekuma_notification_webhook.test"),
				ExpectNonEmptyPlan: true,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"uptimekuma_notification_webhook.test",
							plancheck.ResourceActionCreate,
						),
					},
				},
			},
		},
	})
}

// TestAccNotificationWebhookResource_typeDrift verifies that a notification whose
// type was changed externally is reported rather than read into the wrong
// resource type, which would fill state with zero values and push them back on
// the next apply.
func TestAccNotificationWebhookResource_typeDrift(t *testing.T) {
	name := acctest.RandomWithPrefix("TestWebhookTypeDrift")
	kumaClient := testAccOutOfBandClient(t)

	config := testAccNotificationWebhookResourceConfig(name, "https://example.com/hook", "json")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// The drift happens in the check, between the apply and the plan
			// the framework runs right after it, so the refresh that reads the
			// notification back belongs to this step and so does its error.
			{
				Config: config,
				Check: testAccChangeNotificationTypeToGotify(
					t,
					kumaClient,
					"uptimekuma_notification_webhook.test",
				),
				ExpectError: regexp.MustCompile(`Incorrect notification type`),
			},
		},
	})
}
