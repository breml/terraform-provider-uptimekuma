package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/breml/go-uptime-kuma-client/monitor"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMonitorTypeMatches(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		actual      string
		wantMatches bool
	}{
		"a monitor of the served type matches": {
			actual:      "http",
			wantMatches: true,
		},
		// Base.UnmarshalJSON leaves internalType empty for a record that carried no
		// type, which is not evidence of another type.
		"a record with no type matches": {
			actual:      "",
			wantMatches: true,
		},
		"a monitor of another type does not match": {
			actual: "ping",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			got := monitorTypeMatches(&diags, 3, test.actual, "http")
			if got != test.wantMatches {
				t.Fatalf("want matches %t, got %t", test.wantMatches, got)
			}

			if diags.HasError() == test.wantMatches {
				t.Fatalf("want error %t, got %v", !test.wantMatches, diags.Errors())
			}

			if test.wantMatches {
				return
			}

			diagnostic := diags.Errors()[0]
			if diagnostic.Summary() != "Monitor type mismatch" {
				t.Errorf("want summary %q, got %q", "Monitor type mismatch", diagnostic.Summary())
			}

			if !strings.Contains(diagnostic.Detail(), `Monitor ID 3 has type "ping", expected "http"`) {
				t.Errorf("want the detail to name the ID and both types, got %q", diagnostic.Detail())
			}
		})
	}
}

func TestDataSourceNameMatches(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configured  types.String
		wantMatches bool
	}{
		"a lookup that states no name matches": {
			configured:  types.StringNull(),
			wantMatches: true,
		},
		"a name that is not known yet matches": {
			configured:  types.StringUnknown(),
			wantMatches: true,
		},
		"a name that agrees with the resolved one matches": {
			configured:  types.StringValue("ntp-pool"),
			wantMatches: true,
		},
		"a name that contradicts the resolved one does not match": {
			configured: types.StringValue("other"),
		},
		// An empty string is a stated name like any other, not an absent one.
		"an empty name does not match a named entity": {
			configured: types.StringValue(""),
		},
		// Uptime Kuma treats names as opaque, so the comparison is exact.
		"a name that differs only in case does not match": {
			configured: types.StringValue("NTP-Pool"),
		},
		"a name that differs only in whitespace does not match": {
			configured: types.StringValue("ntp-pool "),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			got := dataSourceNameMatches(&diags, "Monitor", 5, test.configured, "ntp-pool")
			if got != test.wantMatches {
				t.Fatalf("want matches %t, got %t", test.wantMatches, got)
			}

			if diags.HasError() == test.wantMatches {
				t.Fatalf("want error %t, got %v", !test.wantMatches, diags.Errors())
			}

			if test.wantMatches {
				return
			}

			diagnostic := diags.Errors()[0]
			if diagnostic.Summary() != "Monitor name mismatch" {
				t.Errorf("want summary %q, got %q", "Monitor name mismatch", diagnostic.Summary())
			}

			if !strings.Contains(diagnostic.Detail(), `Monitor ID 5 is named "ntp-pool"`) {
				t.Errorf("want the detail to name the ID and the resolved name, got %q", diagnostic.Detail())
			}

			wantConfigured := fmt.Sprintf("expects %q", test.configured.ValueString())
			if !strings.Contains(diagnostic.Detail(), wantConfigured) {
				t.Errorf("want the detail to name the configured name, got %q", diagnostic.Detail())
			}
		})
	}
}

// TestDataSourceReadByIDGuards scans the monitor and notification data sources
// for the guards their readByID must apply. The data sources are cloned from
// one another, so a guard missing from one would otherwise only be noticed by
// the acceptance test of that very type, if it has one.
func TestDataSourceReadByIDGuards(t *testing.T) {
	t.Parallel()

	dir := os.DirFS(".")

	var files []string

	for _, pattern := range []string{"data_source_monitor_*.go", "data_source_notification*.go"} {
		matches, err := fs.Glob(dir, pattern)
		if err != nil {
			t.Fatalf("glob %q: %v", pattern, err)
		}

		for _, match := range matches {
			if !strings.HasSuffix(match, "_test.go") {
				files = append(files, match)
			}
		}
	}

	// A rename of the files must not turn the scan into one that checks nothing.
	if len(files) == 0 {
		t.Fatal("found no data source files to scan")
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			content, err := fs.ReadFile(dir, file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}

			src := string(content)

			nameGuard := soleOffset(t, src, "dataSourceNameMatches(")

			// A data source that assigns the name elsewhere has no offset to compare.
			assign := strings.Index(src, "data.Name = ")
			if assign >= 0 && nameGuard > assign {
				t.Error("dataSourceNameMatches is called after data.Name is overwritten")
			}

			if !strings.HasPrefix(file, "data_source_monitor_") {
				return
			}

			typeGuard := soleOffset(t, src, "monitorTypeMatches(")
			if typeGuard > nameGuard {
				t.Error("monitorTypeMatches is called after dataSourceNameMatches")
			}
		})
	}
}

// soleOffset returns the offset of the only occurrence of call in src, and
// fails the test when there is none or more than one.
func soleOffset(t *testing.T, src string, call string) int {
	t.Helper()

	count := strings.Count(src, call)
	if count != 1 {
		t.Fatalf("want exactly one %s call, got %d", strings.TrimSuffix(call, "("), count)
	}

	return strings.Index(src, call)
}

// decodedMonitor builds a monitor.Base the way the client does, from the server's
// JSON, because Type() reads a field only UnmarshalJSON sets.
func decodedMonitor(t *testing.T, id int64, name string, monitorType string) monitor.Base {
	t.Helper()

	var mon monitor.Base

	raw := fmt.Sprintf(`{"id":%d,"name":%q,"type":%q}`, id, name, monitorType)
	err := json.Unmarshal([]byte(raw), &mon)
	if err != nil {
		t.Fatalf("decode monitor: %v", err)
	}

	return mon
}

func TestFindMonitorByName(t *testing.T) {
	t.Parallel()

	web := decodedMonitor(t, 7, "web", "http")
	other := decodedMonitor(t, 9, "other", "http")
	wrongType := decodedMonitor(t, 11, "web", "ping")
	twin := decodedMonitor(t, 12, "web", "http")

	tests := map[string]struct {
		client      *fakeMonitorLister
		wantID      int64
		wantSummary string
		wantCalls   int
	}{
		"a listed monitor is found after one look": {
			client:    &fakeMonitorLister{lists: [][]monitor.Base{{other, web}}},
			wantID:    7,
			wantCalls: 1,
		},
		// The first answer was an unrequested monitorList older than the monitor,
		// which GetMonitors returns because it matches the broadcast by name alone.
		"a monitor only the second look lists is found": {
			client:    &fakeMonitorLister{lists: [][]monitor.Base{{other}, {other, web}}},
			wantID:    7,
			wantCalls: 2,
		},
		"a monitor of another type is not found after a second look": {
			client:      &fakeMonitorLister{lists: [][]monitor.Base{{other, wrongType}}},
			wantSummary: "http monitor not found",
			wantCalls:   2,
		},
		// An ambiguous name is an answer, not a miss, so it is not looked at twice.
		"two monitors of the name are an error after one look": {
			client:      &fakeMonitorLister{lists: [][]monitor.Base{{web, twin}}},
			wantSummary: "Multiple monitors found",
			wantCalls:   1,
		},
		"a list that will not answer is an error": {
			client:      &fakeMonitorLister{err: errors.New("socket closed")},
			wantSummary: "failed to read monitors",
			wantCalls:   1,
		},
		"a second look that fails is an error, not a miss": {
			client: &fakeMonitorLister{
				lists: [][]monitor.Base{{other}},
				err:   errors.New("connection reset"),
				errOn: 2,
			},
			wantSummary: "failed to read monitors",
			wantCalls:   2,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			got := findMonitorByName(t.Context(), test.client, "web", "http", &diags)

			if test.wantSummary == "" {
				if diags.HasError() {
					t.Fatalf("want no error, got %v", diags.Errors())
				}

				if got == nil || got.GetID() != test.wantID {
					t.Fatalf("want monitor %d, got %v", test.wantID, got)
				}
			} else {
				if got != nil {
					t.Errorf("want no monitor, got %d", got.GetID())
				}

				if diags.ErrorsCount() != 1 {
					t.Fatalf("want 1 error, got %v", diags.Errors())
				}

				if summary := diags.Errors()[0].Summary(); summary != test.wantSummary {
					t.Errorf("want summary %q, got %q", test.wantSummary, summary)
				}
			}

			if test.client.calls != test.wantCalls {
				t.Errorf("want %d list calls, got %d", test.wantCalls, test.client.calls)
			}
		})
	}
}
