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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorGameDigResource{}
	_ resource.ResourceWithImportState = &MonitorGameDigResource{}
)

// NewMonitorGameDigResource returns a new instance of the GameDig monitor resource.
func NewMonitorGameDigResource() resource.Resource {
	return &MonitorGameDigResource{}
}

// MonitorGameDigResource defines the resource implementation for GameDig game server monitors.
type MonitorGameDigResource struct {
	client *kuma.Client
}

// MonitorGameDigResourceModel describes the resource data model for GameDig monitors.
type MonitorGameDigResourceModel struct {
	MonitorBaseModel

	// Hostname is the game server host or IP address.
	Hostname types.String `tfsdk:"hostname"`
	// Port is the game server query port.
	Port types.Int64 `tfsdk:"port"`
	// Game is the game type identifier (e.g. minecraft, csgo).
	Game types.String `tfsdk:"game"`
	// GameDigGivenPortOnly indicates whether to use only the given port without auto-detection.
	GameDigGivenPortOnly types.Bool `tfsdk:"gamedig_given_port_only"`
}

// Metadata returns the metadata for the resource.
func (*MonitorGameDigResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_gamedig"
}

// Schema returns the schema for the resource.
func (*MonitorGameDigResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "GameDig game server monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Game server IP address or hostname",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Game server port",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"game": schema.StringAttribute{
				MarkdownDescription: "Game type identifier (e.g. minecraft, csgo)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"gamedig_given_port_only": schema.BoolAttribute{
				MarkdownDescription: "Use only the given port without auto-detection",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		}),
	}
}

// Configure configures the GameDig monitor resource with the API client.
func (r *MonitorGameDigResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new GameDig monitor resource.
func (r *MonitorGameDigResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorGameDigResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	gameDigMonitor := buildGameDigMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &gameDigMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create GameDig monitor", err.Error())
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the GameDig monitor resource.
func (r *MonitorGameDigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorGameDigResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var gameDigMonitor monitor.GameDig
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &gameDigMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read GameDig monitor", err.Error())
		return
	}

	if actual := gameDigMonitor.Base.Type(); actual != "" && actual != gameDigMonitor.Type() {
		tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
			"id":            data.ID.ValueInt64(),
			"expected_type": gameDigMonitor.Type(),
			"actual_type":   actual,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	populateGameDigModel(&gameDigMonitor, &data)
	populateGameDigOptionalFields(ctx, &gameDigMonitor, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the GameDig monitor resource.
func (r *MonitorGameDigResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorGameDigResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorGameDigResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gameDigMonitor := buildGameDigMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	gameDigMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &gameDigMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update GameDig monitor", err.Error())
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the GameDig monitor resource.
func (r *MonitorGameDigResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorGameDigResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete GameDig monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorGameDigResource) ImportState(
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

// buildGameDigMonitor constructs a GameDig monitor API object from the Terraform resource model.
func buildGameDigMonitor(
	ctx context.Context,
	data *MonitorGameDigResourceModel,
	diags *diag.Diagnostics,
) monitor.GameDig {
	gameDigMonitor := monitor.GameDig{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		GameDigDetails: monitor.GameDigDetails{
			Hostname:             data.Hostname.ValueString(),
			Port:                 int(data.Port.ValueInt64()),
			Game:                 data.Game.ValueString(),
			GameDigGivenPortOnly: data.GameDigGivenPortOnly.ValueBool(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		gameDigMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		gameDigMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if diags.HasError() {
			return gameDigMonitor
		}

		gameDigMonitor.NotificationIDs = notificationIDs
	}

	return gameDigMonitor
}

// populateGameDigModel populates the base fields of the Terraform model from the API response.
func populateGameDigModel(gameDigMonitor *monitor.GameDig, data *MonitorGameDigResourceModel) {
	data.Name = types.StringValue(gameDigMonitor.Name)
	if gameDigMonitor.Description != nil {
		data.Description = types.StringValue(*gameDigMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(gameDigMonitor.Interval)
	data.RetryInterval = types.Int64Value(gameDigMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(gameDigMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(gameDigMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(gameDigMonitor.UpsideDown)
	data.Active = types.BoolValue(gameDigMonitor.IsActive)
	data.Hostname = types.StringValue(gameDigMonitor.Hostname)
	data.Port = types.Int64Value(int64(gameDigMonitor.Port))
	data.Game = types.StringValue(gameDigMonitor.Game)
	data.GameDigGivenPortOnly = types.BoolValue(gameDigMonitor.GameDigGivenPortOnly)
}

// populateGameDigOptionalFields populates optional and computed fields from the API response.
func populateGameDigOptionalFields(
	ctx context.Context,
	gameDigMonitor *monitor.GameDig,
	data *MonitorGameDigResourceModel,
	diags *diag.Diagnostics,
) {
	if gameDigMonitor.Parent != nil {
		data.Parent = types.Int64Value(*gameDigMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(gameDigMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, gameDigMonitor.NotificationIDs)
		diags.Append(d...)
		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, gameDigMonitor.Tags, data.Tags, diags)
}
