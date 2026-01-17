package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorPingResource{}
	_ resource.ResourceWithImportState = &MonitorPingResource{}
)

// NewMonitorPingResource returns a new instance of the Ping monitor resource.
func NewMonitorPingResource() resource.Resource {
	return &MonitorPingResource{}
}

// MonitorPingResource defines the resource implementation.
type MonitorPingResource struct {
	client *kuma.Client
}

// MonitorPingResourceModel describes the resource data model.
type MonitorPingResourceModel struct {
	MonitorBaseModel

	Hostname   types.String `tfsdk:"hostname"`
	PacketSize types.Int64  `tfsdk:"packet_size"`
}

// Metadata returns the metadata for the resource.
func (*MonitorPingResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_ping"
}

// Schema returns the schema for the resource.
func (*MonitorPingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Ping monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address to ping",
				Required:            true,
			},
			"packet_size": schema.Int64Attribute{
				MarkdownDescription: "Ping packet size in bytes",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(56),
				Validators: []validator.Int64{
					int64validator.Between(1, 65500),
				},
			},
		}),
	}
}

// Configure configures the Ping monitor resource with the API client.
func (r *MonitorPingResource) Configure(
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

// Create creates a new Ping monitor resource.
func (r *MonitorPingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorPingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pingMonitor := monitor.Ping{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		PingDetails: monitor.PingDetails{
			Hostname:   data.Hostname.ValueString(),
			PacketSize: int(data.PacketSize.ValueInt64()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		pingMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		pingMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		pingMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, &pingMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create Ping monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Ping monitor resource.
func (r *MonitorPingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorPingResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var pingMonitor monitor.Ping
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &pingMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to read Ping monitor", err.Error())
		return
	}

	data.Name = types.StringValue(pingMonitor.Name)
	if pingMonitor.Description != nil {
		data.Description = types.StringValue(*pingMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(pingMonitor.Interval)
	data.RetryInterval = types.Int64Value(pingMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(pingMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(pingMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(pingMonitor.UpsideDown)
	data.Active = types.BoolValue(pingMonitor.IsActive)
	data.Hostname = types.StringValue(pingMonitor.Hostname)
	data.PacketSize = types.Int64Value(int64(pingMonitor.PacketSize))

	if pingMonitor.Parent != nil {
		data.Parent = types.Int64Value(*pingMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(pingMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, pingMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, pingMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Ping monitor resource.
func (r *MonitorPingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorPingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorPingResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pingMonitor := monitor.Ping{
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
		PingDetails: monitor.PingDetails{
			Hostname:   data.Hostname.ValueString(),
			PacketSize: int(data.PacketSize.ValueInt64()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		pingMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		pingMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		pingMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, &pingMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update Ping monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Ping monitor resource.
func (r *MonitorPingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorPingResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Ping monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorPingResource) ImportState(
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
