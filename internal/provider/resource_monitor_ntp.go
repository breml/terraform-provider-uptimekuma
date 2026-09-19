package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorNTPResource{}
	_ resource.ResourceWithImportState = &MonitorNTPResource{}
)

// defaultNTPTimeout is the query timeout in seconds the client sends when the
// timeout is not configured. The monitor.timeout column is NOT NULL, so an
// unset timeout is stored as this value and reads back as it, which makes it
// the only default that does not produce a perpetual diff.
const defaultNTPTimeout = 10

// NewMonitorNTPResource returns a new instance of the NTP monitor resource.
func NewMonitorNTPResource() resource.Resource {
	return &MonitorNTPResource{}
}

// MonitorNTPResource defines the resource implementation for NTP monitors.
type MonitorNTPResource struct {
	client *kuma.Client
}

// MonitorNTPResourceModel describes the resource data model for NTP monitors.
type MonitorNTPResourceModel struct {
	MonitorBaseModel

	// Hostname is the NTP server to query.
	Hostname types.String `tfsdk:"hostname"`
	// Port is the UDP port of the NTP server.
	Port types.Int64 `tfsdk:"port"`
	// Timeout is the query timeout in seconds.
	Timeout types.Float64 `tfsdk:"timeout"`
	// NTPStratumThreshold is the stratum at which the monitor is considered down.
	NTPStratumThreshold types.Int64 `tfsdk:"ntp_stratum_threshold"`
	// NTPTimeOffsetThreshold is the absolute time offset in milliseconds at which the monitor is down.
	NTPTimeOffsetThreshold types.Int64 `tfsdk:"ntp_time_offset_threshold"`
	// NTPRootDispersionThreshold is the root dispersion in milliseconds at which the monitor is down.
	NTPRootDispersionThreshold types.Int64 `tfsdk:"ntp_root_dispersion_threshold"`
}

// Metadata returns the metadata for the resource.
func (*MonitorNTPResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_ntp"
}

// Schema returns the schema for the resource.
func (*MonitorNTPResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "NTP (Network Time Protocol) monitor resource. The monitor queries an NTP " +
			"server and goes down when the reported stratum, the absolute time offset or the root " +
			"dispersion reaches the configured threshold.\n\n" +
			"The Uptime Kuma web UI suggests an `interval` of 300 seconds for NTP monitors, because " +
			"public NTP servers rate-limit frequent queries. That is a UI convention only and is not " +
			"enforced here, but it is a sensible value for public servers.",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "NTP server IP address or hostname",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "UDP port of the NTP server. While unset, the check falls back to 123.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"timeout": schema.Float64Attribute{
				MarkdownDescription: "Query timeout in seconds, between 1 and 3600. Fractional values are " +
					"supported and round-trip unchanged.",
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(defaultNTPTimeout),
				Validators: []validator.Float64{
					float64validator.Between(1, 3600),
				},
			},
			"ntp_stratum_threshold": schema.Int64Attribute{
				MarkdownDescription: "Stratum at which the monitor is considered down. The check fails " +
					"when the reported stratum is greater than or equal to this value, so a threshold " +
					"of 5 already rejects stratum 5. Stratum 16 is down regardless of the threshold. " +
					"While unset, the check falls back to 5.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 15),
				},
			},
			"ntp_time_offset_threshold": schema.Int64Attribute{
				MarkdownDescription: "Absolute time offset in milliseconds at which the monitor is " +
					"considered down. The check fails when the absolute offset is greater than or " +
					"equal to this value. While unset, the check falls back to 1000.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"ntp_root_dispersion_threshold": schema.Int64Attribute{
				MarkdownDescription: "Root dispersion in milliseconds at which the monitor is considered " +
					"down. The check fails when the root dispersion is greater than or equal to this " +
					"value. While unset, the check falls back to 500.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the NTP monitor resource with the API client.
func (r *MonitorNTPResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new NTP monitor resource.
func (r *MonitorNTPResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorNTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ntpMonitor := buildNTPMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &ntpMonitor)
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create NTP monitor") {
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// The monitor exists, so record it rather than leaving it unmanaged.
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

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

// Read reads the current state of the NTP monitor resource.
func (r *MonitorNTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorNTPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var ntpMonitor monitor.NTP
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &ntpMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read NTP monitor", err.Error())
		return
	}

	if actual := ntpMonitor.Base.Type(); actual != "" && actual != ntpMonitor.Type() {
		tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
			"id":            data.ID.ValueInt64(),
			"expected_type": ntpMonitor.Type(),
			"actual_type":   actual,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	populateNTPModel(&ntpMonitor, &data)
	populateNTPOptionalFields(ctx, &ntpMonitor, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the NTP monitor resource.
func (r *MonitorNTPResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorNTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorNTPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ntpMonitor := buildNTPMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ntpMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &ntpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update NTP monitor", err.Error())
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

// Delete deletes the NTP monitor resource.
func (r *MonitorNTPResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorNTPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete NTP monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorNTPResource) ImportState(
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

// buildNTPMonitor constructs an NTP monitor API object from the Terraform resource model.
func buildNTPMonitor(
	ctx context.Context,
	data *MonitorNTPResourceModel,
	diags *diag.Diagnostics,
) monitor.NTP {
	ntpMonitor := monitor.NTP{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		NTPDetails: monitor.NTPDetails{
			Hostname:                   data.Hostname.ValueString(),
			Port:                       optionalInt64Pointer(data.Port),
			NTPStratumThreshold:        optionalInt64Pointer(data.NTPStratumThreshold),
			NTPTimeOffsetThreshold:     optionalInt64Pointer(data.NTPTimeOffsetThreshold),
			NTPRootDispersionThreshold: optionalInt64Pointer(data.NTPRootDispersionThreshold),
		},
	}

	if !data.Timeout.IsNull() && !data.Timeout.IsUnknown() {
		timeout := data.Timeout.ValueFloat64()
		ntpMonitor.Timeout = &timeout
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		ntpMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		ntpMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if diags.HasError() {
			return ntpMonitor
		}

		ntpMonitor.NotificationIDs = notificationIDs
	}

	return ntpMonitor
}

// optionalInt64Pointer converts an optional Terraform Int64 into the pointer the client expects.
// A null or unknown value yields nil, which the server stores as SQL NULL so the check applies
// its own fallback.
func optionalInt64Pointer(value types.Int64) *int64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	v := value.ValueInt64()

	return &v
}

// optionalInt64Value converts a pointer returned by the client into a Terraform Int64.
func optionalInt64Value(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}

	return types.Int64Value(*value)
}

// populateNTPModel populates the base fields of the Terraform model from the API response.
func populateNTPModel(ntpMonitor *monitor.NTP, data *MonitorNTPResourceModel) {
	data.Name = types.StringValue(ntpMonitor.Name)
	if ntpMonitor.Description != nil {
		data.Description = types.StringValue(*ntpMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(ntpMonitor.Interval)
	data.RetryInterval = types.Int64Value(ntpMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(ntpMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(ntpMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(ntpMonitor.UpsideDown)
	data.Active = types.BoolValue(ntpMonitor.IsActive)
	data.Hostname = types.StringValue(ntpMonitor.Hostname)
	data.Port = optionalInt64Value(ntpMonitor.Port)
	data.NTPStratumThreshold = optionalInt64Value(ntpMonitor.NTPStratumThreshold)
	data.NTPTimeOffsetThreshold = optionalInt64Value(ntpMonitor.NTPTimeOffsetThreshold)
	data.NTPRootDispersionThreshold = optionalInt64Value(ntpMonitor.NTPRootDispersionThreshold)

	if ntpMonitor.Timeout != nil {
		data.Timeout = types.Float64Value(*ntpMonitor.Timeout)
	} else {
		data.Timeout = types.Float64Value(defaultNTPTimeout)
	}
}

// populateNTPOptionalFields populates optional and computed fields from the API response.
func populateNTPOptionalFields(
	ctx context.Context,
	ntpMonitor *monitor.NTP,
	data *MonitorNTPResourceModel,
	diags *diag.Diagnostics,
) {
	if ntpMonitor.Parent != nil {
		data.Parent = types.Int64Value(*ntpMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(ntpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, ntpMonitor.NotificationIDs)
		diags.Append(d...)
		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, ntpMonitor.Tags, data.Tags, diags)
}
