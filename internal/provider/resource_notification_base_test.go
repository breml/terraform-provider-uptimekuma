package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestApplyExistingFromState(t *testing.T) {
	tests := []struct {
		name  string
		state types.Bool
		want  types.Bool
	}{
		{name: "true kept", state: types.BoolValue(true), want: types.BoolValue(true)},
		{name: "false kept", state: types.BoolValue(false), want: types.BoolValue(false)},
		{name: "null after import", state: types.BoolNull(), want: types.BoolValue(false)},
		{name: "unknown", state: types.BoolUnknown(), want: types.BoolValue(false)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := applyExistingFromState(tc.state)
			if !got.Equal(tc.want) {
				t.Errorf("applyExistingFromState(%s) = %s, want %s", tc.state, got, tc.want)
			}
		})
	}
}

func TestApplyExistingOnUpdate(t *testing.T) {
	tests := []struct {
		name    string
		prior   bool
		planned bool
		want    bool
	}{
		{name: "false to true applies", prior: false, planned: true, want: true},
		{name: "true stays true does not apply again", prior: true, planned: true, want: false},
		{name: "true to false", prior: true, planned: false, want: false},
		{name: "false stays false", prior: false, planned: false, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := resource.UpdateRequest{
				Plan:  tfsdk.Plan(applyExistingTestObject(tc.planned)),
				State: applyExistingTestObject(tc.prior),
			}

			var diags diag.Diagnostics

			got := applyExistingOnUpdate(t.Context(), req, &diags)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if got != tc.want {
				t.Errorf("applyExistingOnUpdate(prior=%t, planned=%t) = %t, want %t",
					tc.prior, tc.planned, got, tc.want)
			}
		})
	}
}

// applyExistingTestObject returns plan or state data holding only the
// apply_existing attribute.
func applyExistingTestObject(value bool) tfsdk.State {
	objectType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"apply_existing": tftypes.Bool}}

	return tfsdk.State{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"apply_existing": schema.BoolAttribute{Optional: true},
			},
		},
		Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
			"apply_existing": tftypes.NewValue(tftypes.Bool, value),
		}),
	}
}
