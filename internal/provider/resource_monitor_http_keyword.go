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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	// Ensure MonitorHTTPKeywordResource satisfies various resource interfaces.
	_ resource.Resource                = &MonitorHTTPKeywordResource{}
	_ resource.ResourceWithImportState = &MonitorHTTPKeywordResource{}
)

// NewMonitorHTTPKeywordResource returns a new instance of the HTTP Keyword monitor resource.
func NewMonitorHTTPKeywordResource() resource.Resource {
	return &MonitorHTTPKeywordResource{}
}

// MonitorHTTPKeywordResource defines the resource implementation.
type MonitorHTTPKeywordResource struct {
	client *kuma.Client
}

// MonitorHTTPKeywordResourceModel describes the resource data model for HTTP Keyword monitors.
type MonitorHTTPKeywordResourceModel struct {
	MonitorBaseModel
	MonitorHTTPBaseModel

	Keyword       types.String `tfsdk:"keyword"`
	InvertKeyword types.Bool   `tfsdk:"invert_keyword"`
}

// Metadata returns the metadata for the resource.
func (*MonitorHTTPKeywordResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http_keyword"
}

// Schema returns the schema for the resource.
func (*MonitorHTTPKeywordResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	// Define resource schema attributes and validation.
	resp.Schema = schema.Schema{
		MarkdownDescription: "HTTP Keyword monitor resource checks for the presence (or absence) of a specific keyword in the HTTP response body. The monitor makes an HTTP(S) request and searches for the specified keyword in the response. Use `invert_keyword` to reverse the logic: when false (default), finding the keyword means UP; when true, finding the keyword means DOWN.",
		Attributes: withMonitorBaseAttributes(withHTTPMonitorBaseAttributes(map[string]schema.Attribute{
			"keyword": schema.StringAttribute{
				MarkdownDescription: "Keyword to search for in the response body (case-sensitive). The monitor will search for this exact text in the HTTP response.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"invert_keyword": schema.BoolAttribute{
				MarkdownDescription: "Invert keyword match logic. When false (default), finding the keyword means UP and not finding it means DOWN. When true, finding the keyword means DOWN and not finding it means UP.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		})),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorHTTPKeywordResource) Configure(
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
func (r *MonitorHTTPKeywordResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorHTTPKeywordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpKeywordMonitor := buildHTTPKeywordMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, httpKeywordMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create HTTP Keyword monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// buildHTTPKeywordMonitor constructs a monitor.HTTPKeyword from the Terraform resource model.
func buildHTTPKeywordMonitor(
	ctx context.Context,
	data *MonitorHTTPKeywordResourceModel,
	diags *diag.Diagnostics,
) monitor.HTTPKeyword {
	httpKeywordMonitor := monitor.HTTPKeyword{
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
		HTTPKeywordDetails: monitor.HTTPKeywordDetails{
			Keyword:       data.Keyword.ValueString(),
			InvertKeyword: data.InvertKeyword.ValueBool(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		httpKeywordMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		httpKeywordMonitor.Parent = &parent
	}

	if !data.ProxyID.IsNull() {
		proxyID := data.ProxyID.ValueInt64()
		httpKeywordMonitor.ProxyID = &proxyID
	}

	if !data.AcceptedStatusCodes.IsNull() && !data.AcceptedStatusCodes.IsUnknown() {
		var statusCodes []string
		diags.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if !diags.HasError() {
			httpKeywordMonitor.AcceptedStatusCodes = statusCodes
		}
	} else {
		httpKeywordMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			httpKeywordMonitor.NotificationIDs = notificationIDs
		}
	}

	return httpKeywordMonitor
}

// stringOrNull returns a Terraform String type that is null if the input string is empty, otherwise returns the string value.
func stringOrNullKeyword(s string) types.String {
	if s == "" {
		return types.StringNull()
	}

	return types.StringValue(s)
}

// populateHTTPBaseFieldsForKeyword populates base fields for HTTP Keyword monitor.
// Extracts common HTTP fields from API response for Keyword specific model.
func populateHTTPBaseFieldsForKeyword(httpMonitor *monitor.HTTP, m *MonitorHTTPKeywordResourceModel) {
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
	m.Body = stringOrNullKeyword(httpMonitor.Body)
	m.Headers = stringOrNullKeyword(httpMonitor.Headers)
	m.AuthMethod = types.StringValue(string(httpMonitor.AuthMethod))
	m.BasicAuthUser = stringOrNullKeyword(httpMonitor.BasicAuthUser)
	m.BasicAuthPass = stringOrNullKeyword(httpMonitor.BasicAuthPass)
	m.AuthDomain = stringOrNullKeyword(httpMonitor.AuthDomain)
	m.AuthWorkstation = stringOrNullKeyword(httpMonitor.AuthWorkstation)
	m.TLSCert = stringOrNullKeyword(httpMonitor.TLSCert)
	m.TLSKey = stringOrNullKeyword(httpMonitor.TLSKey)
	m.TLSCa = stringOrNullKeyword(httpMonitor.TLSCa)
	m.OAuthAuthMethod = stringOrNullKeyword(httpMonitor.OAuthAuthMethod)
	m.OAuthTokenURL = stringOrNullKeyword(httpMonitor.OAuthTokenURL)
	m.OAuthClientID = stringOrNullKeyword(httpMonitor.OAuthClientID)
	m.OAuthClientSecret = stringOrNullKeyword(httpMonitor.OAuthClientSecret)
	m.OAuthScopes = stringOrNullKeyword(httpMonitor.OAuthScopes)
	m.CacheBust = types.BoolValue(httpMonitor.CacheBust)
}

// populateOptionalFieldsForKeyword populates optional fields for HTTP Keyword monitor.
// Handles parent group, proxy, status codes, and notification configuration.
// This function follows the same pattern as JSON Query to ensure consistency.
// Parent and proxy fields are optional and may be null in the API response.
// Lists are properly converted to Terraform types for accurate state representation.
func populateOptionalFieldsForKeyword(
	ctx context.Context,
	httpMonitor *monitor.HTTP,
	m *MonitorHTTPKeywordResourceModel,
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
func (r *MonitorHTTPKeywordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorHTTPKeywordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var httpKeywordMonitor monitor.HTTPKeyword
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &httpKeywordMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read HTTP Keyword monitor", err.Error())
		return
	}

	var httpMonitor monitor.HTTP
	httpMonitor.Base = httpKeywordMonitor.Base
	httpMonitor.HTTPDetails = httpKeywordMonitor.HTTPDetails
	populateHTTPBaseFieldsForKeyword(&httpMonitor, &data)
	populateOptionalFieldsForKeyword(ctx, &httpMonitor, &data, &resp.Diagnostics)

	data.Keyword = types.StringValue(httpKeywordMonitor.Keyword)
	data.InvertKeyword = types.BoolValue(httpKeywordMonitor.InvertKeyword)

	data.Tags = handleMonitorTagsRead(ctx, httpKeywordMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MonitorHTTPKeywordResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorHTTPKeywordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorHTTPKeywordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpKeywordMonitor := buildHTTPKeywordMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpKeywordMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, httpKeywordMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update HTTP Keyword monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *MonitorHTTPKeywordResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorHTTPKeywordResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API.
	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete HTTP Keyword monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorHTTPKeywordResource) ImportState(
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
