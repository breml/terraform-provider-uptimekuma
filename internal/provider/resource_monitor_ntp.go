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

// defaultNTPTimeout is the query timeout in seconds the NTP check falls back to
// when no timeout is stored. It is used both as the schema default and as the
// read fallback: the monitor.timeout column is NOT NULL, so this value is stored
// and reads back unchanged, which makes it the only NTP fallback that can be
// modelled as a Terraform default without causing a perpetual diff.
//
// The value mirrors the unexported defaultNTPTimeout in the client library
// (monitor/monitor_ntp.go); re-check it when bumping go-uptime-kuma-client.
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

	Hostname                   types.String  `tfsdk:"hostname"`                      // NTP server to query.
	Port                       types.Int64   `tfsdk:"port"`                          // UDP port of the NTP server.
	Timeout                    types.Float64 `tfsdk:"timeout"`                       // Query timeout in seconds.
	NTPStratumThreshold        types.Int64   `tfsdk:"ntp_stratum_threshold"`         // Stratum threshold.
	NTPTimeOffsetThreshold     types.Int64   `tfsdk:"ntp_time_offset_threshold"`     // Offset threshold in ms.
	NTPRootDispersionThreshold types.Int64   `tfsdk:"ntp_root_dispersion_threshold"` // Dispersion threshold in ms.
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
			"The Uptime Kuma web UI sets `interval` to 300 seconds for new NTP monitors, because " +
			"public NTP servers rate-limit frequent queries. This provider does not special-case " +
			"NTP and keeps the shared default of 60 seconds, so set `interval = 300` explicitly " +
			"when monitoring public pool servers.\n\n" +
			"`port` and the three thresholds are stored as SQL NULL while unset, which is a " +
			"different state from setting them to the value the check falls back to. Leave them " +
			"unset to follow the check's own fallbacks.",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "NTP server IP address or hostname",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "UDP port of the NTP server, between 1 and 65535. While unset, " +
					"the check falls back to 123.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"timeout": schema.Float64Attribute{
				MarkdownDescription: "Query timeout in seconds, between 1 and 3600. Fractional values are " +
					"supported and round-trip unchanged. Defaults to 10, the value the NTP check itself " +
					"falls back to when no timeout is stored. Note that the Uptime Kuma web UI pre-fills " +
					"48 for new NTP monitors, so a monitor created there and then imported reports 48.",
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(defaultNTPTimeout),
				Validators: []validator.Float64{
					float64validator.Between(1, 3600),
				},
			},
			"ntp_stratum_threshold": schema.Int64Attribute{
				MarkdownDescription: "Stratum at which the monitor is considered down, between 1 and 15 " +
					"(the range the Uptime Kuma web UI allows). The check fails when the reported " +
					"stratum is greater than or equal to this value, so a threshold of 5 already " +
					"rejects stratum 5. Stratum 16 is down regardless of the threshold. While unset, " +
					"the check falls back to 5.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 15),
				},
			},
			"ntp_time_offset_threshold": schema.Int64Attribute{
				MarkdownDescription: "Absolute time offset in milliseconds at which the monitor is " +
					"considered down, at least 1. The check fails when the absolute offset is greater " +
					"than or equal to this value. While unset, the check falls back to 1000. `0` is " +
					"rejected: the check treats 0 as unset and would silently apply the fallback, so " +
					"leave the attribute unset to request it.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"ntp_root_dispersion_threshold": schema.Int64Attribute{
				MarkdownDescription: "Root dispersion in milliseconds at which the monitor is considered " +
					"down, at least 1. The check fails when the root dispersion is greater than or " +
					"equal to this value. While unset, the check falls back to 500. `0` is rejected: " +
					"the check treats 0 as unset and would silently apply the fallback, so leave the " +
					"attribute unset to request it.",
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

	err = handleMonitorActiveStateCreate(ctx, r.client, id, data.Active, &resp.Diagnostics)
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
		resp.Diagnostics.AddWarning(
			"Monitor type changed outside Terraform",
			fmt.Sprintf(
				"Monitor %d is of type %q but is managed as %q, so it was removed from state and "+
					"Terraform will plan to create a replacement. Manage it with the resource type "+
					"matching %q, or remove it from the configuration, to avoid a duplicate.",
				data.ID.ValueInt64(), actual, ntpMonitor.Type(), actual,
			),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	populateNTPModel(ctx, &ntpMonitor, &data)
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
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update NTP monitor") {
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
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete NTP monitor")
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
			Port:                       int64ToPtr(data.Port),
			NTPStratumThreshold:        int64ToPtr(data.NTPStratumThreshold),
			NTPTimeOffsetThreshold:     int64ToPtr(data.NTPTimeOffsetThreshold),
			NTPRootDispersionThreshold: int64ToPtr(data.NTPRootDispersionThreshold),
		},
	}

	ntpMonitor.Timeout = float64ToPtr(data.Timeout)

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

// populateNTPModel populates the base fields of the Terraform model from the API response.
func populateNTPModel(ctx context.Context, ntpMonitor *monitor.NTP, data *MonitorNTPResourceModel) {
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
	data.Port = int64PtrToTypes(ntpMonitor.Port)
	data.NTPStratumThreshold = int64PtrToTypes(ntpMonitor.NTPStratumThreshold)
	data.NTPTimeOffsetThreshold = int64PtrToTypes(ntpMonitor.NTPTimeOffsetThreshold)
	data.NTPRootDispersionThreshold = int64PtrToTypes(ntpMonitor.NTPRootDispersionThreshold)
	data.Timeout = ntpTimeoutValue(ctx, ntpMonitor)
}

// ntpTimeoutValue converts the timeout returned by the client into a Terraform Float64.
//
// The monitor.timeout column is NOT NULL, so a nil pointer means the invariant behind
// defaultNTPTimeout no longer holds. Substituting the fallback keeps the Computed attribute
// consistent after apply, but the substitution is logged so a broken invariant is not silent.
func ntpTimeoutValue(ctx context.Context, ntpMonitor *monitor.NTP) types.Float64 {
	if ntpMonitor.Timeout != nil {
		return types.Float64Value(*ntpMonitor.Timeout)
	}

	tflog.Warn(ctx, "NTP monitor returned a null timeout, assuming the check fallback", map[string]any{
		"id":      ntpMonitor.ID,
		"assumed": defaultNTPTimeout,
	})

	return types.Float64Value(defaultNTPTimeout)
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

		if diags.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, ntpMonitor.Tags, data.Tags, diags)
}
