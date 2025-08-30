package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NotificationBaseModel struct {
	Id            types.Int32  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	IsActive      types.Bool   `tfsdk:"is_active"`
	IsDefault     types.Bool   `tfsdk:"is_default"`
	ApplyExisting types.Bool   `tfsdk:"apply_existing"`
}

func withNotificationBaseAttributes(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	attrs["id"] = schema.Int32Attribute{
		Computed:            true,
		MarkdownDescription: "Notification identifier",
		PlanModifiers: []planmodifier.Int32{
			int32planmodifier.UseStateForUnknown(),
		},
	}

	attrs["name"] = schema.StringAttribute{
		MarkdownDescription: "Notification name",
		Required:            true,
	}

	attrs["is_active"] = schema.BoolAttribute{
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(true),
	}

	attrs["is_default"] = schema.BoolAttribute{
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(false),
	}

	attrs["apply_existing"] = schema.BoolAttribute{
		Optional: true,
		Computed: true,
		Default:  booldefault.StaticBool(false),
	}

	return attrs
}
