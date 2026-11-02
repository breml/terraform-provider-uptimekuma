package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestValidateAmootSMSSendMode covers the values each send mode requires, including the omitted
// use_pattern, which is null during validation and means a plain SMS.
func TestValidateAmootSMSSendMode(t *testing.T) {
	line := types.StringValue("public")
	code := types.Int64Value(1234)

	tests := []struct {
		name           string
		usePattern     types.Bool
		patternCodeID  types.Int64
		useOwnLine     types.Bool
		lineNumber     types.String
		expectedErrors []string
	}{
		{
			name:       "plain, use_pattern omitted, line number set",
			usePattern: types.BoolNull(),
			lineNumber: line,
		},
		{
			name:       "plain, use_pattern false, line number set",
			usePattern: types.BoolValue(false),
			lineNumber: line,
		},
		{
			name:           "plain, use_pattern omitted, line number missing",
			usePattern:     types.BoolNull(),
			lineNumber:     types.StringNull(),
			expectedErrors: []string{"line_number"},
		},
		{
			name:           "plain, use_pattern false, line number missing",
			usePattern:     types.BoolValue(false),
			lineNumber:     types.StringNull(),
			expectedErrors: []string{"line_number"},
		},
		{
			name:       "plain, line number only known after apply",
			usePattern: types.BoolNull(),
			lineNumber: types.StringUnknown(),
		},
		{
			name:           "plain, use_own_line unknown does not lift the line number requirement",
			usePattern:     types.BoolValue(false),
			useOwnLine:     types.BoolUnknown(),
			lineNumber:     types.StringNull(),
			expectedErrors: []string{"line_number"},
		},
		{
			name:       "plain, pattern fields are ignored by the server and allowed",
			usePattern: types.BoolValue(false),
			useOwnLine: types.BoolValue(true),
			lineNumber: line,
		},
		{
			name:          "pattern, code set, no line number",
			usePattern:    types.BoolValue(true),
			patternCodeID: code,
			lineNumber:    types.StringNull(),
		},
		{
			name:          "pattern, use_own_line false, no line number",
			usePattern:    types.BoolValue(true),
			patternCodeID: code,
			useOwnLine:    types.BoolValue(false),
			lineNumber:    types.StringNull(),
		},
		{
			name:          "pattern, line number is ignored by the server and allowed",
			usePattern:    types.BoolValue(true),
			patternCodeID: code,
			lineNumber:    line,
		},
		{
			name:           "pattern, code missing",
			usePattern:     types.BoolValue(true),
			patternCodeID:  types.Int64Null(),
			lineNumber:     types.StringNull(),
			expectedErrors: []string{"pattern_code_id"},
		},
		{
			name:          "pattern, code only known after apply",
			usePattern:    types.BoolValue(true),
			patternCodeID: types.Int64Unknown(),
			lineNumber:    types.StringNull(),
		},
		{
			name:          "pattern and own line, all set",
			usePattern:    types.BoolValue(true),
			patternCodeID: code,
			useOwnLine:    types.BoolValue(true),
			lineNumber:    line,
		},
		{
			name:           "pattern and own line, line number missing",
			usePattern:     types.BoolValue(true),
			patternCodeID:  code,
			useOwnLine:     types.BoolValue(true),
			lineNumber:     types.StringNull(),
			expectedErrors: []string{"line_number"},
		},
		{
			name:           "pattern and own line, both missing",
			usePattern:     types.BoolValue(true),
			patternCodeID:  types.Int64Null(),
			useOwnLine:     types.BoolValue(true),
			lineNumber:     types.StringNull(),
			expectedErrors: []string{"pattern_code_id", "line_number"},
		},
		{
			name:          "pattern, use_own_line unknown, no line number",
			usePattern:    types.BoolValue(true),
			patternCodeID: code,
			useOwnLine:    types.BoolUnknown(),
			lineNumber:    types.StringNull(),
		},
		{
			name:           "pattern, use_own_line unknown, code still required",
			usePattern:     types.BoolValue(true),
			patternCodeID:  types.Int64Null(),
			useOwnLine:     types.BoolUnknown(),
			lineNumber:     types.StringNull(),
			expectedErrors: []string{"pattern_code_id"},
		},
		{
			name:       "use_pattern unknown, nothing to validate against",
			usePattern: types.BoolUnknown(),
			useOwnLine: types.BoolValue(true),
			lineNumber: types.StringNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NotificationAmootSMSResourceModel{
				UsePattern:    tt.usePattern,
				PatternCodeID: tt.patternCodeID,
				UseOwnLine:    tt.useOwnLine,
				LineNumber:    tt.lineNumber,
			}

			resp := &resource.ValidateConfigResponse{}

			validateAmootSMSSendMode(&config, resp)

			gotPaths := make([]string, 0, len(resp.Diagnostics))

			for _, d := range resp.Diagnostics {
				withPath, ok := d.(diag.DiagnosticWithPath)
				if !ok {
					t.Fatalf("diagnostic %q has no attribute path", d.Summary())
				}

				gotPaths = append(gotPaths, withPath.Path().String())
			}

			if len(gotPaths) != len(tt.expectedErrors) {
				t.Fatalf("got errors for %v, want errors for %v", gotPaths, tt.expectedErrors)
			}

			for i, want := range tt.expectedErrors {
				if gotPaths[i] != want {
					t.Errorf("error %d is for %q, want %q", i, gotPaths[i], want)
				}
			}
		})
	}
}
