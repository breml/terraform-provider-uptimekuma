package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// providerData holds the configured client and credentials passed from the
// provider to each resource and data source via Configure.
type providerData struct {
	client   *kuma.Client
	password string
	// apiKey reports whether the provider authenticates with `api_key`, which
	// rules out a password, so that uptimekuma_settings can name the reason
	// it has none.
	apiKey bool
}

// configureClient extracts the Uptime Kuma client from provider data.
// Returns nil when provider data is nil (early call before Configure).
func configureClient(pd any, diags *diag.Diagnostics) *kuma.Client {
	if pd == nil {
		return nil
	}

	data, ok := pd.(*providerData)
	if !ok {
		diags.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf(
				"Expected *providerData, got: %T. Please report this issue to the provider developers.",
				pd,
			),
		)

		return nil
	}

	return data.client
}
