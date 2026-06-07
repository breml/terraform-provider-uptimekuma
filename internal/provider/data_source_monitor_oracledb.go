package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorOracleDBDataSource{}

// NewMonitorOracleDBDataSource returns a new instance of the OracleDB monitor data source.
func NewMonitorOracleDBDataSource() datasource.DataSource {
	return &MonitorOracleDBDataSource{}
}

// MonitorOracleDBDataSource manages OracleDB monitor data source operations.
type MonitorOracleDBDataSource struct {
	client *kuma.Client
}

// MonitorOracleDBDataSourceModel describes the data model for OracleDB monitor data source.
type MonitorOracleDBDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorOracleDBDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_oracledb"
}

// Schema returns the schema for the data source.
func (*MonitorOracleDBDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get OracleDB monitor information by ID or name",
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
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorOracleDBDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorOracleDBDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorOracleDBDataSourceModel

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

// readByID fetches the OracleDB monitor data by its ID.
func (d *MonitorOracleDBDataSource) readByID(
	ctx context.Context,
	data *MonitorOracleDBDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var oracleDBMonitor monitor.OracleDB
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &oracleDBMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read OracleDB monitor", err.Error())
		return
	}

	data.Name = types.StringValue(oracleDBMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the OracleDB monitor data by its name.
func (d *MonitorOracleDBDataSource) readByName(
	ctx context.Context,
	data *MonitorOracleDBDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "oracledb", &resp.Diagnostics)
	if found == nil {
		return
	}

	var oracleDBMon monitor.OracleDB
	err := found.As(&oracleDBMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(oracleDBMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
