package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

func findNotificationByName(
	ctx context.Context,
	client *kuma.Client,
	name string,
	notificationType string,
	diags *diag.Diagnostics,
) (int64, bool) {
	notifications := client.GetNotifications(ctx)

	var found int64
	var foundCount int

	for i := range notifications {
		if notifications[i].Name == name && notifications[i].Type() == notificationType {
			// Error if multiple matches found.
			if foundCount > 0 {
				diags.AddError(
					"Multiple notifications found",
					fmt.Sprintf(
						"Multiple %s notifications with name '%s' found. Please use 'id' to specify the notification uniquely.",
						notificationType,
						name,
					),
				)
				return 0, false
			}

			found = notifications[i].GetID()
			foundCount++
		}
	}

	// Error if no matching item found.
	if foundCount == 0 {
		diags.AddError(
			"Notification not found",
			fmt.Sprintf("No %s notification with name '%s' found.", notificationType, name),
		)
		return 0, false
	}

	return found, true
}

func validateNotificationDataSourceInput(
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
