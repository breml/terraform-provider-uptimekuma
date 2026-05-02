package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorRabbitMQDataSource{}

// NewMonitorRabbitMQDataSource returns a new instance of the RabbitMQ monitor data source.
func NewMonitorRabbitMQDataSource() datasource.DataSource {
	return &MonitorRabbitMQDataSource{}
}

// MonitorRabbitMQDataSource manages RabbitMQ monitor data source operations.
type MonitorRabbitMQDataSource struct {
	client *kuma.Client
}

// MonitorRabbitMQDataSourceModel describes the data model for RabbitMQ monitor data source.
type MonitorRabbitMQDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorRabbitMQDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_rabbitmq"
}

// Schema returns the schema for the data source.
func (*MonitorRabbitMQDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get RabbitMQ monitor information by ID or name",
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
func (d *MonitorRabbitMQDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorRabbitMQDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorRabbitMQDataSourceModel

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

// readByID fetches the RabbitMQ monitor data by its ID.
func (d *MonitorRabbitMQDataSource) readByID(
	ctx context.Context,
	data *MonitorRabbitMQDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var rabbitMQMonitor monitor.RabbitMQ
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &rabbitMQMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read RabbitMQ monitor", err.Error())
		return
	}

	data.Name = types.StringValue(rabbitMQMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the RabbitMQ monitor data by its name.
func (d *MonitorRabbitMQDataSource) readByName(
	ctx context.Context,
	data *MonitorRabbitMQDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "rabbitmq", &resp.Diagnostics)
	if found == nil {
		return
	}

	var rabbitMQMon monitor.RabbitMQ
	err := found.As(&rabbitMQMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(rabbitMQMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
