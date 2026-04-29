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
	_ resource.Resource                = &MonitorTailscalePingResource{}
	_ resource.ResourceWithImportState = &MonitorTailscalePingResource{}
)

// NewMonitorTailscalePingResource returns a new instance of the Tailscale Ping monitor resource.
func NewMonitorTailscalePingResource() resource.Resource {
	return &MonitorTailscalePingResource{}
}

// MonitorTailscalePingResource defines the resource implementation.
type MonitorTailscalePingResource struct {
	client *kuma.Client
}

// MonitorTailscalePingResourceModel describes the resource data model.
type MonitorTailscalePingResourceModel struct {
	MonitorBaseModel

	Hostname types.String `tfsdk:"hostname"`
}

// Metadata returns the metadata for the resource.
func (*MonitorTailscalePingResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_tailscale_ping"
}

// Schema returns the schema for the resource.
func (*MonitorTailscalePingResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Tailscale Ping monitor resource. Checks connectivity to Tailscale VPN nodes and measures latency.",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Tailscale hostname or IP address to ping",
				Required:            true,
			},
		}),
	}
}

// Configure configures the Tailscale Ping monitor resource with the API client.
func (r *MonitorTailscalePingResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Tailscale Ping monitor resource.
func (r *MonitorTailscalePingResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorTailscalePingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tailscalePingMonitor := monitor.TailscalePing{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		TailscalePingDetails: monitor.TailscalePingDetails{
			Hostname: data.Hostname.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		tailscalePingMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		tailscalePingMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		tailscalePingMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, &tailscalePingMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create Tailscale Ping monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = handleMonitorActiveStateCreate(ctx, r.client, id, data.Active)
	if err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError("failed to apply monitor active state", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Tailscale Ping monitor resource.
func (r *MonitorTailscalePingResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorTailscalePingResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var tailscalePingMonitor monitor.TailscalePing
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &tailscalePingMonitor)
	// Handle error.
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read Tailscale Ping monitor", err.Error())
		return
	}

	data.Name = types.StringValue(tailscalePingMonitor.Name)
	if tailscalePingMonitor.Description != nil {
		data.Description = types.StringValue(*tailscalePingMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(tailscalePingMonitor.Interval)
	data.RetryInterval = types.Int64Value(tailscalePingMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(tailscalePingMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(tailscalePingMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(tailscalePingMonitor.UpsideDown)
	data.Active = types.BoolValue(tailscalePingMonitor.IsActive)
	data.Hostname = types.StringValue(tailscalePingMonitor.Hostname)

	if tailscalePingMonitor.Parent != nil {
		data.Parent = types.Int64Value(*tailscalePingMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(tailscalePingMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, tailscalePingMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, tailscalePingMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Tailscale Ping monitor resource.
func (r *MonitorTailscalePingResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorTailscalePingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorTailscalePingResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tailscalePingMonitor := monitor.TailscalePing{
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
		TailscalePingDetails: monitor.TailscalePingDetails{
			Hostname: data.Hostname.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		tailscalePingMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		tailscalePingMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		tailscalePingMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, &tailscalePingMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update Tailscale Ping monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	handleMonitorActiveStateUpdate(ctx, r.client, data.ID.ValueInt64(), state.Active, data.Active, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Tailscale Ping monitor resource.
func (r *MonitorTailscalePingResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorTailscalePingResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Tailscale Ping monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorTailscalePingResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
