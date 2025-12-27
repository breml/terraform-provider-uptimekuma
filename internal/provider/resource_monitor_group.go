package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorGroupResource{}
	_ resource.ResourceWithImportState = &MonitorGroupResource{}
)

// NewMonitorGroupResource returns a new instance of the group monitor resource.
func NewMonitorGroupResource() resource.Resource {
	return &MonitorGroupResource{}
}

// MonitorGroupResource defines the resource implementation.
type MonitorGroupResource struct {
	client *kuma.Client
}

// MonitorGroupResourceModel describes the resource data model.
type MonitorGroupResourceModel struct {
	MonitorBaseModel
}

func (_ *MonitorGroupResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_group"
}

func (_ *MonitorGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Monitor group resource for organizing monitors",
		Attributes:          withMonitorBaseAttributes(map[string]schema.Attribute{}),
	}
}

// Configure configures the group monitor resource with the API client.
func (r *MonitorGroupResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	r.client = client
}

// Create creates a new group monitor resource.
func (r *MonitorGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	groupMonitor := monitor.Group{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		groupMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		groupMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		groupMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, groupMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create group monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the group monitor resource.
func (r *MonitorGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var groupMonitor monitor.Group
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &groupMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read group monitor", err.Error())
		return
	}

	data.Name = types.StringValue(groupMonitor.Name)
	if groupMonitor.Description != nil {
		data.Description = types.StringValue(*groupMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(groupMonitor.Interval)
	data.RetryInterval = types.Int64Value(groupMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(groupMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(groupMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(groupMonitor.UpsideDown)
	data.Active = types.BoolValue(groupMonitor.IsActive)

	if groupMonitor.Parent != nil {
		data.Parent = types.Int64Value(*groupMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(groupMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, groupMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, groupMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the group monitor resource.
func (r *MonitorGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupMonitor := monitor.Group{
		Base: monitor.Base{
			ID:             data.ID.ValueInt64(),
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		groupMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		groupMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		groupMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, groupMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update group monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the group monitor resource.
func (r *MonitorGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete group monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (_ *MonitorGroupResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
