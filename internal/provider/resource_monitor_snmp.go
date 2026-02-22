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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	// Ensure MonitorSNMPResource satisfies various resource interfaces.
	_ resource.Resource                = &MonitorSNMPResource{}
	_ resource.ResourceWithImportState = &MonitorSNMPResource{}
)

// NewMonitorSNMPResource returns a new instance of the SNMP monitor resource.
func NewMonitorSNMPResource() resource.Resource {
	return &MonitorSNMPResource{}
}

// MonitorSNMPResource defines the resource implementation.
type MonitorSNMPResource struct {
	client *kuma.Client
}

// MonitorSNMPResourceModel describes the resource data model for SNMP monitors.
type MonitorSNMPResourceModel struct {
	MonitorBaseModel

	Hostname         types.String `tfsdk:"hostname"`
	Port             types.Int64  `tfsdk:"port"`
	SNMPVersion      types.String `tfsdk:"snmp_version"`
	SNMPOID          types.String `tfsdk:"snmp_oid"`
	SNMPCommunity    types.String `tfsdk:"snmp_community"`
	JSONPath         types.String `tfsdk:"json_path"`
	JSONPathOperator types.String `tfsdk:"json_path_operator"`
	ExpectedValue    types.String `tfsdk:"expected_value"`
}

// Metadata returns the metadata for the resource.
func (*MonitorSNMPResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_snmp"
}

// Schema returns the schema for the resource.
func (*MonitorSNMPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SNMP monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "SNMP device hostname or IP address",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "SNMP device port",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(161),
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
			"snmp_version": schema.StringAttribute{
				MarkdownDescription: "SNMP version (e.g., '2c', '3')",
				Required:            true,
			},
			"snmp_oid": schema.StringAttribute{
				MarkdownDescription: "SNMP Object Identifier (OID) to query",
				Required:            true,
			},
			"snmp_community": schema.StringAttribute{
				MarkdownDescription: "SNMP community string",
				Required:            true,
				Sensitive:           true,
			},
			"json_path": schema.StringAttribute{
				MarkdownDescription: "JSON path for extracting value from SNMP response",
				Optional:            true,
			},
			"json_path_operator": schema.StringAttribute{
				MarkdownDescription: "Comparison operator for JSON path result. Valid values: `>`, `>=`, `<`, `<=`, `!=`, `==`, `contains`",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(">", ">=", "<", "<=", "!=", "==", "contains"),
				},
			},
			"expected_value": schema.StringAttribute{
				MarkdownDescription: "Expected value to match",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorSNMPResource) Configure(
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
func (r *MonitorSNMPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorSNMPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snmpMonitor := buildSNMPMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, &snmpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create SNMP monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildSNMPMonitor(ctx context.Context, data *MonitorSNMPResourceModel, diags *diag.Diagnostics) monitor.SNMP {
	port := data.Port.ValueInt64()
	snmpMonitor := monitor.SNMP{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		SNMPDetails: monitor.SNMPDetails{
			Hostname:      data.Hostname.ValueString(),
			Port:          &port,
			SNMPVersion:   data.SNMPVersion.ValueString(),
			SNMPOID:       data.SNMPOID.ValueString(),
			SNMPCommunity: data.SNMPCommunity.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		snmpMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		snmpMonitor.Parent = &parent
	}

	if !data.JSONPath.IsNull() {
		jsonPath := data.JSONPath.ValueString()
		snmpMonitor.JSONPath = &jsonPath
	}

	if !data.JSONPathOperator.IsNull() {
		jsonPathOperator := data.JSONPathOperator.ValueString()
		snmpMonitor.JSONPathOperator = &jsonPathOperator
	}

	if !data.ExpectedValue.IsNull() {
		expectedValue := data.ExpectedValue.ValueString()
		snmpMonitor.ExpectedValue = &expectedValue
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			snmpMonitor.NotificationIDs = notificationIDs
		}
	}

	return snmpMonitor
}

// populateSNMPMonitorBaseFields populates the resource model with data from the SNMP monitor API response.
func populateSNMPMonitorBaseFields(snmpMonitor *monitor.SNMP, m *MonitorSNMPResourceModel) {
	m.Name = types.StringValue(snmpMonitor.Name)
	if snmpMonitor.Description != nil {
		m.Description = types.StringValue(*snmpMonitor.Description)
	} else {
		m.Description = types.StringNull()
	}

	m.Interval = types.Int64Value(snmpMonitor.Interval)
	m.RetryInterval = types.Int64Value(snmpMonitor.RetryInterval)
	m.ResendInterval = types.Int64Value(snmpMonitor.ResendInterval)
	m.MaxRetries = types.Int64Value(snmpMonitor.MaxRetries)
	m.UpsideDown = types.BoolValue(snmpMonitor.UpsideDown)
	m.Active = types.BoolValue(snmpMonitor.IsActive)
	m.Hostname = types.StringValue(snmpMonitor.Hostname)
	m.SNMPVersion = types.StringValue(snmpMonitor.SNMPVersion)
	m.SNMPOID = types.StringValue(snmpMonitor.SNMPOID)
	m.SNMPCommunity = types.StringValue(snmpMonitor.SNMPCommunity)
	m.JSONPath = stringOrNullPtr(snmpMonitor.JSONPath)
	m.JSONPathOperator = stringOrNullPtr(snmpMonitor.JSONPathOperator)
	m.ExpectedValue = stringOrNullPtr(snmpMonitor.ExpectedValue)

	if snmpMonitor.Port != nil {
		m.Port = types.Int64Value(*snmpMonitor.Port)
	} else {
		m.Port = types.Int64Value(161)
	}
}

// stringOrNullPtr returns a Terraform String type that is null if the input pointer is nil or empty,
// otherwise returns the string value.
func stringOrNullPtr(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}

	return types.StringValue(*s)
}

// populateOptionalFieldsForSNMP populates optional parent and notification fields from the SNMP monitor API response.
func populateOptionalFieldsForSNMP(
	ctx context.Context,
	snmpMonitor *monitor.SNMP,
	m *MonitorSNMPResourceModel,
	diags *diag.Diagnostics,
) {
	// Set parent monitor group if present.
	if snmpMonitor.Parent != nil {
		m.Parent = types.Int64Value(*snmpMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	// Convert notification IDs list if present.
	if len(snmpMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, snmpMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// Read reads the current state of the resource.
func (r *MonitorSNMPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorSNMPResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var snmpMonitor monitor.SNMP
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &snmpMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read SNMP monitor", err.Error())
		return
	}

	populateSNMPMonitorBaseFields(&snmpMonitor, &data)
	populateOptionalFieldsForSNMP(ctx, &snmpMonitor, &data, &resp.Diagnostics)

	data.Tags = handleMonitorTagsRead(ctx, snmpMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MonitorSNMPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorSNMPResourceModel
	var state MonitorSNMPResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snmpMonitor := buildSNMPMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	snmpMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, &snmpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update SNMP monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *MonitorSNMPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorSNMPResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API.
	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete SNMP monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorSNMPResource) ImportState(
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
