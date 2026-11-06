package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

// NotificationBaseModel describes the base data model for all notification types.
//
// There is deliberately no is_active attribute: Uptime Kuma does not store a
// notification's active value and does not check it when sending alerts. Create
// and Update always send IsActive: true, so notifications stay enabled if a
// server version starts honoring the field. See #478.
type NotificationBaseModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	IsDefault     types.Bool   `tfsdk:"is_default"`
	ApplyExisting types.Bool   `tfsdk:"apply_existing"`
}

// applyExistingDescription documents the apply_existing attribute shared by all
// notification resources.
const applyExistingDescription = "Attach this notification to all existing monitors. This is a one-time " +
	"action: it runs when the notification is created with `true`, or when the value changes from `false` " +
	"to `true`. Other updates do not attach it again, so monitors you detach from the notification afterwards " +
	"stay detached. Uptime Kuma does not store the value, so the provider does not refresh it from the " +
	"server; the value in state is the one last applied. An imported notification starts with `false`, so " +
	"the first apply of a configuration with `true` after an import attaches it to all monitors."

// withNotificationBaseAttributes adds common notification schema attributes to the provided attribute map.
// These attributes are shared across all notification types: id, name, is_default, apply_existing.
func withNotificationBaseAttributes(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	// Notification identifier (computed).
	attrs["id"] = schema.Int64Attribute{
		Computed:            true,
		MarkdownDescription: "Notification identifier",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	}

	// Human-readable notification name.
	attrs["name"] = schema.StringAttribute{
		MarkdownDescription: "Notification name",
		Required:            true,
	}

	// Default notification flag.
	attrs["is_default"] = schema.BoolAttribute{
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(false),
	}

	// Apply notification to existing monitors, once; see applyExistingDescription.
	attrs["apply_existing"] = schema.BoolAttribute{
		MarkdownDescription: applyExistingDescription,
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	}

	// Return enriched attributes map.
	return attrs
}

// readNotification fetches a notification for a resource read and verifies its
// type.
//
// Client.GetNotification serves from the client's local state cache, which a
// missed update event leaves stale: a notification that is alive on the server
// is reported as not found until the next resync. Treating that as a deletion
// would drop the resource from state and make the next apply create a
// duplicate, so a resync is forced once before the miss is believed.
//
// The type check guards notification.Base.As, which unmarshals whatever it is
// given. Without it, reading a notification of a different type into this
// resource would fill state with zero values and the next apply would push
// those values to the server.
//
// found reports whether the notification exists. It is also false when the read
// failed, so the caller must check diags for errors before treating a false as
// a deletion - unlike readNotificationWithResync, whose false is never silent.
func readNotification(
	ctx context.Context,
	client *kuma.Client,
	id int64,
	wantType string,
	diags *diag.Diagnostics,
) (notification.Base, bool) {
	base, found, err := readWithResync(ctx, client, id, client.GetNotification)
	if err != nil {
		diags.AddError("failed to read notification", err.Error())

		return base, false
	}

	if !found {
		return base, false
	}

	if base.Type() != wantType {
		diags.AddError(
			"Incorrect notification type",
			fmt.Sprintf(
				"Notification with ID %d has type %q, expected %q. It is managed by a different "+
					"notification resource type.",
				id, base.Type(), wantType,
			),
		)

		return base, false
	}

	return base, true
}

// applyExistingFromState returns the apply_existing value for a notification
// Read.
//
// Uptime Kuma 2.x treats applyExisting as a one-time action: Notification.save
// in server/notification.js clears it before storing the config, so it always
// reads back false (1.23 persisted whatever was sent). The server value would
// therefore replace a configured true with false on every refresh. Read keeps
// the prior state value instead, and falls back to false when there is none,
// e.g. right after an import. See #478.
func applyExistingFromState(state types.Bool) types.Bool {
	if state.IsNull() || state.IsUnknown() {
		return types.BoolValue(false)
	}

	return state
}

// applyExistingOnUpdate returns the applyExisting value to send on a
// notification Update.
//
// Uptime Kuma attaches the notification to every monitor each time it is saved
// with applyExisting true. Sending the planned value on every Update would
// therefore attach it again on any unrelated change, undoing monitors detached
// on purpose. The flag is only sent as true when the plan changes it from false
// to true.
//
// apply_existing is Optional and Computed with a static default, so the planned
// value is always known here. An unknown would read as false and skip the
// attach.
func applyExistingOnUpdate(ctx context.Context, req resource.UpdateRequest, diags *diag.Diagnostics) bool {
	var planned, prior types.Bool

	diags.Append(req.Plan.GetAttribute(ctx, path.Root("apply_existing"), &planned)...)
	diags.Append(req.State.GetAttribute(ctx, path.Root("apply_existing"), &prior)...)

	return planned.ValueBool() && !prior.ValueBool()
}
