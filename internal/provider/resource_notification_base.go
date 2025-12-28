package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
