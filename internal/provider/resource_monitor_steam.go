package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &MonitorSteamResource{}
	_ resource.ResourceWithImportState = &MonitorSteamResource{}
)

// NewMonitorSteamResource returns a new instance of the Steam monitor resource.
func NewMonitorSteamResource() resource.Resource {
	return &MonitorSteamResource{}
}

// MonitorSteamResource defines the resource implementation for Steam game server monitors.
type MonitorSteamResource struct {
	client *kuma.Client
}

// MonitorSteamResourceModel describes the resource data model for Steam monitors.
type MonitorSteamResourceModel struct {
	MonitorBaseModel

	Hostname types.String `tfsdk:"hostname"`
	Port     types.Int64  `tfsdk:"port"`
	Timeout  types.Int64  `tfsdk:"timeout"`
}

// Metadata returns the metadata for the resource.
func (*MonitorSteamResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_steam"
}

// Schema returns the schema for the resource.
func (*MonitorSteamResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Steam game server monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Steam game server IP address or hostname",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Steam game server port",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Request timeout in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(48),
				Validators: []validator.Int64{
					int64validator.Between(1, 3600),
				},
			},
		}),
	}
}

// Configure configures the Steam monitor resource with the API client.
func (r *MonitorSteamResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Steam monitor resource.
func (r *MonitorSteamResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorSteamResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	steamMonitor := buildSteamMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &steamMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Steam monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Steam monitor resource.
func (r *MonitorSteamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorSteamResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var steamMonitor monitor.Steam
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &steamMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read Steam monitor", err.Error())
		return
	}

	populateSteamModel(&steamMonitor, &data)
	populateSteamOptionalFields(ctx, &steamMonitor, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Steam monitor resource.
func (r *MonitorSteamResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorSteamResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorSteamResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	steamMonitor := buildSteamMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	steamMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &steamMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update Steam monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Steam monitor resource.
func (r *MonitorSteamResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorSteamResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Steam monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorSteamResource) ImportState(
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

// buildSteamMonitor constructs a Steam monitor API object from the Terraform resource model.
func buildSteamMonitor(
	ctx context.Context,
	data *MonitorSteamResourceModel,
	diags *diag.Diagnostics,
) monitor.Steam {
	steamMonitor := monitor.Steam{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		SteamDetails: monitor.SteamDetails{
			Hostname: data.Hostname.ValueString(),
			Port:     int(data.Port.ValueInt64()),
		},
	}

	if !data.Timeout.IsNull() && !data.Timeout.IsUnknown() {
		timeout := data.Timeout.ValueInt64()
		steamMonitor.Timeout = &timeout
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		steamMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		steamMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if diags.HasError() {
			return steamMonitor
		}

		steamMonitor.NotificationIDs = notificationIDs
	}

	return steamMonitor
}

// populateSteamModel populates the base fields of the Terraform model from the API response.
func populateSteamModel(steamMonitor *monitor.Steam, data *MonitorSteamResourceModel) {
	data.Name = types.StringValue(steamMonitor.Name)
	if steamMonitor.Description != nil {
		data.Description = types.StringValue(*steamMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(steamMonitor.Interval)
	data.RetryInterval = types.Int64Value(steamMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(steamMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(steamMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(steamMonitor.UpsideDown)
	data.Active = types.BoolValue(steamMonitor.IsActive)
	data.Hostname = types.StringValue(steamMonitor.Hostname)
	data.Port = types.Int64Value(int64(steamMonitor.Port))

	if steamMonitor.Timeout != nil {
		data.Timeout = types.Int64Value(*steamMonitor.Timeout)
	} else {
		data.Timeout = types.Int64Null()
	}
}

// populateSteamOptionalFields populates optional and computed fields from the API response.
func populateSteamOptionalFields(
	ctx context.Context,
	steamMonitor *monitor.Steam,
	data *MonitorSteamResourceModel,
	diags *diag.Diagnostics,
) {
	if steamMonitor.Parent != nil {
		data.Parent = types.Int64Value(*steamMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(steamMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, steamMonitor.NotificationIDs)
		diags.Append(d...)
		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, steamMonitor.Tags, data.Tags, diags)
}
