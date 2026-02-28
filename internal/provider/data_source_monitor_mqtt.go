package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorMQTTDataSource{}

// NewMonitorMQTTDataSource returns a new instance of the MQTT monitor data source.
func NewMonitorMQTTDataSource() datasource.DataSource {
	return &MonitorMQTTDataSource{}
}

// MonitorMQTTDataSource manages MQTT monitor data source operations.
type MonitorMQTTDataSource struct {
	client *kuma.Client
}

// MonitorMQTTDataSourceModel describes the data model for MQTT monitor data source.
type MonitorMQTTDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Topic types.String `tfsdk:"topic"`
}

// Metadata returns the metadata for the data source.
func (*MonitorMQTTDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_mqtt"
}

// Schema returns the schema for the data source.
func (*MonitorMQTTDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get MQTT monitor information by ID or name",
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
			"topic": schema.StringAttribute{
				MarkdownDescription: "MQTT topic",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorMQTTDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorMQTTDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorMQTTDataSourceModel

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

// readByID fetches the MQTT monitor data by its ID.
func (d *MonitorMQTTDataSource) readByID(
	ctx context.Context,
	data *MonitorMQTTDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var mqttMonitor monitor.MQTT
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &mqttMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read MQTT monitor", err.Error())
		return
	}

	data.Name = types.StringValue(mqttMonitor.Name)
	data.Topic = types.StringValue(mqttMonitor.MQTTTopic)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the MQTT monitor data by its name.
func (d *MonitorMQTTDataSource) readByName(
	ctx context.Context,
	data *MonitorMQTTDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "mqtt", &resp.Diagnostics)
	if found == nil {
		return
	}

	var mqttMon monitor.MQTT
	err := found.As(&mqttMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(mqttMon.ID)
	data.Topic = types.StringValue(mqttMon.MQTTTopic)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
