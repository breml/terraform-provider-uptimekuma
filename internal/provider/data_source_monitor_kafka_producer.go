package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorKafkaProducerDataSource{}

// NewMonitorKafkaProducerDataSource returns a new instance of the Kafka Producer monitor data source.
func NewMonitorKafkaProducerDataSource() datasource.DataSource {
	return &MonitorKafkaProducerDataSource{}
}

// MonitorKafkaProducerDataSource manages Kafka Producer monitor data source operations.
type MonitorKafkaProducerDataSource struct {
	client *kuma.Client
}

// MonitorKafkaProducerDataSourceModel describes the data model for Kafka Producer monitor data source.
type MonitorKafkaProducerDataSourceModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Brokers types.List   `tfsdk:"brokers"`
	Topic   types.String `tfsdk:"topic"`
}

// Metadata returns the metadata for the data source.
func (*MonitorKafkaProducerDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_kafka_producer"
}

// Schema returns the schema for the data source.
func (*MonitorKafkaProducerDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Kafka Producer monitor information by ID or name",
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
			"brokers": schema.ListAttribute{
				MarkdownDescription: "List of Kafka broker addresses",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "Kafka topic to publish messages to",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorKafkaProducerDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorKafkaProducerDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorKafkaProducerDataSourceModel

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

// readByID fetches the Kafka Producer monitor data by its ID from the Uptime Kuma API.
func (d *MonitorKafkaProducerDataSource) readByID(
	ctx context.Context,
	data *MonitorKafkaProducerDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var kafkaMonitor monitor.KafkaProducer
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &kafkaMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Kafka Producer monitor", err.Error())
		return
	}

	data.Name = types.StringValue(kafkaMonitor.Name)
	data.Topic = types.StringValue(kafkaMonitor.Topic)

	brokers, d2 := types.ListValueFrom(ctx, types.StringType, kafkaMonitor.Brokers)
	resp.Diagnostics.Append(d2...)
	data.Brokers = brokers

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the Kafka Producer monitor data by its name from the Uptime Kuma API.
func (d *MonitorKafkaProducerDataSource) readByName(
	ctx context.Context,
	data *MonitorKafkaProducerDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "kafka-producer", &resp.Diagnostics)
	if found == nil {
		return
	}

	var kafkaMon monitor.KafkaProducer
	err := found.As(&kafkaMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(kafkaMon.ID)
	data.Topic = types.StringValue(kafkaMon.Topic)

	brokers, d2 := types.ListValueFrom(ctx, types.StringType, kafkaMon.Brokers)
	resp.Diagnostics.Append(d2...)
	data.Brokers = brokers

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
