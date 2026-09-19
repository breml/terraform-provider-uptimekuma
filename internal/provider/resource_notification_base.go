package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

// NotificationBaseModel describes the base data model for all notification types.
type NotificationBaseModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	IsActive      types.Bool   `tfsdk:"is_active"`
	IsDefault     types.Bool   `tfsdk:"is_default"`
	ApplyExisting types.Bool   `tfsdk:"apply_existing"`
}

// withNotificationBaseAttributes adds common notification schema attributes to the provided attribute map.
// These attributes are shared across all notification types: id, name, is_active, is_default, apply_existing.
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

	// Activation status for the notification.
	attrs["is_active"] = schema.BoolAttribute{
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(true),
	}

	// Default notification flag.
	attrs["is_default"] = schema.BoolAttribute{
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(false),
	}

	// Apply notification to existing monitors.
	attrs["apply_existing"] = schema.BoolAttribute{
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(false),
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
	base, found, err := readWithResync(ctx, client, id, client.GetNotification, diags)
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
