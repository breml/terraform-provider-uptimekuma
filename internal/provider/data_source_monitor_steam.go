package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorSteamDataSource{}

// NewMonitorSteamDataSource returns a new instance of the Steam monitor data source.
func NewMonitorSteamDataSource() datasource.DataSource {
	return &MonitorSteamDataSource{}
}

// MonitorSteamDataSource manages Steam monitor data source operations.
type MonitorSteamDataSource struct {
	client *kuma.Client
}

// MonitorSteamDataSourceModel describes the data model for Steam monitor data source.
type MonitorSteamDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Hostname types.String `tfsdk:"hostname"`
	Port     types.Int64  `tfsdk:"port"`
}

// Metadata returns the metadata for the data source.
func (*MonitorSteamDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_steam"
}

// Schema returns the schema for the data source.
func (*MonitorSteamDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Steam game server monitor information by ID or name",
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
				MarkdownDescription: "Steam game server IP address or hostname",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Steam game server port",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorSteamDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorSteamDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorSteamDataSourceModel

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

// readByID fetches the Steam monitor data by its ID.
func (d *MonitorSteamDataSource) readByID(
	ctx context.Context,
	data *MonitorSteamDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var steamMonitor monitor.Steam
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &steamMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Steam monitor", err.Error())
		return
	}

	data.Name = types.StringValue(steamMonitor.Name)
	data.Hostname = types.StringValue(steamMonitor.Hostname)
	data.Port = types.Int64Value(int64(steamMonitor.Port))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the Steam monitor data by its name.
func (d *MonitorSteamDataSource) readByName(
	ctx context.Context,
	data *MonitorSteamDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "steam", &resp.Diagnostics)
	if found == nil {
		return
	}

	var steamMon monitor.Steam
	err := found.As(&steamMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(steamMon.ID)
	data.Hostname = types.StringValue(steamMon.Hostname)
	data.Port = types.Int64Value(int64(steamMon.Port))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
