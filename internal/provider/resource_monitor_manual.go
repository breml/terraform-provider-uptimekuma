package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &MonitorManualResource{}
	_ resource.ResourceWithImportState = &MonitorManualResource{}
)

// Terraform values of the manual_status attribute.
const (
	manualStatusUp      = "up"
	manualStatusDown    = "down"
	manualStatusPending = "pending"
)

// NewMonitorManualResource returns a new instance of the manual monitor resource.
func NewMonitorManualResource() resource.Resource {
	return &MonitorManualResource{}
}

// MonitorManualResource defines the resource implementation for manual monitors.
type MonitorManualResource struct {
	client *kuma.Client
}

// MonitorManualResourceModel describes the resource data model for manual monitors.
type MonitorManualResourceModel struct {
	MonitorBaseModel

	ManualStatus types.String `tfsdk:"manual_status"` // Status up, down or pending; never returned by the server.
}

// Metadata returns the metadata for the resource.
func (*MonitorManualResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_manual"
}

// Schema returns the schema for the resource.
func (*MonitorManualResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manual monitor resource. A passive monitor whose status is set by the " +
			"user instead of being determined by a check, available since Uptime Kuma 2.0.0. The base " +
			"attributes still apply on top of the status: `upside_down`, retries and maintenance " +
			"behave as for any other monitor.\n\n" +
			"Uptime Kuma does not return the status it stores, so a status changed in the web UI " +
			"is not detected as drift; Terraform reapplies its value on the next update of the " +
			"resource only, and every update, also of an unrelated attribute, replaces a status " +
			"set in the web UI.",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"manual_status": schema.StringAttribute{
				MarkdownDescription: "Status the monitor reports, one of `up`, `down` or `pending`. " +
					"While unset the monitor reports pending with the message " +
					"`Manual monitoring - No status set`. Every update of the resource sends this " +
					"value, so removing it resets the monitor to pending. Uptime Kuma stores the " +
					"status but never returns it, so the provider keeps the configured value in " +
					"state and cannot detect changes made outside Terraform. After an import it is " +
					"null in state; if the configuration sets it, the next plan shows the change and " +
					"the apply sends it.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(manualStatusUp, manualStatusDown, manualStatusPending),
				},
			},
		}),
	}
}

// Configure configures the manual monitor resource with the API client.
func (r *MonitorManualResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// manualStatusFromString maps a manual_status attribute value to the client
// constant. It reports false for a value outside the schema's enum.
func manualStatusFromString(status string) (monitor.ManualStatus, bool) {
	switch status {
	case manualStatusUp:
		return monitor.ManualStatusUp, true

	case manualStatusDown:
		return monitor.ManualStatusDown, true

	case manualStatusPending:
		return monitor.ManualStatusPending, true

	default:
		return 0, false
	}
}

// manualStatusToString maps a client constant to the manual_status attribute
// value. It reports false for a value the server stored but that has no
// meaning, as the server does not validate it.
func manualStatusToString(status monitor.ManualStatus) (string, bool) {
	switch status {
	case monitor.ManualStatusUp:
		return manualStatusUp, true

	case monitor.ManualStatusDown:
		return manualStatusDown, true

	case monitor.ManualStatusPending:
		return manualStatusPending, true

	default:
		return "", false
	}
}

// readManualStatus returns the manual_status value to store after a read.
//
// Uptime Kuma stores the status but does not send it back. The value from
// state is kept, so neither a refresh nor an import fabricates a status; the
// server value is adopted if a future version begins returning it. A server
// value the provider does not recognize keeps the state value and raises a
// warning, as the server does not validate the stored status.
func readManualStatus(
	ctx context.Context,
	id int64,
	server *monitor.ManualStatus,
	current types.String,
	diags *diag.Diagnostics,
) types.String {
	if server == nil {
		return current
	}

	status, ok := manualStatusToString(*server)
	if ok {
		return types.StringValue(status)
	}

	tflog.Warn(ctx, "Manual monitor reports an unrecognized manual status", map[string]any{
		"id":            id,
		"manual_status": int64(*server),
	})

	diags.AddAttributeWarning(
		path.Root("manual_status"),
		"Unrecognized manual status",
		fmt.Sprintf(
			"Monitor %d reports manual status %d, which the provider does not know. Keeping %s "+
				"from state; the next apply overwrites the status on the server.",
			id, int64(*server), current.String(),
		),
	)

	return current
}

// buildManualMonitor assembles the client monitor from the resource model.
//
// Uptime Kuma overwrites the stored status on every edit, and a nil status
// resets the monitor to pending. The planned status is therefore always sent,
// also when only an unrelated attribute changed.
func buildManualMonitor(
	ctx context.Context,
	data *MonitorManualResourceModel,
	diags *diag.Diagnostics,
) *monitor.Manual {
	manualMonitor := &monitor.Manual{
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

	// An unknown status must not fall through to nil, which would silently
	// reset the monitor to pending.
	if data.ManualStatus.IsUnknown() {
		diags.AddAttributeError(
			path.Root("manual_status"),
			"Unknown manual status",
			"The manual status must be known at apply time.",
		)

		return nil
	}

	if !data.ManualStatus.IsNull() {
		status, ok := manualStatusFromString(data.ManualStatus.ValueString())
		if !ok {
			diags.AddAttributeError(
				path.Root("manual_status"),
				"Invalid manual status",
				fmt.Sprintf(
					"Unsupported manual status %q. Expected one of %q, %q or %q.",
					data.ManualStatus.ValueString(), manualStatusUp, manualStatusDown, manualStatusPending,
				),
			)

			return nil
		}

		manualMonitor.ManualStatus = &status
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		manualMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		manualMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64

		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if diags.HasError() {
			return nil
		}

		manualMonitor.NotificationIDs = notificationIDs
	}

	return manualMonitor
}

// Create creates a new manual monitor resource.
func (r *MonitorManualResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorManualResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	manualMonitor := buildManualMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, manualMonitor)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create manual monitor") {
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

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the manual monitor resource.
func (r *MonitorManualResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorManualResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var manualMonitor monitor.Manual

	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &manualMonitor)
	if err != nil {
		removeMonitorOnServerMiss(ctx, r.client, err, data.ID.ValueInt64(), "manual monitor", resp)

		return
	}

	if monitorTypeDrifted(
		ctx, data.ID.ValueInt64(), manualMonitor.Type(), manualMonitor.Base.Type(), resp,
	) {
		return
	}

	data.Name = types.StringValue(manualMonitor.Name)
	if manualMonitor.Description != nil {
		data.Description = types.StringValue(*manualMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(manualMonitor.Interval)
	data.RetryInterval = types.Int64Value(manualMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(manualMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(manualMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(manualMonitor.UpsideDown)
	data.Active = types.BoolValue(manualMonitor.IsActive)

	data.ManualStatus = readManualStatus(
		ctx, data.ID.ValueInt64(), manualMonitor.ManualStatus, data.ManualStatus, &resp.Diagnostics,
	)

	if manualMonitor.Parent != nil {
		data.Parent = types.Int64Value(*manualMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(manualMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, manualMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, manualMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the manual monitor resource.
func (r *MonitorManualResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorManualResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorManualResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	manualMonitor := buildManualMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateMonitor(ctx, manualMonitor)
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update manual monitor") {
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

// Delete deletes the manual monitor resource.
func (r *MonitorManualResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorManualResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete manual monitor")
}

// ImportState imports an existing resource by ID.
func (*MonitorManualResource) ImportState(
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
