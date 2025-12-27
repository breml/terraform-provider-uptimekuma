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

func (d *MonitorGrpcKeywordDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_grpc_keyword"
}

func (d *MonitorGrpcKeywordDataSource) Schema(
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

// Configure configures the gRPC Keyword monitor data source with the API client.
func (d *MonitorGrpcKeywordDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	d.client = client
}

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

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		var grpcKeywordMonitor monitor.GrpcKeyword
		err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &grpcKeywordMonitor)
		if err != nil {
			resp.Diagnostics.AddError("failed to read gRPC Keyword monitor", err.Error())
			return
		}

		data.Name = types.StringValue(grpcKeywordMonitor.Name)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		monitors, err := d.client.GetMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read monitors", err.Error())
			return
		}

		var found *monitor.GrpcKeyword
		for _, mon := range monitors {
			if mon.Name != data.Name.ValueString() || mon.Type() != "grpc-keyword" {
				continue
			}

			if found != nil {
				resp.Diagnostics.AddError(
					"Multiple monitors found",
					fmt.Sprintf(
						"Multiple gRPC Keyword monitors with name '%s' found. Please use 'id' to specify the monitor uniquely.",
						data.Name.ValueString(),
					),
				)
				return
			}

			var grpcKeywordMon monitor.GrpcKeyword
			err := mon.As(&grpcKeywordMon)
			if err != nil {
				resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
				return
			}

			found = &grpcKeywordMon
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"gRPC Keyword monitor not found",
				fmt.Sprintf("No gRPC Keyword monitor with name '%s' found.", data.Name.ValueString()),
			)
			return
		}

		data.ID = types.Int64Value(found.ID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
