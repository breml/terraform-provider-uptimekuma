package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ resource.Resource = &MonitorHTTPResource{}

func NewMonitorHTTPResource() resource.Resource {
	return &MonitorHTTPResource{}
}

type MonitorHTTPResource struct {
	client *kuma.Client
}

type MonitorHTTPResourceModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Parent              types.Int64  `tfsdk:"parent"`
	Interval            types.Int64  `tfsdk:"interval"`
	RetryInterval       types.Int64  `tfsdk:"retry_interval"`
	ResendInterval      types.Int64  `tfsdk:"resend_interval"`
	MaxRetries          types.Int64  `tfsdk:"max_retries"`
	UpsideDown          types.Bool   `tfsdk:"upside_down"`
	Active              types.Bool   `tfsdk:"active"`
	URL                 types.String `tfsdk:"url"`
	Timeout             types.Int64  `tfsdk:"timeout"`
	Method              types.String `tfsdk:"method"`
	ExpiryNotification  types.Bool   `tfsdk:"expiry_notification"`
	IgnoreTLS           types.Bool   `tfsdk:"ignore_tls"`
	MaxRedirects        types.Int64  `tfsdk:"max_redirects"`
	AcceptedStatusCodes types.List   `tfsdk:"accepted_status_codes"`
	ProxyID             types.Int64  `tfsdk:"proxy_id"`
	HTTPBodyEncoding    types.String `tfsdk:"http_body_encoding"`
	Body                types.String `tfsdk:"body"`
	Headers             types.String `tfsdk:"headers"`
	AuthMethod          types.String `tfsdk:"auth_method"`
	BasicAuthUser       types.String `tfsdk:"basic_auth_user"`
	BasicAuthPass       types.String `tfsdk:"basic_auth_pass"`
	AuthDomain          types.String `tfsdk:"auth_domain"`
	AuthWorkstation     types.String `tfsdk:"auth_workstation"`
	TLSCert             types.String `tfsdk:"tls_cert"`
	TLSKey              types.String `tfsdk:"tls_key"`
	TLSCa               types.String `tfsdk:"tls_ca"`
	OAuthAuthMethod     types.String `tfsdk:"oauth_auth_method"`
	OAuthTokenURL       types.String `tfsdk:"oauth_token_url"`
	OAuthClientID       types.String `tfsdk:"oauth_client_id"`
	OAuthClientSecret   types.String `tfsdk:"oauth_client_secret"`
	OAuthScopes         types.String `tfsdk:"oauth_scopes"`
	NotificationIDs     types.List   `tfsdk:"notification_ids"`
}

func (r *MonitorHTTPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http"
}

func (r *MonitorHTTPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "HTTP monitor resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Monitor identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Friendly name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description",
				Optional:            true,
			},
			"parent": schema.Int64Attribute{
				MarkdownDescription: "Parent monitor ID for hierarchical organization",
				Optional:            true,
			},
			"interval": schema.Int64Attribute{
				MarkdownDescription: "Heartbeat interval in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				Validators: []validator.Int64{
					int64validator.Between(20, 2073600),
				},
			},
			"retry_interval": schema.Int64Attribute{
				MarkdownDescription: "Retry interval in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				Validators: []validator.Int64{
					int64validator.Between(20, 2073600),
				},
			},
			"resend_interval": schema.Int64Attribute{
				MarkdownDescription: "Resend interval in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"max_retries": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of retries",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3),
				Validators: []validator.Int64{
					int64validator.Between(0, 10),
				},
			},
			"upside_down": schema.BoolAttribute{
				MarkdownDescription: "Invert monitor status (treat DOWN as UP and vice versa)",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Monitor is active",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "URL to monitor",
				Required:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Request timeout in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(48),
				Validators: []validator.Int64{
					int64validator.Between(1, 3600),
				},
			},
			"method": schema.StringAttribute{
				MarkdownDescription: "HTTP method",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("GET"),
				Validators: []validator.String{
					stringvalidator.OneOf("GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"),
				},
			},
			"expiry_notification": schema.BoolAttribute{
				MarkdownDescription: "Enable certificate expiry notification",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ignore_tls": schema.BoolAttribute{
				MarkdownDescription: "Ignore TLS/SSL errors",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"max_redirects": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of redirects to follow",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(10),
				Validators: []validator.Int64{
					int64validator.Between(0, 20),
				},
			},
			"accepted_status_codes": schema.ListAttribute{
				MarkdownDescription: "Accepted HTTP status codes (e.g., ['200-299', '301'])",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200-299")})),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"proxy_id": schema.Int64Attribute{
				MarkdownDescription: "Proxy ID",
				Optional:            true,
			},
			"http_body_encoding": schema.StringAttribute{
				MarkdownDescription: "HTTP body encoding",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("json"),
			},
			"body": schema.StringAttribute{
				MarkdownDescription: "Request body",
				Optional:            true,
			},
			"headers": schema.StringAttribute{
				MarkdownDescription: "Request headers (JSON format)",
				Optional:            true,
			},
			"auth_method": schema.StringAttribute{
				MarkdownDescription: "Authentication method",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.OneOf("", "basic", "ntlm", "mtls", "oauth2-cc"),
				},
			},
			"basic_auth_user": schema.StringAttribute{
				MarkdownDescription: "Basic authentication username",
				Optional:            true,
			},
			"basic_auth_pass": schema.StringAttribute{
				MarkdownDescription: "Basic authentication password",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_domain": schema.StringAttribute{
				MarkdownDescription: "NTLM authentication domain",
				Optional:            true,
			},
			"auth_workstation": schema.StringAttribute{
				MarkdownDescription: "NTLM authentication workstation",
				Optional:            true,
			},
			"tls_cert": schema.StringAttribute{
				MarkdownDescription: "TLS client certificate",
				Optional:            true,
				Sensitive:           true,
			},
			"tls_key": schema.StringAttribute{
				MarkdownDescription: "TLS client key",
				Optional:            true,
				Sensitive:           true,
			},
			"tls_ca": schema.StringAttribute{
				MarkdownDescription: "TLS CA certificate",
				Optional:            true,
			},
			"oauth_auth_method": schema.StringAttribute{
				MarkdownDescription: "OAuth authentication method",
				Optional:            true,
			},
			"oauth_token_url": schema.StringAttribute{
				MarkdownDescription: "OAuth token URL",
				Optional:            true,
			},
			"oauth_client_id": schema.StringAttribute{
				MarkdownDescription: "OAuth client ID",
				Optional:            true,
			},
			"oauth_client_secret": schema.StringAttribute{
				MarkdownDescription: "OAuth client secret",
				Optional:            true,
				Sensitive:           true,
			},
			"oauth_scopes": schema.StringAttribute{
				MarkdownDescription: "OAuth scopes",
				Optional:            true,
			},
			"notification_ids": schema.ListAttribute{
				MarkdownDescription: "List of notification IDs",
				ElementType:         types.Int64Type,
				Optional:            true,
			},
		},
	}
}

func (r *MonitorHTTPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *kuma.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorHTTPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorHTTPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
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
