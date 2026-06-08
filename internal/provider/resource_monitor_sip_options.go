package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ resource.Resource                = &MonitorSIPOptionsResource{}
	_ resource.ResourceWithImportState = &MonitorSIPOptionsResource{}
)

// NewMonitorSIPOptionsResource returns a new instance of the SIP Options monitor resource.
func NewMonitorSIPOptionsResource() resource.Resource {
	return &MonitorSIPOptionsResource{}
}

// MonitorSIPOptionsResource defines the resource implementation.
type MonitorSIPOptionsResource struct {
	client *kuma.Client
}

// MonitorSIPOptionsResourceModel describes the resource data model.
type MonitorSIPOptionsResourceModel struct {
	MonitorBaseModel

	Hostname types.String `tfsdk:"hostname"`
	Port     types.Int64  `tfsdk:"port"`
}

// Metadata returns the metadata for the resource.
func (*MonitorSIPOptionsResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_sip_options"
}

// Schema returns the schema for the resource.
func (*MonitorSIPOptionsResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SIP Options monitor resource. Sends a SIP OPTIONS request to a host/port. " +
			"Note: upstream uses `sipsak` and only works on non-container installs of Uptime Kuma.",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address to monitor",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "SIP port number to monitor",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
		}),
	}
}

// Configure configures the SIP Options monitor resource with the API client.
func (r *MonitorSIPOptionsResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new SIP Options monitor resource.
func (r *MonitorSIPOptionsResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorSIPOptionsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sipOptionsMonitor := monitor.SIPOptions{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		SIPOptionsDetails: monitor.SIPOptionsDetails{
			Hostname: data.Hostname.ValueString(),
			Port:     int(data.Port.ValueInt64()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		sipOptionsMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		sipOptionsMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		sipOptionsMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, &sipOptionsMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create SIP Options monitor", err.Error())
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
		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.AddError("failed to apply monitor active state", err.Error())
		}

		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the SIP Options monitor resource.
func (r *MonitorSIPOptionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorSIPOptionsResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var sipOptionsMonitor monitor.SIPOptions
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &sipOptionsMonitor)
	// Handle error.
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read SIP Options monitor", err.Error())
		return
	}

	if actual := sipOptionsMonitor.Base.Type(); actual != "" && actual != sipOptionsMonitor.Type() {
		tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
			"id":            data.ID.ValueInt64(),
			"expected_type": sipOptionsMonitor.Type(),
			"actual_type":   actual,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(sipOptionsMonitor.Name)
	if sipOptionsMonitor.Description != nil {
		data.Description = types.StringValue(*sipOptionsMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(sipOptionsMonitor.Interval)
	data.RetryInterval = types.Int64Value(sipOptionsMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(sipOptionsMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(sipOptionsMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(sipOptionsMonitor.UpsideDown)
	data.Active = types.BoolValue(sipOptionsMonitor.IsActive)
	data.Hostname = types.StringValue(sipOptionsMonitor.Hostname)
	data.Port = types.Int64Value(int64(sipOptionsMonitor.Port))

	if sipOptionsMonitor.Parent != nil {
		data.Parent = types.Int64Value(*sipOptionsMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(sipOptionsMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, sipOptionsMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, sipOptionsMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SIP Options monitor resource.
func (r *MonitorSIPOptionsResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorSIPOptionsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorSIPOptionsResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sipOptionsMonitor := monitor.SIPOptions{
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
		SIPOptionsDetails: monitor.SIPOptionsDetails{
			Hostname: data.Hostname.ValueString(),
			Port:     int(data.Port.ValueInt64()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		sipOptionsMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		sipOptionsMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		sipOptionsMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, &sipOptionsMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update SIP Options monitor", err.Error())
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

// Delete deletes the SIP Options monitor resource.
func (r *MonitorSIPOptionsResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorSIPOptionsResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete SIP Options monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorSIPOptionsResource) ImportState(
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
