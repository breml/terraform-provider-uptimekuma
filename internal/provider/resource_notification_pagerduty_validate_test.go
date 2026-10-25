package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestValidatePagerDutySecrets covers the check that the integration URL and key are set, either
// as the stored attribute or as its write-only variant.
func TestValidatePagerDutySecrets(t *testing.T) {
	tests := []struct {
		name             string
		integrationURL   types.String
		integrationURLWO types.String
		integrationKey   types.String
		integrationKeyWO types.String
		expectedErrors   []string
	}{
		{
			name:           "stored attributes set",
			integrationURL: types.StringValue("https://events.pagerduty.com/v2/enqueue"),
			integrationKey: types.StringValue("key"),
		},
		{
			name:             "write-only attributes set",
			integrationURLWO: types.StringValue("https://events.pagerduty.com/v2/enqueue"),
			integrationKeyWO: types.StringValue("key"),
		},
		{
			name:           "values unknown",
			integrationURL: types.StringUnknown(),
			integrationKey: types.StringUnknown(),
		},
		{
			name:           "nothing set",
			expectedErrors: []string{"integration_url", "integration_key"},
		},
		{
			name:             "key missing",
			integrationURLWO: types.StringValue("https://events.pagerduty.com/v2/enqueue"),
			expectedErrors:   []string{"integration_key"},
		},
		{
			name:             "write-only key empty",
			integrationURL:   types.StringValue("https://events.pagerduty.com/v2/enqueue"),
			integrationKeyWO: types.StringValue(""),
			expectedErrors:   []string{"integration_key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NotificationPagerDutyResourceModel{
				IntegrationURL:   tt.integrationURL,
				IntegrationURLWO: tt.integrationURLWO,
				IntegrationKey:   tt.integrationKey,
				IntegrationKeyWO: tt.integrationKeyWO,
			}

			resp := &resource.ValidateConfigResponse{}

			validatePagerDutySecrets(&config, resp)

			assertDiagnosticPaths(t, resp.Diagnostics, tt.expectedErrors)
		})
	}
}
