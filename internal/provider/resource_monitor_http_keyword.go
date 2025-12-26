package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = &MonitorHTTPKeywordResource{}
	_ resource.ResourceWithImportState = &MonitorHTTPKeywordResource{}
)

func NewMonitorHTTPKeywordResource() resource.Resource {
	return &MonitorHTTPKeywordResource{}
}

type MonitorHTTPKeywordResource struct {
	client *kuma.Client
}

type MonitorHTTPKeywordResourceModel struct {
	MonitorBaseModel
	MonitorHTTPBaseModel

	Keyword       types.String `tfsdk:"keyword"`
	InvertKeyword types.Bool   `tfsdk:"invert_keyword"`
}

func (r *MonitorHTTPKeywordResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http_keyword"
}

func (r *MonitorHTTPKeywordResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
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

func (r *MonitorHTTPKeywordResource) Configure(
	ctx context.Context,
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
		resp.Diagnostics.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		httpKeywordMonitor.AcceptedStatusCodes = statusCodes
	} else {
		httpKeywordMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		httpKeywordMonitor.NotificationIDs = notificationIDs
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

	data.Name = types.StringValue(httpKeywordMonitor.Name)
	if httpKeywordMonitor.Description != nil {
		data.Description = types.StringValue(*httpKeywordMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(httpKeywordMonitor.Interval)
	data.RetryInterval = types.Int64Value(httpKeywordMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(httpKeywordMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(httpKeywordMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(httpKeywordMonitor.UpsideDown)
	data.Active = types.BoolValue(httpKeywordMonitor.IsActive)
	data.URL = types.StringValue(httpKeywordMonitor.URL)
	data.Timeout = types.Int64Value(httpKeywordMonitor.Timeout)
	data.Method = types.StringValue(httpKeywordMonitor.Method)
	data.ExpiryNotification = types.BoolValue(httpKeywordMonitor.ExpiryNotification)
	data.IgnoreTLS = types.BoolValue(httpKeywordMonitor.IgnoreTLS)
	data.MaxRedirects = types.Int64Value(int64(httpKeywordMonitor.MaxRedirects))
	data.HTTPBodyEncoding = types.StringValue(httpKeywordMonitor.HTTPBodyEncoding)
	data.Body = stringOrNull(httpKeywordMonitor.Body)
	data.Headers = stringOrNull(httpKeywordMonitor.Headers)
	data.AuthMethod = types.StringValue(string(httpKeywordMonitor.AuthMethod))
	data.BasicAuthUser = stringOrNull(httpKeywordMonitor.BasicAuthUser)
	data.BasicAuthPass = stringOrNull(httpKeywordMonitor.BasicAuthPass)
	data.AuthDomain = stringOrNull(httpKeywordMonitor.AuthDomain)
	data.AuthWorkstation = stringOrNull(httpKeywordMonitor.AuthWorkstation)
	data.TLSCert = stringOrNull(httpKeywordMonitor.TLSCert)
	data.TLSKey = stringOrNull(httpKeywordMonitor.TLSKey)
	data.TLSCa = stringOrNull(httpKeywordMonitor.TLSCa)
	data.OAuthAuthMethod = stringOrNull(httpKeywordMonitor.OAuthAuthMethod)
	data.OAuthTokenURL = stringOrNull(httpKeywordMonitor.OAuthTokenURL)
	data.OAuthClientID = stringOrNull(httpKeywordMonitor.OAuthClientID)
	data.OAuthClientSecret = stringOrNull(httpKeywordMonitor.OAuthClientSecret)
	data.OAuthScopes = stringOrNull(httpKeywordMonitor.OAuthScopes)
	data.Keyword = types.StringValue(httpKeywordMonitor.Keyword)
	data.InvertKeyword = types.BoolValue(httpKeywordMonitor.InvertKeyword)

	if httpKeywordMonitor.Parent != nil {
		data.Parent = types.Int64Value(*httpKeywordMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if httpKeywordMonitor.ProxyID != nil {
		data.ProxyID = types.Int64Value(*httpKeywordMonitor.ProxyID)
	} else {
		data.ProxyID = types.Int64Null()
	}

	if len(httpKeywordMonitor.AcceptedStatusCodes) > 0 {
		statusCodes, diags := types.ListValueFrom(ctx, types.StringType, httpKeywordMonitor.AcceptedStatusCodes)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.AcceptedStatusCodes = statusCodes
	}

	if len(httpKeywordMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, httpKeywordMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, httpKeywordMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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

	httpKeywordMonitor := monitor.HTTPKeyword{
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
		resp.Diagnostics.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		httpKeywordMonitor.AcceptedStatusCodes = statusCodes
	} else {
		httpKeywordMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		httpKeywordMonitor.NotificationIDs = notificationIDs
	}

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

func (r *MonitorHTTPKeywordResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorHTTPKeywordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete HTTP Keyword monitor", err.Error())
		return
	}
}

func (r *MonitorHTTPKeywordResource) ImportState(
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
