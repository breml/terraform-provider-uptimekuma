package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorSQLServerDataSource{}

// NewMonitorSQLServerDataSource returns a new instance of the SQL Server monitor data source.
func NewMonitorSQLServerDataSource() datasource.DataSource {
	return &MonitorSQLServerDataSource{}
}

// MonitorSQLServerDataSource manages SQL Server monitor data source operations.
type MonitorSQLServerDataSource struct {
	client *kuma.Client
}

// MonitorSQLServerDataSourceModel describes the data model for SQL Server monitor data source.
type MonitorSQLServerDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorSQLServerDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_sqlserver"
}

// Schema returns the schema for the data source.
func (*MonitorSQLServerDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get SQL Server monitor information by ID or name",
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
func (d *MonitorSQLServerDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorSQLServerDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorSQLServerDataSourceModel

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

// readByID fetches the SQL Server monitor data by its ID.
func (d *MonitorSQLServerDataSource) readByID(
	ctx context.Context,
	data *MonitorSQLServerDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var sqlserverMonitor monitor.SQLServer
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &sqlserverMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read SQL Server monitor", err.Error())
		return
	}

	data.Name = types.StringValue(sqlserverMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the SQL Server monitor data by its name.
func (d *MonitorSQLServerDataSource) readByName(
	ctx context.Context,
	data *MonitorSQLServerDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "sqlserver", &resp.Diagnostics)
	if found == nil {
		return
	}

	var sqlserverMon monitor.SQLServer
	err := found.As(&sqlserverMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(sqlserverMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
