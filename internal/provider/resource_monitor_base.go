package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MonitorBaseModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Parent          types.Int64  `tfsdk:"parent"`
	Interval        types.Int64  `tfsdk:"interval"`
	RetryInterval   types.Int64  `tfsdk:"retry_interval"`
	ResendInterval  types.Int64  `tfsdk:"resend_interval"`
	MaxRetries      types.Int64  `tfsdk:"max_retries"`
	UpsideDown      types.Bool   `tfsdk:"upside_down"`
	Active          types.Bool   `tfsdk:"active"`
	NotificationIDs types.List   `tfsdk:"notification_ids"`
}

func withMonitorBaseAttributes(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	attrs["id"] = schema.Int64Attribute{
		Computed:            true,
		MarkdownDescription: "Monitor identifier",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	}

	attrs["name"] = schema.StringAttribute{
		MarkdownDescription: "Friendly name",
		Required:            true,
	}

	attrs["description"] = schema.StringAttribute{
		MarkdownDescription: "Description",
		Optional:            true,
	}

	attrs["parent"] = schema.Int64Attribute{
		MarkdownDescription: "Parent monitor ID for hierarchical organization",
		Optional:            true,
	}

	attrs["interval"] = schema.Int64Attribute{
		MarkdownDescription: "Heartbeat interval in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(60),
		Validators: []validator.Int64{
			int64validator.Between(20, 2073600),
		},
	}

	attrs["retry_interval"] = schema.Int64Attribute{
		MarkdownDescription: "Retry interval in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(60),
		Validators: []validator.Int64{
			int64validator.Between(20, 2073600),
		},
	}

	attrs["resend_interval"] = schema.Int64Attribute{
		MarkdownDescription: "Resend interval in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(0),
	}

	attrs["max_retries"] = schema.Int64Attribute{
		MarkdownDescription: "Maximum number of retries",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(3),
		Validators: []validator.Int64{
			int64validator.Between(0, 10),
		},
	}

	attrs["upside_down"] = schema.BoolAttribute{
		MarkdownDescription: "Invert monitor status (treat DOWN as UP and vice versa)",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	}

	attrs["active"] = schema.BoolAttribute{
		MarkdownDescription: "Monitor is active",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(true),
	}

	attrs["notification_ids"] = schema.ListAttribute{
		MarkdownDescription: "List of notification IDs",
		ElementType:         types.Int64Type,
		Optional:            true,
	}

	return attrs
}
