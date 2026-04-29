package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorTailscalePingDataSource{}

// NewMonitorTailscalePingDataSource returns a new instance of the Tailscale Ping monitor data source.
func NewMonitorTailscalePingDataSource() datasource.DataSource {
	return &MonitorTailscalePingDataSource{}
}

// MonitorTailscalePingDataSource manages Tailscale Ping monitor data source operations.
type MonitorTailscalePingDataSource struct {
	client *kuma.Client
}

// MonitorTailscalePingDataSourceModel describes the data model for Tailscale Ping monitor data source.
type MonitorTailscalePingDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Hostname types.String `tfsdk:"hostname"`
}

// Metadata returns the metadata for the data source.
func (*MonitorTailscalePingDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_tailscale_ping"
}

// Schema returns the schema for the data source.
func (*MonitorTailscalePingDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Tailscale Ping monitor information by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Monitor identifier.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Monitor name.",
				Optional:            true,
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Tailscale hostname or IP address to ping.",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorTailscalePingDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorTailscalePingDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorTailscalePingDataSourceModel

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

// readByID fetches the Tailscale Ping monitor data by its ID.
func (d *MonitorTailscalePingDataSource) readByID(
	ctx context.Context,
	data *MonitorTailscalePingDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var tailscalePingMonitor monitor.TailscalePing
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &tailscalePingMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Tailscale Ping monitor", err.Error())
		return
	}

	data.Name = types.StringValue(tailscalePingMonitor.Name)
	data.Hostname = types.StringValue(tailscalePingMonitor.Hostname)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the Tailscale Ping monitor data by its name.
func (d *MonitorTailscalePingDataSource) readByName(
	ctx context.Context,
	data *MonitorTailscalePingDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "tailscale-ping", &resp.Diagnostics)
	if found == nil {
		return
	}

	var tailscalePingMon monitor.TailscalePing
	err := found.As(&tailscalePingMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(tailscalePingMon.ID)
	data.Hostname = types.StringValue(tailscalePingMon.Hostname)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
