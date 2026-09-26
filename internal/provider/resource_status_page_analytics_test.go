package provider

import (
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/breml/go-uptime-kuma-client/statuspage"
)

// TestStatusPageAnalyticsTypes pins the analytics allowlist exactly. Asserting only that every
// entry passes statuspage.ValidAnalyticsType would be close to a tautology, because the entries
// are built from the very helpers that function switches on, and it could not fail if an entry
// were dropped - which would reject a value users configure today.
func TestStatusPageAnalyticsTypes(t *testing.T) {
	t.Parallel()

	want := []string{"google", "umami", "plausible", "matomo", "rybbit"}

	got := analyticsTypes()
	if !slices.Equal(got, want) {
		t.Fatalf("analyticsTypes() = %v, want %v", got, want)
	}

	for _, analyticsType := range want {
		if !statuspage.ValidAnalyticsType(&analyticsType) {
			t.Errorf("analyticsTypes() offers %q, which the client rejects", analyticsType)
		}
	}
}

// analyticsTypeOneOfValidator returns the allowlist validator the status page resource declares
// for analytics_type.
//
// The attribute also carries stringvalidator.ConflictsWith, which is deliberately left out:
// for a non-null value it resolves its path expressions against req.Config, so calling it with
// the zero value validator.StringRequest the table below uses would dereference a nil schema.
func analyticsTypeOneOfValidator(t *testing.T) validator.String {
	t.Helper()

	resp := &resource.SchemaResponse{}
	(&StatusPageResource{}).Schema(t.Context(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["analytics_type"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("analytics_type is %T, want schema.StringAttribute", resp.Schema.Attributes["analytics_type"])
	}

	for _, v := range attr.Validators {
		if strings.HasPrefix(v.Description(t.Context()), "value must be one of") {
			return v
		}
	}

	t.Fatal("analytics_type declares no allowlist validator")

	return nil
}

func TestStatusPageAnalyticsTypeValidation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     types.String
		wantError bool
	}{
		"google":            {value: types.StringValue("google"), wantError: false},
		"umami":             {value: types.StringValue("umami"), wantError: false},
		"plausible":         {value: types.StringValue("plausible"), wantError: false},
		"matomo":            {value: types.StringValue("matomo"), wantError: false},
		"rybbit":            {value: types.StringValue("rybbit"), wantError: false},
		"null":              {value: types.StringNull(), wantError: false},
		"unknown":           {value: types.StringUnknown(), wantError: false},
		"empty":             {value: types.StringValue(""), wantError: true},
		"wrong case":        {value: types.StringValue("Rybbit"), wantError: true},
		"trailing space":    {value: types.StringValue("rybbit "), wantError: true},
		"unsupported":       {value: types.StringValue("fathom"), wantError: true},
		"deprecated google": {value: types.StringValue("google-analytics"), wantError: true},
	}

	oneOf := analyticsTypeOneOfValidator(t)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.StringResponse{}
			oneOf.ValidateString(
				t.Context(),
				validator.StringRequest{ConfigValue: test.value},
				resp,
			)

			if got := resp.Diagnostics.HasError(); got != test.wantError {
				t.Errorf("HasError() = %v, want %v (%v)", got, test.wantError, resp.Diagnostics)
			}
		})
	}
}

// TestValidateStatusPageAnalytics covers the companion attributes each analytics type needs.
// Uptime Kuma accepts an incomplete configuration on write and then renders a snippet that does
// not track anything, so the provider has to reject it at plan time.
func TestValidateStatusPageAnalytics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		analyticsType     types.String
		analyticsID       types.String
		analyticsScript   types.String
		expectedErrors    []string
		expectedSummaries []string
	}{
		{
			name:            "type omitted, which is also the deprecated google_analytics_id path",
			analyticsType:   types.StringNull(),
			analyticsID:     types.StringNull(),
			analyticsScript: types.StringNull(),
		},
		{
			name:            "type only known after apply",
			analyticsType:   types.StringUnknown(),
			analyticsID:     types.StringNull(),
			analyticsScript: types.StringNull(),
		},
		{
			name:            "google with a tracking ID",
			analyticsType:   types.StringValue("google"),
			analyticsID:     types.StringValue("G-TEST12345"),
			analyticsScript: types.StringNull(),
		},
		{
			name:              "google without a tracking ID",
			analyticsType:     types.StringValue("google"),
			analyticsID:       types.StringNull(),
			analyticsScript:   types.StringNull(),
			expectedErrors:    []string{"analytics_id"},
			expectedSummaries: []string{"Missing Attribute Configuration"},
		},
		{
			name:            "rybbit with both companions",
			analyticsType:   types.StringValue("rybbit"),
			analyticsID:     types.StringValue("site-1"),
			analyticsScript: types.StringValue("https://app.rybbit.io/api/script.js"),
		},
		{
			name:              "rybbit without a site ID",
			analyticsType:     types.StringValue("rybbit"),
			analyticsID:       types.StringNull(),
			analyticsScript:   types.StringValue("https://app.rybbit.io/api/script.js"),
			expectedErrors:    []string{"analytics_id"},
			expectedSummaries: []string{"Missing Attribute Configuration"},
		},
		{
			name:              "rybbit without a script URL",
			analyticsType:     types.StringValue("rybbit"),
			analyticsID:       types.StringValue("site-1"),
			analyticsScript:   types.StringNull(),
			expectedErrors:    []string{"analytics_script_url"},
			expectedSummaries: []string{"Missing Attribute Configuration"},
		},
		{
			name:            "rybbit companions only known after apply",
			analyticsType:   types.StringValue("rybbit"),
			analyticsID:     types.StringUnknown(),
			analyticsScript: types.StringUnknown(),
		},
		{
			name:              "rybbit with neither companion",
			analyticsType:     types.StringValue("rybbit"),
			analyticsID:       types.StringNull(),
			analyticsScript:   types.StringNull(),
			expectedErrors:    []string{"analytics_id", "analytics_script_url"},
			expectedSummaries: []string{"Missing Attribute Configuration", "Missing Attribute Configuration"},
		},
		{
			name:              "matomo without a host",
			analyticsType:     types.StringValue("matomo"),
			analyticsID:       types.StringValue("1"),
			analyticsScript:   types.StringNull(),
			expectedErrors:    []string{"analytics_script_url"},
			expectedSummaries: []string{"Missing Attribute Configuration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := StatusPageResourceModel{
				AnalyticsType:      tt.analyticsType,
				AnalyticsID:        tt.analyticsID,
				AnalyticsScriptURL: tt.analyticsScript,
			}

			resp := &resource.ValidateConfigResponse{}

			validateStatusPageAnalytics(&config, resp)

			gotPaths := make([]string, 0, len(resp.Diagnostics))
			gotSummaries := make([]string, 0, len(resp.Diagnostics))

			for _, d := range resp.Diagnostics {
				withPath, ok := d.(diag.DiagnosticWithPath)
				if !ok {
					t.Fatalf("diagnostic %q has no attribute path", d.Summary())
				}

				gotPaths = append(gotPaths, withPath.Path().String())
				gotSummaries = append(gotSummaries, d.Summary())
			}

			if len(gotPaths) != len(tt.expectedErrors) {
				t.Fatalf("got errors for %v, want errors for %v", gotPaths, tt.expectedErrors)
			}

			for i, want := range tt.expectedErrors {
				if gotPaths[i] != want {
					t.Errorf("error %d is for %q, want %q", i, gotPaths[i], want)
				}
			}

			for i, want := range tt.expectedSummaries {
				if gotSummaries[i] != want {
					t.Errorf("error %d has summary %q, want %q", i, gotSummaries[i], want)
				}
			}
		})
	}
}
