package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

// findMonitorByName searches for a monitor by name and type.
// Returns nil if not found or if multiple matches exist.
func findMonitorByName(
	ctx context.Context,
	client *kuma.Client,
	name string,
	monitorType string,
	diags *diag.Diagnostics,
) monitor.Monitor {
	// Fetch all monitors from the API.
	monitors, err := client.GetMonitors(ctx)
	if err != nil {
		diags.AddError("failed to read monitors", err.Error())
		return nil
	}

	// Search for the monitor matching the given name and type.
	var found monitor.Monitor
	for i := range monitors {
		mon := &monitors[i]
		// Skip monitors that don't match the name or type.
		if mon.Name != name || mon.Type() != monitorType {
			continue
		}

		// Report error if multiple monitors match.
		if found != nil {
			diags.AddError(
				"Multiple monitors found",
				fmt.Sprintf(
					"Multiple %s monitors with name '%s' found. Please use 'id' to specify the monitor uniquely.",
					monitorType,
					name,
				),
			)
			return nil
		}

		found = mon
	}

	// Report error if no monitor matches.
	if found == nil {
		diags.AddError(
			fmt.Sprintf("%s monitor not found", monitorType),
			fmt.Sprintf("No %s monitor with name '%s' found.", monitorType, name),
		)
		return nil
	}

	return found
}

// validateMonitorDataSourceInput validates that either id or name is provided.
//
// Both may be set: the lookup then goes by id, and readByID rejects a name that
// contradicts the monitor the id resolved to, see dataSourceNameMatches.
func validateMonitorDataSourceInput(
	resp *datasource.ReadResponse,
	idValue types.Int64,
	nameValue types.String,
) bool {
	if !idValue.IsNull() && !idValue.IsUnknown() {
		return true
	}

	if !nameValue.IsNull() && !nameValue.IsUnknown() {
		return true
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
	return false
}

// monitorTypeMatches reports whether the monitor an id lookup resolved to is of
// the type the data source serves, and adds an error to diags when it is not.
//
// It guards GetMonitorAs, which unmarshals into whatever target it is given
// without checking the type: a monitor of another type would otherwise decode
// into plausible-looking values, empty where the fields differ and the other
// monitor's where they happen to share a name. An empty actual type is let
// through, because it means the server reported none rather than another one.
func monitorTypeMatches(diags *diag.Diagnostics, id int64, actual string, want string) bool {
	if actual == "" || actual == want {
		return true
	}

	diags.AddError(
		"Monitor type mismatch",
		fmt.Sprintf("Monitor ID %d has type %q, expected %q.", id, actual, want),
	)

	return false
}

// dataSourceNameMatches reports whether the name an id lookup resolved to
// agrees with the name the configuration states, and adds an error to diags
// when it does not.
//
// A lookup goes by id whenever id is set, so without this a configured name
// that contradicts it would be overwritten with the resolved one and the
// expectation it expressed dropped without comment. A null or unknown name
// states no expectation and always matches.
//
// kind names the entity in the diagnostic, capitalised, e.g. "Monitor".
func dataSourceNameMatches(
	diags *diag.Diagnostics,
	kind string,
	id int64,
	configured types.String,
	actual string,
) bool {
	if configured.IsNull() || configured.IsUnknown() || configured.ValueString() == actual {
		return true
	}

	diags.AddError(
		kind+" name mismatch",
		fmt.Sprintf(
			"%s ID %d is named %q, but the configuration expects %q. "+
				"Set only one of 'id' and 'name', or correct the one that is wrong.",
			kind, id, actual, configured.ValueString(),
		),
	)

	return false
}
