package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
	_ resource.Resource                = &MonitorPM2Resource{}
	_ resource.ResourceWithImportState = &MonitorPM2Resource{}
)

// NewMonitorPM2Resource returns a new instance of the PM2 monitor resource.
func NewMonitorPM2Resource() resource.Resource {
	return &MonitorPM2Resource{}
}

// MonitorPM2Resource defines the resource implementation for PM2 monitors.
type MonitorPM2Resource struct {
	client *kuma.Client
}

// MonitorPM2ResourceModel describes the resource data model for PM2 monitors.
type MonitorPM2ResourceModel struct {
	MonitorBaseModel

	ProcessName types.String `tfsdk:"process_name"` // PM2 process name or numeric PM2 id.
}

// Metadata returns the metadata for the resource.
func (*MonitorPM2Resource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_pm2"
}

// Schema returns the schema for the resource.
func (*MonitorPM2Resource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PM2 monitor resource. The check runs `pm2 jlist` on the Uptime Kuma " +
			"host and goes up while the named PM2 process reports status `online`.\n\n" +
			"The `pm2` CLI must be installed on the Uptime Kuma host and must see the PM2 daemon " +
			"that owns the process. The official `louislam/uptime-kuma` container image does not " +
			"ship it, so this monitor stays down there even though it can be created and managed.",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"process_name": schema.StringAttribute{
				MarkdownDescription: "PM2 process to check. The server matches the value against both " +
					"the process name and the stringified numeric PM2 id reported by `pm2 jlist`, so " +
					"either form works. Prefer the name: PM2 reassigns ids after a process is deleted " +
					"and recreated.\n\n" +
					"Spaces inside the name are allowed, unlike for the system-service monitor this " +
					"type shares its wire field with. Leading and trailing spaces are ignored, " +
					"because Uptime Kuma trims the value before storing it. ASCII control " +
					"characters (U+0000..U+001F and U+007F, which includes tabs and newlines) are " +
					"rejected: Uptime Kuma refuses them in the value it stores.",
				Required: true,
				Validators: []validator.String{
					nonBlank(),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[^\x00-\x1f\x7f]*$`),
						"must not contain ASCII control characters",
					),
				},
			},
		}),
	}
}

// Configure configures the PM2 monitor resource with the API client.
func (r *MonitorPM2Resource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// buildPM2Monitor assembles the client monitor from the resource model.
func buildPM2Monitor(
	ctx context.Context,
	data *MonitorPM2ResourceModel,
	diags *diag.Diagnostics,
) *monitor.PM2 {
	pm2Monitor := &monitor.PM2{
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
		PM2Details: monitor.PM2Details{
			// Uptime Kuma trims the process name before storing it, so send the
			// trimmed value to keep state and server in agreement.
			ProcessName: strings.TrimSpace(data.ProcessName.ValueString()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		pm2Monitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		pm2Monitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64

		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if diags.HasError() {
			return nil
		}

		pm2Monitor.NotificationIDs = notificationIDs
	}

	return pm2Monitor
}

// Create creates a new PM2 monitor resource.
func (r *MonitorPM2Resource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorPM2ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pm2Monitor := buildPM2Monitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, pm2Monitor)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create PM2 monitor") {
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

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the PM2 monitor resource.
func (r *MonitorPM2Resource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorPM2ResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var pm2Monitor monitor.PM2

	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &pm2Monitor)
	// Handle error.
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read PM2 monitor", err.Error())

		return
	}

	if actual := pm2Monitor.Base.Type(); actual != "" && actual != pm2Monitor.Type() {
		tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
			"id":            data.ID.ValueInt64(),
			"expected_type": pm2Monitor.Type(),
			"actual_type":   actual,
		})
		resp.State.RemoveResource(ctx)

		return
	}

	data.Name = types.StringValue(pm2Monitor.Name)
	if pm2Monitor.Description != nil {
		data.Description = types.StringValue(*pm2Monitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(pm2Monitor.Interval)
	data.RetryInterval = types.Int64Value(pm2Monitor.RetryInterval)
	data.ResendInterval = types.Int64Value(pm2Monitor.ResendInterval)
	data.MaxRetries = types.Int64Value(pm2Monitor.MaxRetries)
	data.UpsideDown = types.BoolValue(pm2Monitor.UpsideDown)
	data.Active = types.BoolValue(pm2Monitor.IsActive)

	// Both Terraform and Uptime Kuma see the trimmed process name, so padding in
	// the configuration is not drift. Keep the configured value in that case,
	// otherwise every plan would want to write the padding back.
	if strings.TrimSpace(data.ProcessName.ValueString()) != pm2Monitor.ProcessName {
		data.ProcessName = types.StringValue(pm2Monitor.ProcessName)
	}

	if pm2Monitor.Parent != nil {
		data.Parent = types.Int64Value(*pm2Monitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(pm2Monitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, pm2Monitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, pm2Monitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the PM2 monitor resource.
func (r *MonitorPM2Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorPM2ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorPM2ResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pm2Monitor := buildPM2Monitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateMonitor(ctx, pm2Monitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update PM2 monitor", err.Error())
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

// Delete deletes the PM2 monitor resource.
func (r *MonitorPM2Resource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorPM2ResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete PM2 monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorPM2Resource) ImportState(
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
