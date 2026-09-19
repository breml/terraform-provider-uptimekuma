package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

// findNotificationByName searches for a notification by name and type.
//
// The notification getters serve from the client's local state cache, so a
// notification the server already has can be missing from it for as long as the
// broadcast that carries it is outstanding. findWithResync rebuilds the cache
// once before the miss is reported, see its comment for why the broadcast can
// be late.
func findNotificationByName(
	ctx context.Context,
	client *kuma.Client,
	name string,
	notificationType string,
	diags *diag.Diagnostics,
) (int64, bool) {
	matches, found := findWithResync(ctx, client, func(ctx context.Context) ([]int64, bool) {
		ids := matchNotificationsByName(ctx, client, name, notificationType)

		return ids, len(ids) > 0
	}, diags)
	if diags.HasError() {
		return 0, false
	}

	// Error if no matching item found.
	if !found {
		diags.AddError(
			"Notification not found",
			fmt.Sprintf("No %s notification with name '%s' found.", notificationType, name),
		)

		return 0, false
	}

	// Error if multiple matches found.
	if len(matches) > 1 {
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

	return matches[0], true
}

// matchNotificationsByName returns the IDs of the notifications of the given
// type that carry name. It reports every match rather than the first, so that
// the caller can tell an ambiguous name from a missing one and only the latter
// is worth a resync.
func matchNotificationsByName(
	ctx context.Context,
	client *kuma.Client,
	name string,
	notificationType string,
) []int64 {
	notifications := client.GetNotifications(ctx)

	var ids []int64

	for i := range notifications {
		if notifications[i].Name == name && notifications[i].Type() == notificationType {
			ids = append(ids, notifications[i].GetID())
		}
	}

	return ids
}

// readNotificationWithResync reads a notification by ID for a data source.
//
// Like findNotificationByName it resyncs once before it reports a miss, because
// the getter serves from the state cache. The second return value reports
// whether the read succeeded; it is false only with an error in diags, because
// a data source that cannot produce its resource has nothing else to say.
func readNotificationWithResync(
	ctx context.Context,
	client *kuma.Client,
	id int64,
	diags *diag.Diagnostics,
) (notification.Base, bool) {
	errorsBefore := diags.ErrorsCount()

	notif, found := readWithResync(ctx, client, id, "failed to read notification", client.GetNotification, diags)
	if diags.ErrorsCount() > errorsBefore {
		return notif, false
	}

	if !found {
		diags.AddError(
			"Notification not found",
			fmt.Sprintf("No notification with ID %d found.", id),
		)

		return notif, false
	}

	return notif, true
}

// readNotificationByID reads a notification of the given type by ID for a data
// source.
//
// wantDescription names the expected type the way the error reads it, e.g.
// "a 46elks notification".
func readNotificationByID(
	ctx context.Context,
	client *kuma.Client,
	id int64,
	notificationType string,
	wantDescription string,
	diags *diag.Diagnostics,
) (notification.Base, bool) {
	notif, found := readNotificationWithResync(ctx, client, id, diags)
	if !found {
		return notif, false
	}

	if notif.Type() != notificationType {
		diags.AddError("Incorrect notification type", "Notification is not "+wantDescription)

		return notif, false
	}

	return notif, true
}

// validateNotificationDataSourceInput validates that either id or name is provided.
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
