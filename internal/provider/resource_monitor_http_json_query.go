// Package provider implements the Uptime Kuma Terraform provider.
// This file provides HTTP JSON Query monitor resource management.
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorHTTPJSONQueryResource{}
	_ resource.ResourceWithImportState = &MonitorHTTPJSONQueryResource{}
)

// NewMonitorHTTPJSONQueryResource returns a new instance of the HTTP JSON Query monitor resource.
func NewMonitorHTTPJSONQueryResource() resource.Resource {
	return &MonitorHTTPJSONQueryResource{}
}

// MonitorHTTPJSONQueryResource defines the resource implementation.
type MonitorHTTPJSONQueryResource struct {
	client *kuma.Client
}

// MonitorHTTPJSONQueryResourceModel describes the resource data model.
type MonitorHTTPJSONQueryResourceModel struct {
	MonitorBaseModel
	MonitorHTTPBaseModel

	JSONPath         types.String `tfsdk:"json_path"`
	ExpectedValue    types.String `tfsdk:"expected_value"`
	JSONPathOperator types.String `tfsdk:"json_path_operator"`
}

// Metadata returns the metadata for the resource.
func (*MonitorHTTPJSONQueryResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http_json_query"
}

// Schema returns the schema for the resource.
func (*MonitorHTTPJSONQueryResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	// Define resource schema attributes and validation.
	resp.Schema = schema.Schema{
		MarkdownDescription: "The HTTP JSON Query monitor allows you to monitor an HTTP endpoint by querying its JSON response using a JSONPath expression. This monitor extracts a value from the JSON response at the specified path and compares it to an expected value using a configurable comparison operator.",
		Attributes: withMonitorBaseAttributes(withHTTPMonitorBaseAttributes(map[string]schema.Attribute{
			"json_path": schema.StringAttribute{
				MarkdownDescription: "JSON Path expression to query the response",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"expected_value": schema.StringAttribute{
				MarkdownDescription: "Expected value to compare against the JSON path result",
				Required:            true,
			},
			"json_path_operator": schema.StringAttribute{
				MarkdownDescription: "Comparison operator for JSON path result. Valid values: `>`, `>=`, `<`, `<=`, `!=`, `==`, `contains`",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("=="),
				Validators: []validator.String{
					stringvalidator.OneOf(">", ">=", "<", "<=", "!=", "==", "contains"),
				},
			},
		})),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorHTTPJSONQueryResource) Configure(
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
func (r *MonitorHTTPJSONQueryResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorHTTPJSONQueryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpJSONQueryMonitor := buildHTTPJSONQueryMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, httpJSONQueryMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create HTTP JSON Query monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildHTTPJSONQueryMonitor(
	ctx context.Context,
	data *MonitorHTTPJSONQueryResourceModel,
	diags *diag.Diagnostics,
) monitor.HTTPJSONQuery {
	httpJSONQueryMonitor := monitor.HTTPJSONQuery{
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
		HTTPJSONQueryDetails: monitor.HTTPJSONQueryDetails{
			JSONPath:         data.JSONPath.ValueString(),
			ExpectedValue:    data.ExpectedValue.ValueString(),
			JSONPathOperator: data.JSONPathOperator.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		httpJSONQueryMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		httpJSONQueryMonitor.Parent = &parent
	}

	if !data.ProxyID.IsNull() {
		proxyID := data.ProxyID.ValueInt64()
		httpJSONQueryMonitor.ProxyID = &proxyID
	}

	if !data.AcceptedStatusCodes.IsNull() && !data.AcceptedStatusCodes.IsUnknown() {
		var statusCodes []string
		diags.Append(data.AcceptedStatusCodes.ElementsAs(ctx, &statusCodes, false)...)
		if !diags.HasError() {
			httpJSONQueryMonitor.AcceptedStatusCodes = statusCodes
		}
	} else {
		httpJSONQueryMonitor.AcceptedStatusCodes = []string{"200-299"}
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			httpJSONQueryMonitor.NotificationIDs = notificationIDs
		}
	}

	return httpJSONQueryMonitor
}

// Read reads the current state of the resource.
func (r *MonitorHTTPJSONQueryResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorHTTPJSONQueryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var httpJSONQueryMonitor monitor.HTTPJSONQuery
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &httpJSONQueryMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read HTTP JSON Query monitor", err.Error())
		return
	}

	var httpMonitor monitor.HTTP
	httpMonitor.Base = httpJSONQueryMonitor.Base
	httpMonitor.HTTPDetails = httpJSONQueryMonitor.HTTPDetails
	populateHTTPMonitorBaseFields(ctx, &httpMonitor, &data, &resp.Diagnostics)
	populateHTTPMonitorOptionalFields(ctx, &httpMonitor, &data, &resp.Diagnostics)

	data.JSONPath = types.StringValue(httpJSONQueryMonitor.JSONPath)
	data.ExpectedValue = types.StringValue(httpJSONQueryMonitor.ExpectedValue)
	data.JSONPathOperator = types.StringValue(httpJSONQueryMonitor.JSONPathOperator)

	data.Tags = handleMonitorTagsRead(ctx, httpJSONQueryMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MonitorHTTPJSONQueryResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorHTTPJSONQueryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorHTTPJSONQueryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpJSONQueryMonitor := buildHTTPJSONQueryMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpJSONQueryMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, httpJSONQueryMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update HTTP JSON Query monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *MonitorHTTPJSONQueryResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorHTTPJSONQueryResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API.
	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete HTTP JSON Query monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorHTTPJSONQueryResource) ImportState(
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
