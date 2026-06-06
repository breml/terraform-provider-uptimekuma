package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorGlobalpingResource{}
	_ resource.ResourceWithImportState = &MonitorGlobalpingResource{}
)

// NewMonitorGlobalpingResource returns a new instance of the Globalping monitor resource.
func NewMonitorGlobalpingResource() resource.Resource {
	return &MonitorGlobalpingResource{}
}

// MonitorGlobalpingResource defines the resource implementation for Globalping monitors.
type MonitorGlobalpingResource struct {
	client *kuma.Client
}

// MonitorGlobalpingResourceModel describes the resource data model for Globalping monitors.
type MonitorGlobalpingResourceModel struct {
	MonitorBaseModel
	MonitorHTTPBaseModel

	// Subtype is the check type performed by Globalping probes.
	Subtype types.String `tfsdk:"subtype"`
	// Location is the probe location selector (e.g. "Europe", "us-east").
	Location types.String `tfsdk:"location"`
	// IPFamily selects the IP protocol version ("", "ipv4", or "ipv6").
	IPFamily types.String `tfsdk:"ip_family"`
	// Protocol is the protocol used for ping/traceroute checks.
	Protocol types.String `tfsdk:"protocol"`
	// PingCount is the number of ping packets to send.
	PingCount types.Int64 `tfsdk:"ping_count"`
	// Hostname is the target hostname for DNS or port checks.
	Hostname types.String `tfsdk:"hostname"`
	// Port is the target port for port checks.
	Port types.Int64 `tfsdk:"port"`
	// DNSResolveType is the DNS record type to resolve (e.g. "A", "AAAA").
	DNSResolveType types.String `tfsdk:"dns_resolve_type"`
	// DNSResolveServer is the DNS server to use for resolution.
	DNSResolveServer types.String `tfsdk:"dns_resolve_server"`
	// Keyword is the keyword to search for in the HTTP response.
	Keyword types.String `tfsdk:"keyword"`
	// InvertKeyword inverts the keyword match logic.
	InvertKeyword types.Bool `tfsdk:"invert_keyword"`
	// ExpectedValue is the expected value for JSON path checks.
	ExpectedValue types.String `tfsdk:"expected_value"`
	// JSONPath is the JSON path expression to evaluate.
	JSONPath types.String `tfsdk:"json_path"`
	// JSONPathOperator is the comparison operator for the JSON path check.
	JSONPathOperator types.String `tfsdk:"json_path_operator"`
}

// Metadata returns the metadata for the resource.
func (*MonitorGlobalpingResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_globalping"
}

// Schema returns the schema for the resource.
func (*MonitorGlobalpingResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Globalping monitor resource. Globalping runs distributed checks (ping, traceroute, DNS, HTTP) from a network of community probes.",
		Attributes: withMonitorBaseAttributes(withHTTPMonitorBaseAttributes(map[string]schema.Attribute{
			"subtype": schema.StringAttribute{
				MarkdownDescription: "Check type performed by Globalping probes. One of `ping`, `traceroute`, `dns`, `http`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("ping", "traceroute", "dns", "http"),
				},
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Probe location selector (e.g. `Europe`, `us-east`). Leave empty to use any available probe.",
				Optional:            true,
			},
			"ip_family": schema.StringAttribute{
				MarkdownDescription: "IP protocol version to use. One of `\"\"` (auto), `ipv4`, `ipv6`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.OneOf("", "ipv4", "ipv6"),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol used for ping or traceroute checks (e.g. `ICMP`, `TCP`).",
				Optional:            true,
			},
			"ping_count": schema.Int64Attribute{
				MarkdownDescription: "Number of ping packets to send.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Target hostname for DNS or port checks.",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Target port for port checks.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
			"dns_resolve_type": schema.StringAttribute{
				MarkdownDescription: "DNS record type to resolve (e.g. `A`, `AAAA`, `MX`). Used when `subtype` is `dns`.",
				Optional:            true,
			},
			"dns_resolve_server": schema.StringAttribute{
				MarkdownDescription: "DNS server to use for resolution. Used when `subtype` is `dns`.",
				Optional:            true,
			},
			"keyword": schema.StringAttribute{
				MarkdownDescription: "Keyword to search for in the HTTP response body. Used when `subtype` is `http`.",
				Optional:            true,
			},
			"invert_keyword": schema.BoolAttribute{
				MarkdownDescription: "Invert the keyword match logic. When true, the monitor is UP when the keyword is NOT found.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"expected_value": schema.StringAttribute{
				MarkdownDescription: "Expected value for JSON path evaluation.",
				Optional:            true,
			},
			"json_path": schema.StringAttribute{
				MarkdownDescription: "JSON path expression to evaluate in the response.",
				Optional:            true,
			},
			"json_path_operator": schema.StringAttribute{
				MarkdownDescription: "Comparison operator for the JSON path check.",
				Optional:            true,
			},
		})),
	}
}

// Configure configures the Globalping monitor resource with the API client.
func (r *MonitorGlobalpingResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Globalping monitor resource.
func (r *MonitorGlobalpingResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorGlobalpingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	globalpingMonitor := buildGlobalpingMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &globalpingMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Globalping monitor", err.Error())
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
		resp.Diagnostics.AddError("failed to apply monitor active state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Globalping monitor resource.
func (r *MonitorGlobalpingResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorGlobalpingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var globalpingMonitor monitor.Globalping
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &globalpingMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read Globalping monitor", err.Error())
		return
	}

	if actual := globalpingMonitor.Base.Type(); actual != "" && actual != globalpingMonitor.Type() {
		tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
			"id":            data.ID.ValueInt64(),
			"expected_type": globalpingMonitor.Type(),
			"actual_type":   actual,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	populateGlobalpingModel(&globalpingMonitor, &data)
	populateGlobalpingOptionalFields(ctx, &globalpingMonitor, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Globalping monitor resource.
func (r *MonitorGlobalpingResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorGlobalpingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorGlobalpingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	globalpingMonitor := buildGlobalpingMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	globalpingMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &globalpingMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update Globalping monitor", err.Error())
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Globalping monitor resource.
func (r *MonitorGlobalpingResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorGlobalpingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Globalping monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorGlobalpingResource) ImportState(
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

// buildGlobalpingMonitor constructs a Globalping monitor API object from the Terraform resource model.
func buildGlobalpingMonitor(
	ctx context.Context,
	data *MonitorGlobalpingResourceModel,
	diags *diag.Diagnostics,
) monitor.Globalping {
	globalpingMonitor := monitor.Globalping{
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
		GlobalpingDetails: monitor.GlobalpingDetails{
			Subtype:          monitor.GlobalpingSubtype(data.Subtype.ValueString()),
			Location:         data.Location.ValueString(),
			IPFamily:         monitor.GlobalpingIPFamily(data.IPFamily.ValueString()),
			Protocol:         data.Protocol.ValueString(),
			PingCount:        int(data.PingCount.ValueInt64()),
			Hostname:         data.Hostname.ValueString(),
			Port:             int(data.Port.ValueInt64()),
			DNSResolveType:   monitor.DNSResolveType(data.DNSResolveType.ValueString()),
			DNSResolveServer: data.DNSResolveServer.ValueString(),
			Keyword:          data.Keyword.ValueString(),
			InvertKeyword:    data.InvertKeyword.ValueBool(),
			ExpectedValue:    data.ExpectedValue.ValueString(),
			JSONPath:         data.JSONPath.ValueString(),
			JSONPathOperator: data.JSONPathOperator.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		globalpingMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		globalpingMonitor.Parent = &parent
	}

	if !data.ProxyID.IsNull() {
		proxyID := data.ProxyID.ValueInt64()
		globalpingMonitor.ProxyID = &proxyID
	}

	if !data.AcceptedStatusCodes.IsNull() && !data.AcceptedStatusCodes.IsUnknown() {
		var statusCodes []string
		diags.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if !diags.HasError() {
			globalpingMonitor.AcceptedStatusCodes = statusCodes
		}
	} else {
		globalpingMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			globalpingMonitor.NotificationIDs = notificationIDs
		}
	}

	return globalpingMonitor
}

// populateGlobalpingModel populates the Terraform model from the Globalping monitor API response.
func populateGlobalpingModel(globalpingMonitor *monitor.Globalping, data *MonitorGlobalpingResourceModel) {
	data.Name = types.StringValue(globalpingMonitor.Name)
	if globalpingMonitor.Description != nil {
		data.Description = types.StringValue(*globalpingMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(globalpingMonitor.Interval)
	data.RetryInterval = types.Int64Value(globalpingMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(globalpingMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(globalpingMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(globalpingMonitor.UpsideDown)
	data.Active = types.BoolValue(globalpingMonitor.IsActive)

	data.URL = types.StringValue(globalpingMonitor.URL)
	data.Timeout = types.Int64Value(globalpingMonitor.Timeout)
	data.Method = types.StringValue(globalpingMonitor.Method)
	data.ExpiryNotification = types.BoolValue(globalpingMonitor.ExpiryNotification)
	data.IgnoreTLS = types.BoolValue(globalpingMonitor.IgnoreTLS)
	data.MaxRedirects = types.Int64Value(int64(globalpingMonitor.MaxRedirects))
	data.HTTPBodyEncoding = types.StringValue(globalpingMonitor.HTTPBodyEncoding)
	data.Body = stringOrNull(globalpingMonitor.Body)
	data.Headers = stringOrNull(globalpingMonitor.Headers)
	data.AuthMethod = types.StringValue(string(globalpingMonitor.AuthMethod))
	data.BasicAuthUser = stringOrNull(globalpingMonitor.BasicAuthUser)
	data.BasicAuthPass = stringOrNull(globalpingMonitor.BasicAuthPass)
	data.AuthDomain = stringOrNull(globalpingMonitor.AuthDomain)
	data.AuthWorkstation = stringOrNull(globalpingMonitor.AuthWorkstation)
	data.TLSCert = stringOrNull(globalpingMonitor.TLSCert)
	data.TLSKey = stringOrNull(globalpingMonitor.TLSKey)
	data.TLSCa = stringOrNull(globalpingMonitor.TLSCa)
	data.OAuthAuthMethod = stringOrNull(globalpingMonitor.OAuthAuthMethod)
	data.OAuthTokenURL = stringOrNull(globalpingMonitor.OAuthTokenURL)
	data.OAuthClientID = stringOrNull(globalpingMonitor.OAuthClientID)
	data.OAuthClientSecret = stringOrNull(globalpingMonitor.OAuthClientSecret)
	data.OAuthScopes = stringOrNull(globalpingMonitor.OAuthScopes)
	data.OAuthAudience = stringOrNull(globalpingMonitor.OAuthAudience)
	data.CacheBust = types.BoolValue(globalpingMonitor.CacheBust)

	data.Subtype = types.StringValue(string(globalpingMonitor.Subtype))
	data.Location = stringOrNull(globalpingMonitor.Location)
	data.IPFamily = types.StringValue(string(globalpingMonitor.IPFamily))
	data.Protocol = stringOrNull(globalpingMonitor.Protocol)
	data.PingCount = types.Int64Value(int64(globalpingMonitor.PingCount))
	data.Hostname = stringOrNull(globalpingMonitor.Hostname)
	data.Port = types.Int64Value(int64(globalpingMonitor.Port))
	data.DNSResolveType = stringOrNull(string(globalpingMonitor.DNSResolveType))
	data.DNSResolveServer = stringOrNull(globalpingMonitor.DNSResolveServer)
	data.Keyword = stringOrNull(globalpingMonitor.Keyword)
	data.InvertKeyword = types.BoolValue(globalpingMonitor.InvertKeyword)
	data.ExpectedValue = stringOrNull(globalpingMonitor.ExpectedValue)
	data.JSONPath = stringOrNull(globalpingMonitor.JSONPath)
	data.JSONPathOperator = stringOrNull(globalpingMonitor.JSONPathOperator)
}

// populateGlobalpingOptionalFields populates optional and computed fields from the API response.
func populateGlobalpingOptionalFields(
	ctx context.Context,
	globalpingMonitor *monitor.Globalping,
	data *MonitorGlobalpingResourceModel,
	diags *diag.Diagnostics,
) {
	if globalpingMonitor.Parent != nil {
		data.Parent = types.Int64Value(*globalpingMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if globalpingMonitor.ProxyID != nil {
		data.ProxyID = types.Int64Value(*globalpingMonitor.ProxyID)
	} else {
		data.ProxyID = types.Int64Null()
	}

	if len(globalpingMonitor.AcceptedStatusCodes) > 0 {
		statusCodes, d := types.ListValueFrom(ctx, types.StringType, globalpingMonitor.AcceptedStatusCodes)
		diags.Append(d...)
		data.AcceptedStatusCodes = statusCodes
	} else {
		data.AcceptedStatusCodes = types.ListNull(types.StringType)
	}

	if len(globalpingMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, globalpingMonitor.NotificationIDs)
		diags.Append(d...)
		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, globalpingMonitor.Tags, data.Tags, diags)
}
