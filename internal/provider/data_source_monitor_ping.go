package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorPingDataSource{}

// NewMonitorPingDataSource returns a new instance of the PING monitor data source.
func NewMonitorPingDataSource() datasource.DataSource {
	return &MonitorPingDataSource{}
}

// MonitorPingDataSource manages PING monitor data source operations.
type MonitorPingDataSource struct {
	client *kuma.Client
}

// MonitorPingDataSourceModel describes the data model for PING monitor data source.
type MonitorPingDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Hostname types.String `tfsdk:"hostname"`
}

// Metadata returns the metadata for the data source.
func (*MonitorPingDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_ping"
}

// Schema returns the schema for the data source.
func (*MonitorPingDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get PING monitor information by ID or name",
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
				MarkdownDescription: "Hostname to ping",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorPingDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorPingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorPingDataSourceModel

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

// readByID fetches the Ping monitor data by its ID.
func (d *MonitorPingDataSource) readByID(
	ctx context.Context,
	data *MonitorPingDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var pingMonitor monitor.Ping
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &pingMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read PING monitor", err.Error())
		return
	}

	data.Name = types.StringValue(pingMonitor.Name)
	data.Hostname = types.StringValue(pingMonitor.Hostname)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the Ping monitor data by its name.
func (d *MonitorPingDataSource) readByName(
	ctx context.Context,
	data *MonitorPingDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "ping", &resp.Diagnostics)
	if found == nil {
		return
	}

	var pingMon monitor.Ping
	err := found.As(&pingMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(pingMon.ID)
	data.Hostname = types.StringValue(pingMon.Hostname)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
