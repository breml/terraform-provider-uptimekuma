package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	// Ensure MonitorWebsocketUpgradeResource satisfies various resource interfaces.
	_ resource.Resource                = &MonitorWebsocketUpgradeResource{}
	_ resource.ResourceWithImportState = &MonitorWebsocketUpgradeResource{}
)

// NewMonitorWebsocketUpgradeResource returns a new instance of the Websocket Upgrade monitor resource.
func NewMonitorWebsocketUpgradeResource() resource.Resource {
	return &MonitorWebsocketUpgradeResource{}
}

// MonitorWebsocketUpgradeResource defines the resource implementation.
type MonitorWebsocketUpgradeResource struct {
	client *kuma.Client
}

// MonitorWebsocketUpgradeResourceModel describes the resource data model for Websocket Upgrade monitors.
type MonitorWebsocketUpgradeResourceModel struct {
	MonitorBaseModel
	MonitorHTTPBaseModel

	WSIgnoreSecWebsocketAcceptHeader types.Bool   `tfsdk:"ws_ignore_sec_websocket_accept_header"`
	WSSubprotocol                    types.String `tfsdk:"ws_subprotocol"`
}

// Metadata returns the metadata for the resource.
func (*MonitorWebsocketUpgradeResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_websocket_upgrade"
}

// Schema returns the schema for the resource.
func (*MonitorWebsocketUpgradeResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	// Define resource schema attributes and validation.
	resp.Schema = schema.Schema{
		MarkdownDescription: "Websocket Upgrade monitor resource checks that a WebSocket handshake " +
			"(`ws://` or `wss://`) completes successfully. The monitor performs the HTTP upgrade " +
			"request and verifies the server responds with a switching protocols response.",
		Attributes: withMonitorBaseAttributes(withHTTPMonitorBaseAttributes(map[string]schema.Attribute{
			"ws_ignore_sec_websocket_accept_header": schema.BoolAttribute{
				MarkdownDescription: "Skip verification of the `Sec-WebSocket-Accept` response header " +
					"during the WebSocket handshake.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"ws_subprotocol": schema.StringAttribute{
				MarkdownDescription: "Requested `Sec-WebSocket-Protocol` value to send during the " +
					"WebSocket handshake.",
				Optional: true,
			},
		})),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorWebsocketUpgradeResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new resource.
func (r *MonitorWebsocketUpgradeResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorWebsocketUpgradeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	websocketUpgradeMonitor := buildWebsocketUpgradeMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &websocketUpgradeMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Websocket Upgrade monitor", err.Error())
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

// buildWebsocketUpgradeMonitor constructs a monitor.WebsocketUpgrade from the Terraform resource model.
func buildWebsocketUpgradeMonitor(
	ctx context.Context,
	data *MonitorWebsocketUpgradeResourceModel,
	diags *diag.Diagnostics,
) monitor.WebsocketUpgrade {
	websocketUpgradeMonitor := monitor.WebsocketUpgrade{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		HTTPDetails: monitor.HTTPDetails{
			URL:                 data.URL.ValueString(),
			Timeout:             data.Timeout.ValueInt64(),
			Method:              data.Method.ValueString(),
			ExpiryNotification:  data.ExpiryNotification.ValueBool(),
			IgnoreTLS:           data.IgnoreTLS.ValueBool(),
			MaxRedirects:        int(data.MaxRedirects.ValueInt64()),
			AcceptedStatusCodes: []string{},
			HTTPBodyEncoding:    data.HTTPBodyEncoding.ValueString(),
			Body:                data.Body.ValueString(),
			Headers:             data.Headers.ValueString(),
			AuthMethod:          monitor.AuthMethod(data.AuthMethod.ValueString()),
			BasicAuthUser:       data.BasicAuthUser.ValueString(),
			BasicAuthPass:       data.BasicAuthPass.ValueString(),
			AuthDomain:          data.AuthDomain.ValueString(),
			AuthWorkstation:     data.AuthWorkstation.ValueString(),
			TLSCert:             data.TLSCert.ValueString(),
			TLSKey:              data.TLSKey.ValueString(),
			TLSCa:               data.TLSCa.ValueString(),
			OAuthAuthMethod:     data.OAuthAuthMethod.ValueString(),
			OAuthTokenURL:       data.OAuthTokenURL.ValueString(),
			OAuthClientID:       data.OAuthClientID.ValueString(),
			OAuthClientSecret:   data.OAuthClientSecret.ValueString(),
			OAuthScopes:         data.OAuthScopes.ValueString(),
			OAuthAudience:       data.OAuthAudience.ValueString(),
			CacheBust:           data.CacheBust.ValueBool(),
		},
		WebsocketUpgradeDetails: monitor.WebsocketUpgradeDetails{
			IgnoreSecWebsocketAcceptHeader: data.WSIgnoreSecWebsocketAcceptHeader.ValueBool(),
			Subprotocol:                    data.WSSubprotocol.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		websocketUpgradeMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		websocketUpgradeMonitor.Parent = &parent
	}

	if !data.ProxyID.IsNull() {
		proxyID := data.ProxyID.ValueInt64()
		websocketUpgradeMonitor.ProxyID = &proxyID
	}

	if !data.AcceptedStatusCodes.IsNull() && !data.AcceptedStatusCodes.IsUnknown() {
		var statusCodes []string
		diags.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if !diags.HasError() {
			websocketUpgradeMonitor.AcceptedStatusCodes = statusCodes
		}
	} else {
		websocketUpgradeMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			websocketUpgradeMonitor.NotificationIDs = notificationIDs
		}
	}

	return websocketUpgradeMonitor
}

// stringOrNullWebsocketUpgrade returns a Terraform String type that is null if the input string is
// empty, otherwise returns the string value.
func stringOrNullWebsocketUpgrade(s string) types.String {
	if s == "" {
		return types.StringNull()
	}

	return types.StringValue(s)
}

// populateHTTPBaseFieldsForWebsocketUpgrade populates base fields for Websocket Upgrade monitor.
// Extracts common HTTP fields from API response for the Websocket Upgrade specific model.
func populateHTTPBaseFieldsForWebsocketUpgrade(httpMonitor *monitor.HTTP, m *MonitorWebsocketUpgradeResourceModel) {
	m.Name = types.StringValue(httpMonitor.Name)
	if httpMonitor.Description != nil {
		m.Description = types.StringValue(*httpMonitor.Description)
	} else {
		m.Description = types.StringNull()
	}

	m.Interval = types.Int64Value(httpMonitor.Interval)
	m.RetryInterval = types.Int64Value(httpMonitor.RetryInterval)
	m.ResendInterval = types.Int64Value(httpMonitor.ResendInterval)
	m.MaxRetries = types.Int64Value(httpMonitor.MaxRetries)
	m.UpsideDown = types.BoolValue(httpMonitor.UpsideDown)
	m.Active = types.BoolValue(httpMonitor.IsActive)
	m.URL = types.StringValue(httpMonitor.URL)
	m.Timeout = types.Int64Value(httpMonitor.Timeout)
	m.Method = types.StringValue(httpMonitor.Method)
	m.ExpiryNotification = types.BoolValue(httpMonitor.ExpiryNotification)
	m.IgnoreTLS = types.BoolValue(httpMonitor.IgnoreTLS)
	m.MaxRedirects = types.Int64Value(int64(httpMonitor.MaxRedirects))
	m.HTTPBodyEncoding = types.StringValue(httpMonitor.HTTPBodyEncoding)
	m.Body = stringOrNullWebsocketUpgrade(httpMonitor.Body)
	m.Headers = stringOrNullWebsocketUpgrade(httpMonitor.Headers)
	m.AuthMethod = types.StringValue(string(httpMonitor.AuthMethod))
	m.BasicAuthUser = stringOrNullWebsocketUpgrade(httpMonitor.BasicAuthUser)
	m.BasicAuthPass = stringOrNullWebsocketUpgrade(httpMonitor.BasicAuthPass)
	m.AuthDomain = stringOrNullWebsocketUpgrade(httpMonitor.AuthDomain)
	m.AuthWorkstation = stringOrNullWebsocketUpgrade(httpMonitor.AuthWorkstation)
	m.TLSCert = stringOrNullWebsocketUpgrade(httpMonitor.TLSCert)
	m.TLSKey = stringOrNullWebsocketUpgrade(httpMonitor.TLSKey)
	m.TLSCa = stringOrNullWebsocketUpgrade(httpMonitor.TLSCa)
	m.OAuthAuthMethod = stringOrNullWebsocketUpgrade(httpMonitor.OAuthAuthMethod)
	m.OAuthTokenURL = stringOrNullWebsocketUpgrade(httpMonitor.OAuthTokenURL)
	m.OAuthClientID = stringOrNullWebsocketUpgrade(httpMonitor.OAuthClientID)
	m.OAuthClientSecret = stringOrNullWebsocketUpgrade(httpMonitor.OAuthClientSecret)
	m.OAuthScopes = stringOrNullWebsocketUpgrade(httpMonitor.OAuthScopes)
	m.OAuthAudience = stringOrNullWebsocketUpgrade(httpMonitor.OAuthAudience)
	m.CacheBust = types.BoolValue(httpMonitor.CacheBust)
}

// populateOptionalFieldsForWebsocketUpgrade populates optional fields for Websocket Upgrade monitor.
// Handles parent group, proxy, status codes, and notification configuration.
// Parent and proxy fields are optional and may be null in the API response.
// Lists are properly converted to Terraform types for accurate state representation.
func populateOptionalFieldsForWebsocketUpgrade(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	m *MonitorWebsocketUpgradeResourceModel,
	diags *diag.Diagnostics,
) {
	// Set parent monitor group if configured in API response.
	if httpMonitor.Parent != nil {
		m.Parent = types.Int64Value(*httpMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	// Set proxy ID if the monitor uses a proxy.
	if httpMonitor.ProxyID != nil {
		m.ProxyID = types.Int64Value(*httpMonitor.ProxyID)
	} else {
		m.ProxyID = types.Int64Null()
	}

	// Convert accepted status codes list if present.
	if len(httpMonitor.AcceptedStatusCodes) > 0 {
		statusCodes, d := types.ListValueFrom(ctx, types.StringType, httpMonitor.AcceptedStatusCodes)
		diags.Append(d...)
		m.AcceptedStatusCodes = statusCodes
	}

	// Convert notification IDs list if present.
	if len(httpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, httpMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// Read reads the current state of the resource.
func (r *MonitorWebsocketUpgradeResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorWebsocketUpgradeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var websocketUpgradeMonitor monitor.WebsocketUpgrade
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &websocketUpgradeMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read Websocket Upgrade monitor", err.Error())
		return
	}

	if actual := websocketUpgradeMonitor.Base.Type(); actual != "" && actual != websocketUpgradeMonitor.Type() {
		tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
			"id":            data.ID.ValueInt64(),
			"expected_type": websocketUpgradeMonitor.Type(),
			"actual_type":   actual,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	var httpMonitor monitor.HTTP
	httpMonitor.Base = websocketUpgradeMonitor.Base
	httpMonitor.HTTPDetails = websocketUpgradeMonitor.HTTPDetails
	populateHTTPBaseFieldsForWebsocketUpgrade(&httpMonitor, &data)
	populateOptionalFieldsForWebsocketUpgrade(ctx, &httpMonitor, &data, &resp.Diagnostics)

	data.WSIgnoreSecWebsocketAcceptHeader = types.BoolValue(websocketUpgradeMonitor.IgnoreSecWebsocketAcceptHeader)
	data.WSSubprotocol = stringOrNullWebsocketUpgrade(websocketUpgradeMonitor.Subprotocol)

	data.Tags = handleMonitorTagsRead(ctx, websocketUpgradeMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MonitorWebsocketUpgradeResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorWebsocketUpgradeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorWebsocketUpgradeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	websocketUpgradeMonitor := buildWebsocketUpgradeMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	websocketUpgradeMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &websocketUpgradeMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update Websocket Upgrade monitor", err.Error())
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

// Delete deletes the resource.
func (r *MonitorWebsocketUpgradeResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorWebsocketUpgradeResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API.
	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Websocket Upgrade monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorWebsocketUpgradeResource) ImportState(
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
