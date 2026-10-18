package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/breml/go-uptime-kuma-client/monitor"
)

func TestManualStatusMapping(t *testing.T) {
	// The numeric values are the Uptime Kuma encoding; pinning them catches a
	// mapping that is swapped consistently in both directions.
	tests := map[string]monitor.ManualStatus{
		"up":      1,
		"down":    0,
		"pending": 2,
	}

	for value, want := range tests {
		t.Run(value, func(t *testing.T) {
			got, ok := manualStatusFromString(value)
			if !ok || got != want {
				t.Fatalf("manualStatusFromString(%q) = %d, %t; want %d, true", value, got, ok, want)
			}

			back, ok := manualStatusToString(got)
			if !ok || back != value {
				t.Fatalf("manualStatusToString(%d) = %q, %t; want %q, true", got, back, ok, value)
			}
		})
	}

	t.Run("invalid string", func(t *testing.T) {
		_, ok := manualStatusFromString("UP")
		if ok {
			t.Fatal("manualStatusFromString(\"UP\") reported ok")
		}
	})

	t.Run("invalid constant", func(t *testing.T) {
		_, ok := manualStatusToString(monitor.ManualStatus(3))
		if ok {
			t.Fatal("manualStatusToString(3) reported ok")
		}
	})
}

// TestBuildManualMonitorStatus guards the unreturned status: Uptime Kuma
// resets the monitor to pending on any edit without a status, so the build
// helper must carry the planned status regardless of which attribute changed.
func TestBuildManualMonitorStatus(t *testing.T) {
	tests := map[string]struct {
		status types.String
		want   *monitor.ManualStatus
	}{
		"up":      {status: types.StringValue("up"), want: new(monitor.ManualStatusUp)},
		"down":    {status: types.StringValue("down"), want: new(monitor.ManualStatusDown)},
		"pending": {status: types.StringValue("pending"), want: new(monitor.ManualStatusPending)},
		"null":    {status: types.StringNull(), want: nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			data := &MonitorManualResourceModel{
				MonitorBaseModel: MonitorBaseModel{
					ID:              types.Int64Value(42),
					Name:            types.StringValue("manual"),
					Description:     types.StringNull(),
					Parent:          types.Int64Null(),
					Interval:        types.Int64Value(120),
					NotificationIDs: types.ListNull(types.Int64Type),
				},
				ManualStatus: tc.status,
			}

			var diags diag.Diagnostics

			got := buildManualMonitor(t.Context(), data, &diags)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if got.Interval != 120 {
				t.Errorf("Interval = %d, want 120", got.Interval)
			}

			switch {
			case tc.want == nil && got.ManualStatus != nil:
				t.Errorf("ManualStatus = %d, want nil", *got.ManualStatus)

			case tc.want != nil && got.ManualStatus == nil:
				t.Errorf("ManualStatus = nil, want %d", *tc.want)

			case tc.want != nil && *got.ManualStatus != *tc.want:
				t.Errorf("ManualStatus = %d, want %d", *got.ManualStatus, *tc.want)

			default:
				// Status matches.
			}
		})
	}
}

func TestBuildManualMonitorInvalidStatus(t *testing.T) {
	tests := map[string]types.String{
		"invalid": types.StringValue("maybe"),
		"unknown": types.StringUnknown(),
	}

	for name, status := range tests {
		t.Run(name, func(t *testing.T) {
			data := &MonitorManualResourceModel{
				MonitorBaseModel: MonitorBaseModel{
					Name:            types.StringValue("manual"),
					Description:     types.StringNull(),
					Parent:          types.Int64Null(),
					NotificationIDs: types.ListNull(types.Int64Type),
				},
				ManualStatus: status,
			}

			var diags diag.Diagnostics

			got := buildManualMonitor(t.Context(), data, &diags)
			if !diags.HasError() {
				t.Fatal("expected an error diagnostic")
			}

			if got != nil {
				t.Fatal("expected no monitor")
			}
		})
	}
}

func TestReadManualStatus(t *testing.T) {
	tests := map[string]struct {
		server      *monitor.ManualStatus
		current     types.String
		want        types.String
		wantWarning bool
	}{
		"not returned keeps state": {
			current: types.StringValue("down"),
			want:    types.StringValue("down"),
		},
		"not returned keeps null": {
			current: types.StringNull(),
			want:    types.StringNull(),
		},
		"returned is adopted": {
			server:  new(monitor.ManualStatusUp),
			current: types.StringValue("down"),
			want:    types.StringValue("up"),
		},
		"unrecognized keeps state and warns": {
			server:      new(monitor.ManualStatus(3)),
			current:     types.StringValue("down"),
			want:        types.StringValue("down"),
			wantWarning: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var diags diag.Diagnostics

			got := readManualStatus(t.Context(), 42, tc.server, tc.current, &diags)
			if !got.Equal(tc.want) {
				t.Errorf("readManualStatus() = %s, want %s", got, tc.want)
			}

			if diags.HasError() {
				t.Fatalf("unexpected error diagnostics: %v", diags)
			}

			if gotWarning := diags.WarningsCount() > 0; gotWarning != tc.wantWarning {
				t.Errorf("warning = %t, want %t (%v)", gotWarning, tc.wantWarning, diags)
			}
		})
	}
}

// TestAccMonitorManualResource covers the create, import and update round-trip.
//
// Uptime Kuma never returns manual_status, so the import step cannot verify
// it, and no step checks the status the server actually reports.
func TestAccMonitorManualResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestManualMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestManualMonitorUpdated")
	description := "Test manual monitor with description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorManualResourceConfig(name, "up", 60, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("manual_status"),
						knownvalue.StringExact("up"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_monitor_manual.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manual_status"},
			},
			{
				Config: testAccMonitorManualResourceConfig(nameUpdated, "down", 120, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("manual_status"),
						knownvalue.StringExact("down"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("description"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
				},
			},
		},
	})
}

func testAccMonitorManualResourceConfig(
	name string, status string,
	interval int64, description string,
) string {
	statusField := ""
	if status != "" {
		statusField = fmt.Sprintf("  manual_status = %q", status)
	}

	descField := ""
	if description != "" {
		descField = fmt.Sprintf("  description   = %q", description)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_manual" "test" {
  name          = %[1]q
%[2]s
%[3]s
  interval      = %[4]d
  active        = true
}
`, name, statusField, descField, interval)
}

// TestAccMonitorManualResourceMinimal verifies the shared monitor defaults
// apply and that an unset status stays null without a perpetual diff.
func TestAccMonitorManualResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestManualMonitorMinimal")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorManualResourceConfigMinimal(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("manual_status"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config:   testAccMonitorManualResourceConfigMinimal(name),
				PlanOnly: true,
			},
		},
	})
}

func testAccMonitorManualResourceConfigMinimal(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_manual" "test" {
  name = %[1]q
}
`, name)
}

// TestAccMonitorManualResourceStatusTransitions switches the status through
// every value and back to unset, and finishes with an update that changes
// only the interval. The server never returns the status, so these steps can
// only prove that each apply succeeds and settles without a diff; that the
// interval-only update still sends the status is guarded by
// TestBuildManualMonitorStatus.
func TestAccMonitorManualResourceStatusTransitions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestManualMonitorStatus")

	steps := make([]resource.TestStep, 0, 5)
	for _, status := range []string{"up", "down", "pending", ""} {
		var want knownvalue.Check = knownvalue.Null()
		if status != "" {
			want = knownvalue.StringExact(status)
		}

		steps = append(steps, resource.TestStep{
			Config: testAccMonitorManualResourceConfig(name, status, 60, ""),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(
					"uptimekuma_monitor_manual.test",
					tfjsonpath.New("manual_status"),
					want,
				),
			},
		})
	}

	steps = append(steps,
		resource.TestStep{
			Config: testAccMonitorManualResourceConfig(name, "up", 60, ""),
		},
		resource.TestStep{
			Config: testAccMonitorManualResourceConfig(name, "up", 90, ""),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(
					"uptimekuma_monitor_manual.test",
					tfjsonpath.New("manual_status"),
					knownvalue.StringExact("up"),
				),
				statecheck.ExpectKnownValue(
					"uptimekuma_monitor_manual.test",
					tfjsonpath.New("interval"),
					knownvalue.Int64Exact(90),
				),
			},
		},
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
}

// TestAccMonitorManualResourceWithAllOptions verifies every supported
// attribute, including the base monitor fields (parent, notification_ids,
// tags) that every monitor type copies into its own build and populate
// helpers rather than sharing.
func TestAccMonitorManualResourceWithAllOptions(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestManualMonitorGroup")
	notificationName := acctest.RandomWithPrefix("TestManualMonitorNotification")
	tagName := acctest.RandomWithPrefix("TestManualMonitorTag")
	name := acctest.RandomWithPrefix("TestManualMonitorFull")
	description := "Full manual monitor test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorManualResourceConfigWithAllOptions(
					groupName, notificationName, tagName, name, description,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("manual_status"),
						knownvalue.StringExact("down"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("notification_ids"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("tags"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("tags").AtSliceIndex(0).AtMapKey("value"),
						knownvalue.StringExact("production"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(90),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(10),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_manual.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_monitor_manual.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manual_status"},
			},
		},
	})
}

func testAccMonitorManualResourceConfigWithAllOptions(
	groupName string,
	notificationName string,
	tagName string,
	name string,
	description string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_notification_webhook" "test" {
  name        = %[2]q
  webhook_url = "https://example.com/webhook"
  is_active   = true
}

resource "uptimekuma_tag" "test" {
  name  = %[3]q
  color = "#00ff00"
}

resource "uptimekuma_monitor_manual" "test" {
  name             = %[4]q
  description      = %[5]q
  manual_status    = "down"
  interval         = 120
  retry_interval   = 90
  resend_interval  = 10
  max_retries      = 5
  upside_down      = true
  active           = false
  parent           = uptimekuma_monitor_group.test.id
  notification_ids = [uptimekuma_notification_webhook.test.id]

  tags = [
    {
      tag_id = uptimekuma_tag.test.id
      value  = "production"
    },
  ]
}
`, groupName, notificationName, tagName, name, description)
}

// TestAccMonitorManualResourceInvalidStatus verifies the enum validator is
// wired into the schema.
func TestAccMonitorManualResourceInvalidStatus(t *testing.T) {
	name := acctest.RandomWithPrefix("TestManualMonitorInvalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMonitorManualResourceConfig(name, "maintenance", 60, ""),
				ExpectError: regexp.MustCompile(`value must be one of`),
			},
		},
	})
}
