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

func findMonitorByName(
	ctx context.Context,
	client *kuma.Client,
	name string,
	monitorType string,
	diags *diag.Diagnostics,
) monitor.Monitor {
	monitors, err := client.GetMonitors(ctx)
	if err != nil {
		diags.AddError("failed to read monitors", err.Error())
		return nil
	}

	var found monitor.Monitor
	for _, mon := range monitors {
		if mon.Name != name || mon.Type() != monitorType {
			continue
		}

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

	if found == nil {
		diags.AddError(
			fmt.Sprintf("%s monitor not found", monitorType),
			fmt.Sprintf("No %s monitor with name '%s' found.", monitorType, name),
		)
		return nil
	}

	return found
}

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
