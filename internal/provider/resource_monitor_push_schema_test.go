package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestMonitorPushTokenIsSensitive guards against push_token leaking in plan and show
// output. Anyone holding the token can send heartbeats for the monitor, so it is a
// credential (see issue #453).
func TestMonitorPushTokenIsSensitive(t *testing.T) {
	t.Parallel()

	resp := &resource.SchemaResponse{}
	(&MonitorPushResource{}).Schema(t.Context(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["push_token"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("push_token is %T, want schema.StringAttribute", resp.Schema.Attributes["push_token"])
	}

	if !attr.Sensitive {
		t.Error("push_token is not marked Sensitive")
	}

	if !attr.Computed || attr.Optional || attr.Required {
		t.Errorf(
			"push_token must be computed only, got Computed=%t Optional=%t Required=%t",
			attr.Computed, attr.Optional, attr.Required,
		)
	}
}
