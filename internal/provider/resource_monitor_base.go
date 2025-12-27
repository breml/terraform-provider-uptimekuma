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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/tag"
)

// MonitorTagModel describes the tag data model for monitors.
type MonitorTagModel struct {
	TagID types.Int64  `tfsdk:"tag_id"`
	Value types.String `tfsdk:"value"`
}

// MonitorBaseModel describes the base data model for all monitor types.
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

// withMonitorBaseAttributes adds common monitor schema attributes to the provided attribute map.
// These attributes are shared across all monitor types: id, name, description, parent, interval, retry, etc.
func withMonitorBaseAttributes(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	// Monitor identifier (computed).
	attrs["id"] = schema.Int64Attribute{
		Computed:            true,
		MarkdownDescription: "Monitor identifier",
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	}

	// Human-readable monitor name.
	attrs["name"] = schema.StringAttribute{
		MarkdownDescription: "Friendly name",
		Required:            true,
	}

	// Optional monitor description.
	attrs["description"] = schema.StringAttribute{
		MarkdownDescription: "Description",
		Optional:            true,
	}

	// Parent monitor ID for hierarchical relationships.
	attrs["parent"] = schema.Int64Attribute{
		MarkdownDescription: "Parent monitor ID for hierarchical organization",
		Optional:            true,
	}

	// Heartbeat interval between checks.
	attrs["interval"] = schema.Int64Attribute{
		MarkdownDescription: "Heartbeat interval in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(60),
		Validators: []validator.Int64{
			int64validator.Between(20, 2073600),
		},
	}

	// Retry interval after failure.
	attrs["retry_interval"] = schema.Int64Attribute{
		MarkdownDescription: "Retry interval in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(60),
		Validators: []validator.Int64{
			int64validator.Between(20, 2073600),
		},
	}

	// Resend interval for notifications.
	attrs["resend_interval"] = schema.Int64Attribute{
		MarkdownDescription: "Resend interval in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(0),
	}

	// Maximum retry attempts.
	attrs["max_retries"] = schema.Int64Attribute{
		MarkdownDescription: "Maximum number of retries",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(3),
		Validators: []validator.Int64{
			int64validator.Between(0, 10),
		},
	}

	// Invert monitor status logic.
	attrs["upside_down"] = schema.BoolAttribute{
		MarkdownDescription: "Invert monitor status (treat DOWN as UP and vice versa)",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	}

	// Monitor activation status.
	attrs["active"] = schema.BoolAttribute{
		MarkdownDescription: "Monitor is active",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(true),
	}

	// Associated notification IDs.
	attrs["notification_ids"] = schema.ListAttribute{
		MarkdownDescription: "List of notification IDs",
		ElementType:         types.Int64Type,
		Optional:            true,
	}

	// Monitor tags with optional values.
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
				},
			},
		},
	}

	// Return enriched attributes map.
	return attrs
}

func handleMonitorTagsCreate(
	ctx context.Context,
	client *kuma.Client,
	monitorID int64,
	tags types.List,
	diags *diag.Diagnostics,
) {
	if tags.IsNull() || tags.IsUnknown() {
		return
	}

	var monitorTags []MonitorTagModel
	diags.Append(tags.ElementsAs(ctx, &monitorTags, false)...)
	if diags.HasError() {
		return
	}

	// Iterate over tags and add each to the monitor.
	for _, monitorTag := range monitorTags {
		tagID := monitorTag.TagID.ValueInt64()
		value := ""
		if !monitorTag.Value.IsNull() {
			value = monitorTag.Value.ValueString()
		}

		// Call API to add tag to monitor.
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

	// Convert API tag models to Terraform models.
	tagModels := make([]MonitorTagModel, len(monitorTags))
	for i, monitorTag := range monitorTags {
		var value types.String
		if monitorTag.Value == "" {
			value = types.StringNull()
		} else {
			value = types.StringValue(monitorTag.Value)
		}

		tagModels[i] = MonitorTagModel{
			TagID: types.Int64Value(monitorTag.TagID),
			Value: value,
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

func handleMonitorTagsUpdate(
	ctx context.Context,
	client *kuma.Client,
	monitorID int64,
	oldTags types.List,
	newTags types.List,
	diags *diag.Diagnostics,
) {
	oldMonitorTags := deserializeMonitorTags(ctx, oldTags, diags)
	if diags.HasError() {
		return
	}

	newMonitorTags := deserializeMonitorTags(ctx, newTags, diags)
	if diags.HasError() {
		return
	}

	oldTagMap := buildMonitorTagMap(oldMonitorTags)
	newTagMap := buildMonitorTagMap(newMonitorTags)

	handleDeletedMonitorTags(ctx, client, monitorID, oldTagMap, newTagMap, diags)
	if diags.HasError() {
		return
	}

	handleAddedMonitorTags(ctx, client, monitorID, oldTagMap, newTagMap, diags)
}

func deserializeMonitorTags(ctx context.Context, tags types.List, diags *diag.Diagnostics) []MonitorTagModel {
	if tags.IsNull() || tags.IsUnknown() {
		return []MonitorTagModel{}
	}

	var monitorTags []MonitorTagModel
	diags.Append(tags.ElementsAs(ctx, &monitorTags, false)...)
	return monitorTags
}

func buildMonitorTagMap(tags []MonitorTagModel) map[string]MonitorTagModel {
	tagMap := make(map[string]MonitorTagModel)
	for _, tag := range tags {
		value := ""
		if !tag.Value.IsNull() {
			value = tag.Value.ValueString()
		}

		key := fmt.Sprintf("%d:%s", tag.TagID.ValueInt64(), value)
		tagMap[key] = tag
	}

	return tagMap
}

func handleDeletedMonitorTags(
	ctx context.Context,
	client *kuma.Client,
	monitorID int64,
	oldTagMap map[string]MonitorTagModel,
	newTagMap map[string]MonitorTagModel,
	diags *diag.Diagnostics,
) {
	for key, oldTag := range oldTagMap {
		if _, exists := newTagMap[key]; !exists {
			value := ""
			if !oldTag.Value.IsNull() {
				value = oldTag.Value.ValueString()
			}

			err := client.DeleteMonitorTagWithValue(ctx, oldTag.TagID.ValueInt64(), monitorID, value)
			if err != nil {
				diags.AddError(
					fmt.Sprintf("failed to remove tag %d from monitor %d", oldTag.TagID.ValueInt64(), monitorID),
					err.Error(),
				)
				return
			}
		}
	}
}

func handleAddedMonitorTags(
	ctx context.Context,
	client *kuma.Client,
	monitorID int64,
	oldTagMap map[string]MonitorTagModel,
	newTagMap map[string]MonitorTagModel,
	diags *diag.Diagnostics,
) {
	for key, newTag := range newTagMap {
		if _, exists := oldTagMap[key]; !exists {
			value := ""
			if !newTag.Value.IsNull() {
				value = newTag.Value.ValueString()
			}

			_, err := client.AddMonitorTag(ctx, newTag.TagID.ValueInt64(), monitorID, value)
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
