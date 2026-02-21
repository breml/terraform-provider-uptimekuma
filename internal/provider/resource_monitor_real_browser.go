package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

	URL                 types.String `tfsdk:"url"`
	Timeout             types.Int64  `tfsdk:"timeout"`
	IgnoreTLS           types.Bool   `tfsdk:"ignore_tls"`
	MaxRedirects        types.Int64  `tfsdk:"max_redirects"`
	AcceptedStatusCodes types.List   `tfsdk:"accepted_status_codes"`
	ProxyID             types.Int64  `tfsdk:"proxy_id"`
	RemoteBrowser       types.Int64  `tfsdk:"remote_browser"`
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

	attrs["timeout"] = schema.Int64Attribute{
		MarkdownDescription: "Request timeout in seconds",
		Optional:            true,
		Computed:            true,
		Default:             int64default.StaticInt64(48),
		Validators: []validator.Int64{
			int64validator.Between(1, 3600),
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

	return attrs
}

// Configure configures the Real Browser monitor resource with the API client.
func (r *MonitorRealBrowserResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	r.client = client
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
			URL:                 data.URL.ValueString(),
			Timeout:             data.Timeout.ValueInt64(),
			IgnoreTLS:           data.IgnoreTLS.ValueBool(),
			MaxRedirects:        int(data.MaxRedirects.ValueInt64()),
			AcceptedStatusCodes: []string{},
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

	if !data.AcceptedStatusCodes.IsNull() && !data.AcceptedStatusCodes.IsUnknown() {
		var statusCodes []string
		resp.Diagnostics.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		realBrowserMonitor.AcceptedStatusCodes = statusCodes
	} else {
		realBrowserMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		realBrowserMonitor.NotificationIDs = notificationIDs
	}

	// Create monitor via API.
	id, err := r.client.CreateMonitor(ctx, &realBrowserMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create Real Browser monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
	data.Timeout = types.Int64Value(m.Timeout)
	data.IgnoreTLS = types.BoolValue(m.IgnoreTLS)
	data.MaxRedirects = types.Int64Value(int64(m.MaxRedirects))
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
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read Real Browser monitor", err.Error())
		return
	}

	populateRealBrowserMonitorBaseFields(&realBrowserMonitor, &data)
	populateOptionalFieldsForRealBrowser(ctx, &realBrowserMonitor, &data, &resp.Diagnostics)

	data.Tags = handleMonitorTagsRead(ctx, realBrowserMonitor.Tags, &resp.Diagnostics)
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

	realBrowserMonitor := monitor.RealBrowser{
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
		RealBrowserDetails: monitor.RealBrowserDetails{
			URL:                 data.URL.ValueString(),
			Timeout:             data.Timeout.ValueInt64(),
			IgnoreTLS:           data.IgnoreTLS.ValueBool(),
			MaxRedirects:        int(data.MaxRedirects.ValueInt64()),
			AcceptedStatusCodes: []string{},
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

	if !data.AcceptedStatusCodes.IsNull() && !data.AcceptedStatusCodes.IsUnknown() {
		var statusCodes []string
		resp.Diagnostics.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		realBrowserMonitor.AcceptedStatusCodes = statusCodes
	} else {
		realBrowserMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		realBrowserMonitor.NotificationIDs = notificationIDs
	}

	// Update monitor via API.
	err := r.client.UpdateMonitor(ctx, &realBrowserMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update Real Browser monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
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
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Real Browser monitor", err.Error())
		return
	}
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
