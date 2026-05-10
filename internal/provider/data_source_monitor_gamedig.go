package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorGameDigDataSource{}

// NewMonitorGameDigDataSource returns a new instance of the GameDig monitor data source.
func NewMonitorGameDigDataSource() datasource.DataSource {
	return &MonitorGameDigDataSource{}
}

// MonitorGameDigDataSource manages GameDig monitor data source operations.
type MonitorGameDigDataSource struct {
	client *kuma.Client
}

// MonitorGameDigDataSourceModel describes the data model for GameDig monitor data source.
type MonitorGameDigDataSourceModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Hostname             types.String `tfsdk:"hostname"`
	Port                 types.Int64  `tfsdk:"port"`
	Game                 types.String `tfsdk:"game"`
	GameDigGivenPortOnly types.Bool   `tfsdk:"gamedig_given_port_only"`
}

// Metadata returns the metadata for the data source.
func (*MonitorGameDigDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_gamedig"
}

// Schema returns the schema for the data source.
func (*MonitorGameDigDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get GameDig game server monitor information by ID or name",
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
				MarkdownDescription: "Game server IP address or hostname",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Game server port",
				Computed:            true,
			},
			"game": schema.StringAttribute{
				MarkdownDescription: "Game type identifier (e.g. minecraft, csgo)",
				Computed:            true,
			},
			"gamedig_given_port_only": schema.BoolAttribute{
				MarkdownDescription: "Use only the given port without auto-detection",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorGameDigDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorGameDigDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorGameDigDataSourceModel

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

// readByID fetches the GameDig monitor data by its ID.
func (d *MonitorGameDigDataSource) readByID(
	ctx context.Context,
	data *MonitorGameDigDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var gameDigMonitor monitor.GameDig
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &gameDigMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read GameDig monitor", err.Error())
		return
	}

	data.Name = types.StringValue(gameDigMonitor.Name)
	data.Hostname = types.StringValue(gameDigMonitor.Hostname)
	data.Port = types.Int64Value(int64(gameDigMonitor.Port))
	data.Game = types.StringValue(gameDigMonitor.Game)
	data.GameDigGivenPortOnly = types.BoolValue(gameDigMonitor.GameDigGivenPortOnly)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the GameDig monitor data by its name.
func (d *MonitorGameDigDataSource) readByName(
	ctx context.Context,
	data *MonitorGameDigDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "gamedig", &resp.Diagnostics)
	if found == nil {
		return
	}

	var gameDigMon monitor.GameDig
	err := found.As(&gameDigMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(gameDigMon.ID)
	data.Hostname = types.StringValue(gameDigMon.Hostname)
	data.Port = types.Int64Value(int64(gameDigMon.Port))
	data.Game = types.StringValue(gameDigMon.Game)
	data.GameDigGivenPortOnly = types.BoolValue(gameDigMon.GameDigGivenPortOnly)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
