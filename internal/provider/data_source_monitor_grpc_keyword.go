package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorGrpcKeywordDataSource{}

// NewMonitorGrpcKeywordDataSource returns a new instance of the gRPC Keyword monitor data source.
func NewMonitorGrpcKeywordDataSource() datasource.DataSource {
	return &MonitorGrpcKeywordDataSource{}
}

// MonitorGrpcKeywordDataSource manages gRPC Keyword monitor data source operations.
type MonitorGrpcKeywordDataSource struct {
	client *kuma.Client
}

// MonitorGrpcKeywordDataSourceModel describes the data model for gRPC Keyword monitor data source.
type MonitorGrpcKeywordDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorGrpcKeywordDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_grpc_keyword"
}

// Schema returns the schema for the data source.
func (*MonitorGrpcKeywordDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get gRPC Keyword monitor information by ID or name",
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
func (d *MonitorGrpcKeywordDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorGrpcKeywordDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorGrpcKeywordDataSourceModel

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

// readByID fetches the gRPC Keyword monitor data by its ID.
func (d *MonitorGrpcKeywordDataSource) readByID(
	ctx context.Context,
	data *MonitorGrpcKeywordDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var grpcKeywordMonitor monitor.GrpcKeyword
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &grpcKeywordMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read gRPC Keyword monitor", err.Error())
		return
	}

	data.Name = types.StringValue(grpcKeywordMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the gRPC Keyword monitor data by its name.
func (d *MonitorGrpcKeywordDataSource) readByName(
	ctx context.Context,
	data *MonitorGrpcKeywordDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "grpc-keyword", &resp.Diagnostics)
	if found == nil {
		return
	}

	var grpcKeywordMon monitor.GrpcKeyword
	err := found.As(&grpcKeywordMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(grpcKeywordMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
