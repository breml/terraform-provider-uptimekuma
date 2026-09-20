package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestValidatePlivoAnswerURL covers the answer_url validation for every message type, including
// the omitted message_type, which is null during validation and means the server default (SMS).
func TestValidatePlivoAnswerURL(t *testing.T) {
	tests := []struct {
		name              string
		messageType       types.String
		answerURL         types.String
		expectedErrors    []string
		expectedSummaries []string
	}{
		{
			name:        "message type omitted, no answer URL",
			messageType: types.StringNull(),
			answerURL:   types.StringNull(),
		},
		{
			name:              "message type omitted, answer URL set",
			messageType:       types.StringNull(),
			answerURL:         types.StringValue("https://example.com/answer"),
			expectedErrors:    []string{"answer_url"},
			expectedSummaries: []string{"Invalid Attribute Combination"},
		},
		{
			name:        "message type unknown, answer URL set",
			messageType: types.StringUnknown(),
			answerURL:   types.StringValue("https://example.com/answer"),
		},
		{
			name:        "message type unknown, no answer URL",
			messageType: types.StringUnknown(),
			answerURL:   types.StringNull(),
		},
		{
			name:        "sms, no answer URL",
			messageType: types.StringValue("sms"),
			answerURL:   types.StringNull(),
		},
		{
			name:              "sms, answer URL set",
			messageType:       types.StringValue("sms"),
			answerURL:         types.StringValue("https://example.com/answer"),
			expectedErrors:    []string{"answer_url"},
			expectedSummaries: []string{"Invalid Attribute Combination"},
		},
		{
			name:        "call, answer URL set",
			messageType: types.StringValue("call"),
			answerURL:   types.StringValue("https://example.com/answer"),
		},
		{
			name:        "call, answer URL only known after apply",
			messageType: types.StringValue("call"),
			answerURL:   types.StringUnknown(),
		},
		{
			name:              "call, answer URL missing",
			messageType:       types.StringValue("call"),
			answerURL:         types.StringNull(),
			expectedErrors:    []string{"answer_url"},
			expectedSummaries: []string{"Missing Attribute Configuration"},
		},
		{
			name:        "unexpected message type is left to the schema validator",
			messageType: types.StringValue("fax"),
			answerURL:   types.StringNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NotificationPlivoResourceModel{
				MessageType: tt.messageType,
				AnswerURL:   tt.answerURL,
			}

			resp := &resource.ValidateConfigResponse{}

			validatePlivoAnswerURL(&config, resp)

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

			// Both failure modes report on answer_url, so only the summary tells them apart.
			if len(gotSummaries) != len(tt.expectedSummaries) {
				t.Fatalf("got summaries %v, want %v", gotSummaries, tt.expectedSummaries)
			}

			for i, want := range tt.expectedSummaries {
				if gotSummaries[i] != want {
					t.Errorf("error %d has summary %q, want %q", i, gotSummaries[i], want)
				}
			}
		})
	}
}
