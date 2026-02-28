package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorRedisDataSource{}

// NewMonitorRedisDataSource returns a new instance of the Redis monitor data source.
func NewMonitorRedisDataSource() datasource.DataSource {
	return &MonitorRedisDataSource{}
}

// MonitorRedisDataSource manages Redis monitor data source operations.
type MonitorRedisDataSource struct {
	client *kuma.Client
}

// MonitorRedisDataSourceModel describes the data model for Redis monitor data source.
type MonitorRedisDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorRedisDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_redis"
}

// Schema returns the schema for the data source.
func (*MonitorRedisDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Redis monitor information by ID or name",
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
func (d *MonitorRedisDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorRedisDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorRedisDataSourceModel

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

// readByID fetches the Redis monitor data by its ID.
func (d *MonitorRedisDataSource) readByID(
	ctx context.Context,
	data *MonitorRedisDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var redisMonitor monitor.Redis
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &redisMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Redis monitor", err.Error())
		return
	}

	data.Name = types.StringValue(redisMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the Redis monitor data by its name.
func (d *MonitorRedisDataSource) readByName(
	ctx context.Context,
	data *MonitorRedisDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "redis", &resp.Diagnostics)
	if found == nil {
		return
	}

	var redisMon monitor.Redis
	err := found.As(&redisMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(redisMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
