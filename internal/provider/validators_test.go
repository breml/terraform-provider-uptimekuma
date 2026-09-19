package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWholeNumberValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     types.Float64
		wantError bool
	}{
		"whole":             {value: types.Float64Value(48), wantError: false},
		"whole negative":    {value: types.Float64Value(-3), wantError: false},
		"zero":              {value: types.Float64Value(0), wantError: false},
		"fractional":        {value: types.Float64Value(2.5), wantError: true},
		"barely fractional": {value: types.Float64Value(2.0000001), wantError: true},
		"null":              {value: types.Float64Null(), wantError: false},
		"unknown":           {value: types.Float64Unknown(), wantError: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.Float64Response{}
			wholeNumber().ValidateFloat64(
				t.Context(),
				validator.Float64Request{ConfigValue: test.value},
				resp,
			)

			if got := resp.Diagnostics.HasError(); got != test.wantError {
				t.Errorf("HasError() = %v, want %v (%v)", got, test.wantError, resp.Diagnostics)
			}
		})
	}
}

func TestFloat64ToPtr(t *testing.T) {
	t.Parallel()

	if got := float64ToPtr(types.Float64Null()); got != nil {
		t.Errorf("float64ToPtr(null) = %v, want nil", *got)
	}

	if got := float64ToPtr(types.Float64Unknown()); got != nil {
		t.Errorf("float64ToPtr(unknown) = %v, want nil", *got)
	}

	got := float64ToPtr(types.Float64Value(2.5))
	if got == nil {
		t.Fatal("float64ToPtr(2.5) = nil, want pointer to 2.5")
	}

	if *got != 2.5 {
		t.Errorf("float64ToPtr(2.5) = %v, want 2.5", *got)
	}
}

// systemServiceNameValidators returns the validators the System Service
// monitor resource declares for system_service_name.
func systemServiceNameValidators(t *testing.T) []validator.String {
	t.Helper()

	resp := &resource.SchemaResponse{}
	(&MonitorSystemServiceResource{}).Schema(t.Context(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["system_service_name"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("system_service_name is %T, want schema.StringAttribute", resp.Schema.Attributes["system_service_name"])
	}

	return attr.Validators
}

func TestMonitorSystemServiceNameValidation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value     string
		wantError bool
	}{
		"systemd unit":      {value: "nginx.service", wantError: false},
		"templated unit":    {value: "sshd@0.service", wantError: false},
		"hyphen":            {value: "user-session.service", wantError: false},
		"underscore":        {value: "my_service", wantError: false},
		"windows scm name":  {value: "Spooler", wantError: false},
		"digits":            {value: "svc42", wantError: false},
		"empty":             {value: "", wantError: true},
		"space":             {value: "my service", wantError: true},
		"leading space":     {value: " nginx.service", wantError: true},
		"slash":             {value: "nginx/1", wantError: true},
		"trailing newline":  {value: "nginx.service\n", wantError: true},
		"shell metacharact": {value: "nginx;reboot", wantError: true},
	}

	validators := systemServiceNameValidators(t)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := &validator.StringResponse{}
			for _, v := range validators {
				v.ValidateString(
					t.Context(),
					validator.StringRequest{ConfigValue: types.StringValue(test.value)},
					resp,
				)
			}

			if got := resp.Diagnostics.HasError(); got != test.wantError {
				t.Errorf("HasError() = %v, want %v (%v)", got, test.wantError, resp.Diagnostics)
			}
		})
	}
}
