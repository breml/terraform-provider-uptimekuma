package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestMonitorSFTPCredentialsAreSensitive guards against the SSH credentials leaking in plan
// and show output. The import tests cannot catch a regression: the server returns the
// credentials to the owning user, so ImportStateVerify passes whether or not they are
// sensitive.
func TestMonitorSFTPCredentialsAreSensitive(t *testing.T) {
	t.Parallel()

	resp := &resource.SchemaResponse{}
	(&MonitorSFTPResource{}).Schema(t.Context(), resource.SchemaRequest{}, resp)

	for _, name := range []string{"ssh_password", "ssh_private_key", "ssh_passphrase"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			attr, ok := resp.Schema.Attributes[name].(schema.StringAttribute)
			if !ok {
				t.Fatalf("%s is %T, want schema.StringAttribute", name, resp.Schema.Attributes[name])
			}

			if !attr.Sensitive {
				t.Errorf("%s is not marked Sensitive", name)
			}

			if !attr.Optional || attr.Computed || attr.Required {
				t.Errorf(
					"%s must be optional only, got Optional=%t Computed=%t Required=%t",
					name, attr.Optional, attr.Computed, attr.Required,
				)
			}
		})
	}
}
