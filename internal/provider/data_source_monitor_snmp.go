package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorSNMPDataSource{}

// NewMonitorSNMPDataSource returns a new instance of the SNMP monitor data source.
func NewMonitorSNMPDataSource() datasource.DataSource {
	return &MonitorSNMPDataSource{}
}

// MonitorSNMPDataSource manages SNMP monitor data source operations.
type MonitorSNMPDataSource struct {
	client *kuma.Client
}

// MonitorSNMPDataSourceModel describes the data model for SNMP monitor data source.
type MonitorSNMPDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Hostname types.String `tfsdk:"hostname"`
	SNMPOID  types.String `tfsdk:"snmp_oid"`
}

// Metadata returns the metadata for the data source.
func (*MonitorSNMPDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_snmp"
}

// Schema returns the schema for the data source.
func (*MonitorSNMPDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get SNMP monitor information by ID or name",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Monitor identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Monitor name",
				Optional:            true,
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "SNMP device hostname or IP address",
				Computed:            true,
			},
			"snmp_oid": schema.StringAttribute{
				MarkdownDescription: "SNMP Object Identifier (OID) to query",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorSNMPDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorSNMPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorSNMPDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !validateMonitorDataSourceInput(resp, data.ID, data.Name) {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		d.readByID(ctx, &data, resp)
		return
	}

	d.readByName(ctx, &data, resp)
}

// readByID fetches the SNMP monitor data by its ID from the Uptime Kuma API.
func (d *MonitorSNMPDataSource) readByID(
	ctx context.Context,
	data *MonitorSNMPDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var snmpMonitor monitor.SNMP
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &snmpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read SNMP monitor", err.Error())
		return
	}

	data.Name = types.StringValue(snmpMonitor.Name)
	data.Hostname = types.StringValue(snmpMonitor.Hostname)
	data.SNMPOID = types.StringValue(snmpMonitor.SNMPOID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the SNMP monitor data by its name from the Uptime Kuma API.
func (d *MonitorSNMPDataSource) readByName(
	ctx context.Context,
	data *MonitorSNMPDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "snmp", &resp.Diagnostics)
	if found == nil {
		return
	}

	var snmpMon monitor.SNMP
	err := found.As(&snmpMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(snmpMon.ID)
	data.Hostname = types.StringValue(snmpMon.Hostname)
	data.SNMPOID = types.StringValue(snmpMon.SNMPOID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
