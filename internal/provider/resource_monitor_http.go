package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	// Ensure MonitorHTTPResource satisfies various resource interfaces.
	_ resource.Resource                = &MonitorHTTPResource{}
	_ resource.ResourceWithImportState = &MonitorHTTPResource{}
)

// NewMonitorHTTPResource returns a new instance of the HTTP monitor resource.
func NewMonitorHTTPResource() resource.Resource {
	return &MonitorHTTPResource{}
}

// MonitorHTTPResource defines the resource implementation.
type MonitorHTTPResource struct {
	client *kuma.Client
}

// MonitorHTTPResourceModel describes the resource data model for HTTP monitors.
type MonitorHTTPResourceModel struct {
	MonitorBaseModel
	MonitorHTTPBaseModel
}

// Metadata returns the metadata for the resource.
func (*MonitorHTTPResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http"
}

// Schema returns the schema for the resource.
func (*MonitorHTTPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Define resource schema attributes and validation.
	resp.Schema = schema.Schema{
		MarkdownDescription: "HTTP monitor resource",
		Attributes:          withMonitorBaseAttributes(withHTTPMonitorBaseAttributes(map[string]schema.Attribute{})),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorHTTPResource) Configure(
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

// Create creates a new resource.
func (r *MonitorHTTPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorHTTPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpMonitor := buildHTTPMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &httpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create HTTP monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildHTTPMonitor(ctx context.Context, data *MonitorHTTPResourceModel, diags *diag.Diagnostics) monitor.HTTP {
	httpMonitor := monitor.HTTP{
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
			CacheBust:           data.CacheBust.ValueBool(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		httpMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		httpMonitor.Parent = &parent
	}

	if !data.ProxyID.IsNull() {
		proxyID := data.ProxyID.ValueInt64()
		httpMonitor.ProxyID = &proxyID
	}

	if !data.AcceptedStatusCodes.IsNull() && !data.AcceptedStatusCodes.IsUnknown() {
		var statusCodes []string
		diags.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if !diags.HasError() {
			httpMonitor.AcceptedStatusCodes = statusCodes
		}
	} else {
		httpMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			httpMonitor.NotificationIDs = notificationIDs
		}
	}

	return httpMonitor
}

// stringOrNull returns a Terraform String type that is null if the input string is empty, otherwise returns the string value.
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}

	return types.StringValue(s)
}

// stringOrNullPreserveEmpty returns a Terraform String that preserves an
// explicit empty string set in the configuration. When the API returns an
// empty string and the current state already holds a non-null value (the
// user wrote e.g. `description = ""`), the empty string is kept. When the
// state is null (the user omitted the attribute), null is preserved.
func stringOrNullPreserveEmpty(apiValue string, stateValue types.String) types.String {
	if apiValue == "" && stateValue.IsNull() {
		return types.StringNull()
	}

	return types.StringValue(apiValue)
}

// populateHTTPMonitorBaseFieldsForHTTP populates the base HTTP monitor fields from the API response.
// Extracts HTTP-specific fields from the API response into the model.
// Handles base monitor fields and all HTTP configuration options.
func populateHTTPMonitorBaseFieldsForHTTP(httpMonitor *monitor.HTTP, m *MonitorHTTPResourceModel) {
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
	m.Body = stringOrNull(httpMonitor.Body)
	m.Headers = stringOrNull(httpMonitor.Headers)
	m.AuthMethod = types.StringValue(string(httpMonitor.AuthMethod))
	m.BasicAuthUser = stringOrNull(httpMonitor.BasicAuthUser)
	m.BasicAuthPass = stringOrNull(httpMonitor.BasicAuthPass)
	m.AuthDomain = stringOrNull(httpMonitor.AuthDomain)
	m.AuthWorkstation = stringOrNull(httpMonitor.AuthWorkstation)
	m.TLSCert = stringOrNull(httpMonitor.TLSCert)
	m.TLSKey = stringOrNull(httpMonitor.TLSKey)
	m.TLSCa = stringOrNull(httpMonitor.TLSCa)
	m.OAuthAuthMethod = stringOrNull(httpMonitor.OAuthAuthMethod)
	m.OAuthTokenURL = stringOrNull(httpMonitor.OAuthTokenURL)
	m.OAuthClientID = stringOrNull(httpMonitor.OAuthClientID)
	m.OAuthClientSecret = stringOrNull(httpMonitor.OAuthClientSecret)
	m.OAuthScopes = stringOrNull(httpMonitor.OAuthScopes)
	m.CacheBust = types.BoolValue(httpMonitor.CacheBust)
}

// populateOptionalFieldsForHTTP populates optional fields for HTTP monitor.
// Handles proxy, parent group, accepted status codes, and notification IDs.
// Converts null API values to Terraform null types appropriately.
func populateOptionalFieldsForHTTP(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	m *MonitorHTTPResourceModel,
	diags *diag.Diagnostics,
) {
	// Set parent monitor group if present.
	if httpMonitor.Parent != nil {
		m.Parent = types.Int64Value(*httpMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	// Set proxy if configured.
	if httpMonitor.ProxyID != nil {
		m.ProxyID = types.Int64Value(*httpMonitor.ProxyID)
	} else {
		m.ProxyID = types.Int64Null()
	}

	// Convert accepted status codes list if non-empty.
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
func (r *MonitorHTTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorHTTPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var httpMonitor monitor.HTTP
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &httpMonitor)
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read HTTP monitor", err.Error())
		return
	}

	populateHTTPMonitorBaseFieldsForHTTP(&httpMonitor, &data)
	populateOptionalFieldsForHTTP(ctx, &httpMonitor, &data, &resp.Diagnostics)

	data.Tags = handleMonitorTagsRead(ctx, httpMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MonitorHTTPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorHTTPResourceModel
	var state MonitorHTTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpMonitor := buildHTTPMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &httpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update HTTP monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *MonitorHTTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorHTTPResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API.
	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete HTTP monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorHTTPResource) ImportState(
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
