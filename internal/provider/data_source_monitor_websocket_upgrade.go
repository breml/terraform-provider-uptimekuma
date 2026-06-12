package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorWebsocketUpgradeDataSource{}

// NewMonitorWebsocketUpgradeDataSource returns a new instance of the Websocket Upgrade monitor data source.
func NewMonitorWebsocketUpgradeDataSource() datasource.DataSource {
	return &MonitorWebsocketUpgradeDataSource{}
}

// MonitorWebsocketUpgradeDataSource manages Websocket Upgrade monitor data source operations.
type MonitorWebsocketUpgradeDataSource struct {
	client *kuma.Client
}

// MonitorWebsocketUpgradeDataSourceModel describes the data model for Websocket Upgrade monitor data source.
type MonitorWebsocketUpgradeDataSourceModel struct {
	ID                               types.Int64  `tfsdk:"id"`
	Name                             types.String `tfsdk:"name"`
	URL                              types.String `tfsdk:"url"`
	WSIgnoreSecWebsocketAcceptHeader types.Bool   `tfsdk:"ws_ignore_sec_websocket_accept_header"`
	WSSubprotocol                    types.String `tfsdk:"ws_subprotocol"`
}

// Metadata returns the metadata for the data source.
func (*MonitorWebsocketUpgradeDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_websocket_upgrade"
}

// Schema returns the schema for the data source.
func (*MonitorWebsocketUpgradeDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Websocket Upgrade monitor information by ID or name",
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
			"url": schema.StringAttribute{
				MarkdownDescription: "URL to monitor",
				Computed:            true,
			},
			"ws_ignore_sec_websocket_accept_header": schema.BoolAttribute{
				MarkdownDescription: "Skip verification of the `Sec-WebSocket-Accept` response header " +
					"during the WebSocket handshake.",
				Computed: true,
			},
			"ws_subprotocol": schema.StringAttribute{
				MarkdownDescription: "Requested `Sec-WebSocket-Protocol` value sent during the " +
					"WebSocket handshake.",
				Computed: true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorWebsocketUpgradeDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorWebsocketUpgradeDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorWebsocketUpgradeDataSourceModel

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

// readByID fetches the Websocket Upgrade monitor data by its ID.
func (d *MonitorWebsocketUpgradeDataSource) readByID(
	ctx context.Context,
	data *MonitorWebsocketUpgradeDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var wsMon monitor.WebsocketUpgrade
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &wsMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Websocket Upgrade monitor", err.Error())
		return
	}

	if actual := wsMon.Base.Type(); actual != "" && actual != wsMon.Type() {
		resp.Diagnostics.AddError(
			"Monitor type mismatch",
			fmt.Sprintf(
				"Monitor ID %d has type %q, expected %q.",
				data.ID.ValueInt64(), actual, wsMon.Type(),
			),
		)
		return
	}

	data.Name = types.StringValue(wsMon.Name)
	data.URL = types.StringValue(wsMon.URL)
	data.WSIgnoreSecWebsocketAcceptHeader = types.BoolValue(wsMon.IgnoreSecWebsocketAcceptHeader)
	data.WSSubprotocol = types.StringValue(wsMon.Subprotocol)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// readByName fetches the Websocket Upgrade monitor data by its name.
func (d *MonitorWebsocketUpgradeDataSource) readByName(
	ctx context.Context,
	data *MonitorWebsocketUpgradeDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "websocket-upgrade", &resp.Diagnostics)
	if found == nil {
		return
	}

	var wsMon monitor.WebsocketUpgrade
	err := found.As(&wsMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(wsMon.ID)
	data.URL = types.StringValue(wsMon.URL)
	data.WSIgnoreSecWebsocketAcceptHeader = types.BoolValue(wsMon.IgnoreSecWebsocketAcceptHeader)
	data.WSSubprotocol = types.StringValue(wsMon.Subprotocol)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
