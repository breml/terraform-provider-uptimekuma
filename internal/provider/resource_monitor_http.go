package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorHTTPResource{}
	_ resource.ResourceWithImportState = &MonitorHTTPResource{}
)

func NewMonitorHTTPResource() resource.Resource {
	return &MonitorHTTPResource{}
}

type MonitorHTTPResource struct {
	client *kuma.Client
}

type MonitorHTTPResourceModel struct {
	MonitorBaseModel
	MonitorHTTPBaseModel
}

func (r *MonitorHTTPResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http"
}

func (r *MonitorHTTPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "HTTP monitor resource",
		Attributes:          withMonitorBaseAttributes(withHTTPMonitorBaseAttributes(map[string]schema.Attribute{})),
	}
}

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

func (r *MonitorHTTPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorHTTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

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
		resp.Diagnostics.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		httpMonitor.AcceptedStatusCodes = statusCodes
	} else {
		httpMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		httpMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, httpMonitor)
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

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}

	return types.StringValue(s)
}

func (r *MonitorHTTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorHTTPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var httpMonitor monitor.HTTP
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &httpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read HTTP monitor", err.Error())
		return
	}

	data.Name = types.StringValue(httpMonitor.Name)
	if httpMonitor.Description != nil {
		data.Description = types.StringValue(*httpMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(httpMonitor.Interval)
	data.RetryInterval = types.Int64Value(httpMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(httpMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(httpMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(httpMonitor.UpsideDown)
	data.Active = types.BoolValue(httpMonitor.IsActive)
	data.URL = types.StringValue(httpMonitor.URL)
	data.Timeout = types.Int64Value(httpMonitor.Timeout)
	data.Method = types.StringValue(httpMonitor.Method)
	data.ExpiryNotification = types.BoolValue(httpMonitor.ExpiryNotification)
	data.IgnoreTLS = types.BoolValue(httpMonitor.IgnoreTLS)
	data.MaxRedirects = types.Int64Value(int64(httpMonitor.MaxRedirects))
	data.HTTPBodyEncoding = types.StringValue(httpMonitor.HTTPBodyEncoding)
	data.Body = stringOrNull(httpMonitor.Body)
	data.Headers = stringOrNull(httpMonitor.Headers)
	data.AuthMethod = types.StringValue(string(httpMonitor.AuthMethod))
	data.BasicAuthUser = stringOrNull(httpMonitor.BasicAuthUser)
	data.BasicAuthPass = stringOrNull(httpMonitor.BasicAuthPass)
	data.AuthDomain = stringOrNull(httpMonitor.AuthDomain)
	data.AuthWorkstation = stringOrNull(httpMonitor.AuthWorkstation)
	data.TLSCert = stringOrNull(httpMonitor.TLSCert)
	data.TLSKey = stringOrNull(httpMonitor.TLSKey)
	data.TLSCa = stringOrNull(httpMonitor.TLSCa)
	data.OAuthAuthMethod = stringOrNull(httpMonitor.OAuthAuthMethod)
	data.OAuthTokenURL = stringOrNull(httpMonitor.OAuthTokenURL)
	data.OAuthClientID = stringOrNull(httpMonitor.OAuthClientID)
	data.OAuthClientSecret = stringOrNull(httpMonitor.OAuthClientSecret)
	data.OAuthScopes = stringOrNull(httpMonitor.OAuthScopes)

	if httpMonitor.Parent != nil {
		data.Parent = types.Int64Value(*httpMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if httpMonitor.ProxyID != nil {
		data.ProxyID = types.Int64Value(*httpMonitor.ProxyID)
	} else {
		data.ProxyID = types.Int64Null()
	}

	if len(httpMonitor.AcceptedStatusCodes) > 0 {
		statusCodes, diags := types.ListValueFrom(ctx, types.StringType, httpMonitor.AcceptedStatusCodes)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.AcceptedStatusCodes = statusCodes
	}

	if len(httpMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, httpMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, httpMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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

	httpMonitor := monitor.HTTP{
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
		resp.Diagnostics.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		httpMonitor.AcceptedStatusCodes = statusCodes
	} else {
		httpMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		httpMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, httpMonitor)
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

func (r *MonitorHTTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorHTTPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete HTTP monitor", err.Error())
		return
	}
}

func (r *MonitorHTTPResource) ImportState(
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
