package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	// Ensure MonitorRadiusResource satisfies various resource interfaces.
	_ resource.Resource                = &MonitorRadiusResource{}
	_ resource.ResourceWithImportState = &MonitorRadiusResource{}
)

// NewMonitorRadiusResource returns a new instance of the Radius monitor resource.
func NewMonitorRadiusResource() resource.Resource {
	return &MonitorRadiusResource{}
}

// MonitorRadiusResource defines the resource implementation.
type MonitorRadiusResource struct {
	client *kuma.Client
}

// MonitorRadiusResourceModel describes the resource data model for Radius monitors.
type MonitorRadiusResourceModel struct {
	MonitorBaseModel

	Hostname         types.String `tfsdk:"hostname"`
	Port             types.Int64  `tfsdk:"port"`
	RadiusUsername   types.String `tfsdk:"radius_username"`
	RadiusPassword   types.String `tfsdk:"radius_password"`
	RadiusSecret     types.String `tfsdk:"radius_secret"`
	CalledStationID  types.String `tfsdk:"called_station_id"`
	CallingStationID types.String `tfsdk:"calling_station_id"`
}

// Metadata returns the metadata for the resource.
func (*MonitorRadiusResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_radius"
}

// Schema returns the schema for the resource.
func (*MonitorRadiusResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Radius monitor resource for testing Radius server authentication.",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Radius server hostname or IP address",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Radius server port",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1812),
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
			"radius_username": schema.StringAttribute{
				MarkdownDescription: "Username for Radius authentication",
				Required:            true,
			},
			"radius_password": schema.StringAttribute{
				MarkdownDescription: "Password for Radius authentication",
				Required:            true,
				Sensitive:           true,
			},
			"radius_secret": schema.StringAttribute{
				MarkdownDescription: "Shared secret for the Radius server",
				Required:            true,
				Sensitive:           true,
			},
			"called_station_id": schema.StringAttribute{
				MarkdownDescription: "Optional Called-Station-Id attribute",
				Optional:            true,
			},
			"calling_station_id": schema.StringAttribute{
				MarkdownDescription: "Optional Calling-Station-Id attribute",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorRadiusResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Radius monitor resource.
func (r *MonitorRadiusResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorRadiusResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	radiusMonitor := buildRadiusMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &radiusMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Radius monitor", err.Error())
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

// buildRadiusMonitor builds a Radius monitor API object from the resource model.
func buildRadiusMonitor(
	ctx context.Context,
	data *MonitorRadiusResourceModel,
	diags *diag.Diagnostics,
) monitor.Radius {
	port := data.Port.ValueInt64()
	radiusMonitor := monitor.Radius{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		RadiusDetails: monitor.RadiusDetails{
			Hostname: data.Hostname.ValueString(),
			Port:     &port,
			Username: data.RadiusUsername.ValueString(),
			Password: data.RadiusPassword.ValueString(),
			Secret:   data.RadiusSecret.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		radiusMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		radiusMonitor.Parent = &parent
	}

	if !data.CalledStationID.IsNull() {
		calledStationID := data.CalledStationID.ValueString()
		radiusMonitor.CalledStationID = &calledStationID
	}

	if !data.CallingStationID.IsNull() {
		callingStationID := data.CallingStationID.ValueString()
		radiusMonitor.CallingStationID = &callingStationID
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			radiusMonitor.NotificationIDs = notificationIDs
		}
	}

	return radiusMonitor
}

// populateRadiusMonitorBaseFields populates the resource model with data from the Radius monitor API response.
func populateRadiusMonitorBaseFields(radiusMonitor *monitor.Radius, m *MonitorRadiusResourceModel) {
	m.Name = types.StringValue(radiusMonitor.Name)
	if radiusMonitor.Description != nil {
		m.Description = types.StringValue(*radiusMonitor.Description)
	} else {
		m.Description = types.StringNull()
	}

	m.Interval = types.Int64Value(radiusMonitor.Interval)
	m.RetryInterval = types.Int64Value(radiusMonitor.RetryInterval)
	m.ResendInterval = types.Int64Value(radiusMonitor.ResendInterval)
	m.MaxRetries = types.Int64Value(radiusMonitor.MaxRetries)
	m.UpsideDown = types.BoolValue(radiusMonitor.UpsideDown)
	m.Active = types.BoolValue(radiusMonitor.IsActive)
	m.Hostname = types.StringValue(radiusMonitor.Hostname)
	m.RadiusUsername = types.StringValue(radiusMonitor.Username)

	// Uptime Kuma may not return sensitive fields in the API response.
	// Preserve the existing state values to avoid erasing the configured
	// secrets and to prevent perpetual diffs after a refresh.
	if radiusMonitor.Password != "" {
		m.RadiusPassword = types.StringValue(radiusMonitor.Password)
	}

	if radiusMonitor.Secret != "" {
		m.RadiusSecret = types.StringValue(radiusMonitor.Secret)
	}

	m.CalledStationID = stringOrNullPtr(radiusMonitor.CalledStationID)
	m.CallingStationID = stringOrNullPtr(radiusMonitor.CallingStationID)

	if radiusMonitor.Port != nil {
		m.Port = types.Int64Value(*radiusMonitor.Port)
	} else {
		m.Port = types.Int64Value(1812)
	}
}

// populateOptionalFieldsForRadius populates optional parent and notification fields from the Radius monitor API
// response.
func populateOptionalFieldsForRadius(
	ctx context.Context,
	radiusMonitor *monitor.Radius,
	m *MonitorRadiusResourceModel,
	diags *diag.Diagnostics,
) {
	if radiusMonitor.Parent != nil {
		m.Parent = types.Int64Value(*radiusMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	if len(radiusMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, radiusMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// Read reads the current state of the Radius monitor resource.
func (r *MonitorRadiusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorRadiusResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var radiusMonitor monitor.Radius
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &radiusMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read Radius monitor", err.Error())
		return
	}

	if radiusMonitor.Base.Type() != radiusMonitor.Type() {
		resp.State.RemoveResource(ctx)
		return
	}

	populateRadiusMonitorBaseFields(&radiusMonitor, &data)
	populateOptionalFieldsForRadius(ctx, &radiusMonitor, &data, &resp.Diagnostics)

	data.Tags = handleMonitorTagsRead(ctx, radiusMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Radius monitor resource.
func (r *MonitorRadiusResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorRadiusResourceModel
	var state MonitorRadiusResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	radiusMonitor := buildRadiusMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	radiusMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &radiusMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update Radius monitor", err.Error())
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

// Delete deletes the Radius monitor resource.
func (r *MonitorRadiusResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorRadiusResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Radius monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorRadiusResource) ImportState(
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
