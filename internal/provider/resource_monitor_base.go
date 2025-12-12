package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/tag"
)

type MonitorTagModel struct {
	TagID types.Int64  `tfsdk:"tag_id"`
	Value types.String `tfsdk:"value"`
}

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
	Tags            types.List   `tfsdk:"tags"`
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

	attrs["tags"] = schema.ListNestedAttribute{
		MarkdownDescription: "List of tags assigned to this monitor",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"tag_id": schema.Int64Attribute{
					MarkdownDescription: "Tag ID",
					Required:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Optional value for this tag",
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString(""),
				},
			},
		},
	}

	return attrs
}

func handleMonitorTagsCreate(ctx context.Context, client *kuma.Client, monitorID int64, tags types.List, diags *diag.Diagnostics) {
	if tags.IsNull() || tags.IsUnknown() {
		return
	}

	var monitorTags []MonitorTagModel
	diags.Append(tags.ElementsAs(ctx, &monitorTags, false)...)
	if diags.HasError() {
		return
	}

	for _, monitorTag := range monitorTags {
		tagID := monitorTag.TagID.ValueInt64()
		value := monitorTag.Value.ValueString()

		_, err := client.AddMonitorTag(ctx, tagID, monitorID, value)
		if err != nil {
			diags.AddError(
				fmt.Sprintf("failed to add tag %d to monitor %d", tagID, monitorID),
				err.Error(),
			)
			return
		}
	}
}

func handleMonitorTagsRead(ctx context.Context, monitorTags []tag.MonitorTag, diags *diag.Diagnostics) types.List {
	if len(monitorTags) == 0 {
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"tag_id": types.Int64Type,
				"value":  types.StringType,
			},
		})
	}

	tagModels := make([]MonitorTagModel, len(monitorTags))
	for i, monitorTag := range monitorTags {
		tagModels[i] = MonitorTagModel{
			TagID: types.Int64Value(monitorTag.TagID),
			Value: types.StringValue(monitorTag.Value),
		}
	}

	tagsList, diagsLocal := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"tag_id": types.Int64Type,
			"value":  types.StringType,
		},
	}, tagModels)

	diags.Append(diagsLocal...)
	return tagsList
}

func handleMonitorTagsUpdate(ctx context.Context, client *kuma.Client, monitorID int64, oldTags types.List, newTags types.List, diags *diag.Diagnostics) {
	var oldMonitorTags []MonitorTagModel
	var newMonitorTags []MonitorTagModel

	if !oldTags.IsNull() && !oldTags.IsUnknown() {
		diags.Append(oldTags.ElementsAs(ctx, &oldMonitorTags, false)...)
		if diags.HasError() {
			return
		}
	}

	if !newTags.IsNull() && !newTags.IsUnknown() {
		diags.Append(newTags.ElementsAs(ctx, &newMonitorTags, false)...)
		if diags.HasError() {
			return
		}
	}

	oldTagMap := make(map[string]MonitorTagModel)
	for _, tag := range oldMonitorTags {
		key := fmt.Sprintf("%d:%s", tag.TagID.ValueInt64(), tag.Value.ValueString())
		oldTagMap[key] = tag
	}

	newTagMap := make(map[string]MonitorTagModel)
	for _, tag := range newMonitorTags {
		key := fmt.Sprintf("%d:%s", tag.TagID.ValueInt64(), tag.Value.ValueString())
		newTagMap[key] = tag
	}

	for key, oldTag := range oldTagMap {
		if _, exists := newTagMap[key]; !exists {
			err := client.DeleteMonitorTagWithValue(ctx, oldTag.TagID.ValueInt64(), monitorID, oldTag.Value.ValueString())
			if err != nil {
				diags.AddError(
					fmt.Sprintf("failed to remove tag %d from monitor %d", oldTag.TagID.ValueInt64(), monitorID),
					err.Error(),
				)
				return
			}
		}
	}

	for key, newTag := range newTagMap {
		if _, exists := oldTagMap[key]; !exists {
			_, err := client.AddMonitorTag(ctx, newTag.TagID.ValueInt64(), monitorID, newTag.Value.ValueString())
			if err != nil {
				diags.AddError(
					fmt.Sprintf("failed to add tag %d to monitor %d", newTag.TagID.ValueInt64(), monitorID),
					err.Error(),
				)
				return
			}
		}
	}
}
