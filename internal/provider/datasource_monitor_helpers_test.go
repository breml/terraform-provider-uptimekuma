package provider

import (
	"strings"
	"testing"

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

			detail := diags.Errors()[0].Detail()
			if !strings.Contains(detail, `Monitor ID 3 has type "ping", expected "http"`) {
				t.Errorf("want the detail to name the ID and both types, got %q", detail)
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
		})
	}
}
