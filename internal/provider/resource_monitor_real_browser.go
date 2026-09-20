package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorRealBrowserResource{}
	_ resource.ResourceWithImportState = &MonitorRealBrowserResource{}
)

// NewMonitorRealBrowserResource returns a new instance of the Real Browser monitor resource.
func NewMonitorRealBrowserResource() resource.Resource {
	return &MonitorRealBrowserResource{}
}

// MonitorRealBrowserResource defines the resource implementation.
type MonitorRealBrowserResource struct {
	client *kuma.Client
}

// MonitorRealBrowserResourceModel describes the resource data model.
type MonitorRealBrowserResourceModel struct {
	MonitorBaseModel

	URL                      types.String  `tfsdk:"url"`
	Timeout                  types.Float64 `tfsdk:"timeout"`
	IgnoreTLS                types.Bool    `tfsdk:"ignore_tls"`
	MaxRedirects             types.Int64   `tfsdk:"max_redirects"`
	AcceptedStatusCodes      types.List    `tfsdk:"accepted_status_codes"`
	ProxyID                  types.Int64   `tfsdk:"proxy_id"`
	RemoteBrowser            types.Int64   `tfsdk:"remote_browser"`
	ScreenshotDelay          types.Int64   `tfsdk:"screenshot_delay"`
	DomainExpiryNotification types.Bool    `tfsdk:"domain_expiry_notification"`
}

// Metadata returns the metadata for the resource.
func (*MonitorRealBrowserResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_real_browser"
}

// Schema returns the schema for the resource.
func (*MonitorRealBrowserResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	// Define resource schema attributes and validation.
	resp.Schema = schema.Schema{
		MarkdownDescription: "Real Browser monitor resource",
		Attributes:          withMonitorBaseAttributes(withRealBrowserMonitorAttributes(map[string]schema.Attribute{})),
	}
}

func withRealBrowserMonitorAttributes(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	attrs["url"] = schema.StringAttribute{
		MarkdownDescription: "URL to monitor",
		Required:            true,
	}

	attrs["timeout"] = schema.Float64Attribute{
		MarkdownDescription: "Request timeout in seconds. Has no effect: the real browser check derives " +
			"its timeout from 80% of `interval` and never reads this value. It is kept because Uptime " +
			"Kuma stores it with every monitor.",
		DeprecationMessage: "timeout has no effect on a real browser monitor; the check uses 80% of " +
			"interval. Remove it from the configuration.",
		Optional: true,
		Computed: true,
		Default:  float64default.StaticFloat64(48),
		Validators: []validator.Float64{
			float64validator.Between(1, 3600),
		},
	}

	attrs["ignore_tls"] = schema.BoolAttribute{
		MarkdownDescription: "Ignore TLS/SSL errors",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	}

	attrs["max_redirects"] = schema.Int64Attribute{
		MarkdownDescription: "Maximum number of redirects to follow",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(10),
		Validators: []validator.Int64{
			int64validator.Between(0, 20),
		},
	}

	attrs["accepted_status_codes"] = schema.ListAttribute{
		MarkdownDescription: "Accepted HTTP status codes (e.g., ['200-299', '301'])",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Default: listdefault.StaticValue(
			types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200-299")}),
		),
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
	}

	attrs["proxy_id"] = schema.Int64Attribute{
		MarkdownDescription: "Proxy ID",
		Optional:            true,
	}

	attrs["remote_browser"] = schema.Int64Attribute{
		MarkdownDescription: "Remote Browser ID (if using a remote browser for monitoring)",
		Optional:            true,
	}

	attrs["screenshot_delay"] = schema.Int64Attribute{
		MarkdownDescription: "Delay in milliseconds before taking a screenshot. Must be less than " +
			"`interval * 500`, that is half the interval converted to milliseconds; the server rejects " +
			"larger values and negative ones. The delay cannot be cleared through the API, so this " +
			"attribute is computed: removing it from configuration keeps the value the server already " +
			"has rather than producing a plan that never converges. Set it to `0` to disable the delay. " +
			"Uptime Kuma only returns the value since 2.5.0; against earlier versions it cannot be " +
			"detected as drift or recovered on import.",
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
		Validators: []validator.Int64{
			int64validator.AtLeast(0),
		},
	}

	attrs["domain_expiry_notification"] = domainExpiryNotificationAttribute()

	return attrs
}

// Configure configures the Real Browser monitor resource with the API client.
func (r *MonitorRealBrowserResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// buildRealBrowserMonitor constructs a Real Browser monitor from the resource model.
func buildRealBrowserMonitor(
	ctx context.Context,
	data *MonitorRealBrowserResourceModel,
	diags *diag.Diagnostics,
) monitor.RealBrowser {
	realBrowserMonitor := monitor.RealBrowser{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		RealBrowserDetails: monitor.RealBrowserDetails{
			URL:                      data.URL.ValueString(),
			Timeout:                  data.Timeout.ValueFloat64(),
			IgnoreTLS:                data.IgnoreTLS.ValueBool(),
			MaxRedirects:             int(data.MaxRedirects.ValueInt64()),
			AcceptedStatusCodes:      []string{},
			DomainExpiryNotification: data.DomainExpiryNotification.ValueBool(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		realBrowserMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		realBrowserMonitor.Parent = &parent
	}

	if !data.ProxyID.IsNull() {
		proxyID := data.ProxyID.ValueInt64()
		realBrowserMonitor.ProxyID = &proxyID
	}

	if !data.RemoteBrowser.IsNull() {
		remoteBrowser := data.RemoteBrowser.ValueInt64()
		realBrowserMonitor.RemoteBrowser = &remoteBrowser
	}

	if !data.ScreenshotDelay.IsNull() && !data.ScreenshotDelay.IsUnknown() {
		screenshotDelay := int(data.ScreenshotDelay.ValueInt64())
		realBrowserMonitor.ScreenshotDelay = &screenshotDelay
	}

	if !data.AcceptedStatusCodes.IsNull() && !data.AcceptedStatusCodes.IsUnknown() {
		var statusCodes []string
		diags.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if diags.HasError() {
			return realBrowserMonitor
		}

		realBrowserMonitor.AcceptedStatusCodes = statusCodes
	} else {
		realBrowserMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if diags.HasError() {
			return realBrowserMonitor
		}

		realBrowserMonitor.NotificationIDs = notificationIDs
	}

	return realBrowserMonitor
}

// Create creates a new Real Browser monitor resource.
func (r *MonitorRealBrowserResource) Create(
	// Extract and validate configuration.
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorRealBrowserResourceModel

	// Extract plan data.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	realBrowserMonitor := buildRealBrowserMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create monitor via API.
	id, err := r.client.CreateMonitor(ctx, &realBrowserMonitor)
	// Handle error.
	if err != nil && !createdWithoutEvent(&resp.Diagnostics, err, id, "failed to create Real Browser monitor") {
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// The monitor exists, so record it rather than leaving it unmanaged.
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

		return
	}

	resolveScreenshotDelay(ctx, r.client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
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

// resolveScreenshotDelay gives the computed screenshot_delay a known value.
//
// The attribute is computed because Uptime Kuma cannot clear the delay, so an
// absent configuration must keep whatever the server holds. On create, and on
// update from a state written before the attribute became computed, there is
// no prior value to carry over, so the server is asked for the current one.
// Servers older than 2.5.0 do not echo it back; they report no delay.
func resolveScreenshotDelay(
	ctx context.Context,
	client *kuma.Client,
	data *MonitorRealBrowserResourceModel,
	diags *diag.Diagnostics,
) {
	if !data.ScreenshotDelay.IsUnknown() {
		return
	}

	var current monitor.RealBrowser

	err := client.GetMonitorAs(ctx, data.ID.ValueInt64(), &current)
	if err != nil {
		diags.AddError("failed to read back Real Browser monitor screenshot delay", err.Error())
		return
	}

	if current.ScreenshotDelay != nil {
		data.ScreenshotDelay = types.Int64Value(int64(*current.ScreenshotDelay))
		return
	}

	data.ScreenshotDelay = types.Int64Value(0)
}

// populateRealBrowserMonitorBaseFields populates base fields for Real Browser monitor.
func populateRealBrowserMonitorBaseFields(m *monitor.RealBrowser, data *MonitorRealBrowserResourceModel) {
	data.Name = types.StringValue(m.Name)
	if m.Description != nil {
		data.Description = types.StringValue(*m.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(m.Interval)
	data.RetryInterval = types.Int64Value(m.RetryInterval)
	data.ResendInterval = types.Int64Value(m.ResendInterval)
	data.MaxRetries = types.Int64Value(m.MaxRetries)
	data.UpsideDown = types.BoolValue(m.UpsideDown)
	data.Active = types.BoolValue(m.IsActive)
	data.URL = types.StringValue(m.URL)
	data.Timeout = types.Float64Value(m.Timeout)
	data.IgnoreTLS = types.BoolValue(m.IgnoreTLS)
	data.MaxRedirects = types.Int64Value(int64(m.MaxRedirects))
	data.DomainExpiryNotification = types.BoolValue(m.DomainExpiryNotification)
}

// populateOptionalFieldsForRealBrowser populates optional fields for Real Browser monitor.
// Handles parent group, proxy, remote browser, accepted status codes, and notification IDs.
// Converts null API values to Terraform null types appropriately.
func populateOptionalFieldsForRealBrowser(
	ctx context.Context,
	m *monitor.RealBrowser,
	data *MonitorRealBrowserResourceModel,
	diags *diag.Diagnostics,
) {
	if m.Parent != nil {
		data.Parent = types.Int64Value(*m.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if m.ProxyID != nil {
		data.ProxyID = types.Int64Value(*m.ProxyID)
	} else {
		data.ProxyID = types.Int64Null()
	}

	if m.RemoteBrowser != nil {
		data.RemoteBrowser = types.Int64Value(*m.RemoteBrowser)
	} else {
		data.RemoteBrowser = types.Int64Null()
	}

	// screenshot_delay is computed, so state always mirrors the server once it
	// is known. Uptime Kuma only echoes the value back since 2.5.0; older
	// servers omit the field, so the value already in state is preserved.
	if m.ScreenshotDelay != nil {
		data.ScreenshotDelay = types.Int64Value(int64(*m.ScreenshotDelay))
	}

	if len(m.AcceptedStatusCodes) > 0 {
		statusCodes, d := types.ListValueFrom(ctx, types.StringType, m.AcceptedStatusCodes)
		diags.Append(d...)
		data.AcceptedStatusCodes = statusCodes
	}

	if len(m.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, m.NotificationIDs)
		diags.Append(d...)
		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// Read reads the current state of the Real Browser monitor resource.
func (r *MonitorRealBrowserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorRealBrowserResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var realBrowserMonitor monitor.RealBrowser
	// Fetch monitor from API.
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &realBrowserMonitor)
	// Handle error.
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read Real Browser monitor", err.Error())
		return
	}

	if actual := realBrowserMonitor.Base.Type(); actual != "" && actual != realBrowserMonitor.Type() {
		tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
			"id":            data.ID.ValueInt64(),
			"expected_type": realBrowserMonitor.Type(),
			"actual_type":   actual,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	populateRealBrowserMonitorBaseFields(&realBrowserMonitor, &data)
	populateOptionalFieldsForRealBrowser(ctx, &realBrowserMonitor, &data, &resp.Diagnostics)

	data.Tags = handleMonitorTagsRead(ctx, realBrowserMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Real Browser monitor resource.
func (r *MonitorRealBrowserResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorRealBrowserResourceModel

	// Extract plan data.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorRealBrowserResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	realBrowserMonitor := buildRealBrowserMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	realBrowserMonitor.ID = data.ID.ValueInt64()

	// Update monitor via API.
	err := r.client.UpdateMonitor(ctx, &realBrowserMonitor)
	// Handle error.
	if err != nil && !updatedWithoutEvent(&resp.Diagnostics, err, "failed to update Real Browser monitor") {
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

	resolveScreenshotDelay(ctx, r.client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Real Browser monitor resource.
func (r *MonitorRealBrowserResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorRealBrowserResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API.
	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	deletedWithoutEvent(&resp.Diagnostics, err, "failed to delete Real Browser monitor")
}

// ImportState imports an existing resource by ID.
func (*MonitorRealBrowserResource) ImportState(
	// Import monitor by ID.
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
