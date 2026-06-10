package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorSystemServiceResource{}
	_ resource.ResourceWithImportState = &MonitorSystemServiceResource{}
)

// NewMonitorSystemServiceResource returns a new instance of the System Service monitor resource.
func NewMonitorSystemServiceResource() resource.Resource {
	return &MonitorSystemServiceResource{}
}

// MonitorSystemServiceResource defines the resource implementation.
type MonitorSystemServiceResource struct {
	client *kuma.Client
}

// MonitorSystemServiceResourceModel describes the resource data model.
type MonitorSystemServiceResourceModel struct {
	MonitorBaseModel

	SystemServiceName types.String `tfsdk:"system_service_name"`
}

// Metadata returns the metadata for the resource.
func (*MonitorSystemServiceResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_system_service"
}

// Schema returns the schema for the resource.
func (*MonitorSystemServiceResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "System Service monitor resource. Checks a systemd service (Linux) or " +
			"SCM service (Windows).",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"system_service_name": schema.StringAttribute{
				MarkdownDescription: "Name of the service to check. On Linux (systemd), this is the unit " +
					"name (e.g. `nginx.service`, `sshd@0.service`); on Windows, this is the SCM service " +
					"name (e.g. `Spooler`).",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the System Service monitor resource with the API client.
func (r *MonitorSystemServiceResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new System Service monitor resource.
func (r *MonitorSystemServiceResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorSystemServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	systemServiceMonitor := monitor.SystemService{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		SystemServiceDetails: monitor.SystemServiceDetails{
			SystemServiceName: data.SystemServiceName.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		systemServiceMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		systemServiceMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		systemServiceMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, &systemServiceMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create System Service monitor", err.Error())
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

// Read reads the current state of the System Service monitor resource.
func (r *MonitorSystemServiceResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorSystemServiceResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var systemServiceMonitor monitor.SystemService
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &systemServiceMonitor)
	// Handle error.
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read System Service monitor", err.Error())
		return
	}

	if actual := systemServiceMonitor.Base.Type(); actual != "" && actual != systemServiceMonitor.Type() {
		tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
			"id":            data.ID.ValueInt64(),
			"expected_type": systemServiceMonitor.Type(),
			"actual_type":   actual,
		})
		resp.State.RemoveResource(ctx)

		return
	}

	data.Name = types.StringValue(systemServiceMonitor.Name)
	if systemServiceMonitor.Description != nil {
		data.Description = types.StringValue(*systemServiceMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(systemServiceMonitor.Interval)
	data.RetryInterval = types.Int64Value(systemServiceMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(systemServiceMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(systemServiceMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(systemServiceMonitor.UpsideDown)
	data.Active = types.BoolValue(systemServiceMonitor.IsActive)
	data.SystemServiceName = types.StringValue(systemServiceMonitor.SystemServiceName)

	if systemServiceMonitor.Parent != nil {
		data.Parent = types.Int64Value(*systemServiceMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(systemServiceMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, systemServiceMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, systemServiceMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the System Service monitor resource.
func (r *MonitorSystemServiceResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorSystemServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorSystemServiceResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	systemServiceMonitor := monitor.SystemService{
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
		SystemServiceDetails: monitor.SystemServiceDetails{
			SystemServiceName: data.SystemServiceName.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		systemServiceMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		systemServiceMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		systemServiceMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, &systemServiceMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update System Service monitor", err.Error())
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

// Delete deletes the System Service monitor resource.
func (r *MonitorSystemServiceResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorSystemServiceResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete System Service monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorSystemServiceResource) ImportState(
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
