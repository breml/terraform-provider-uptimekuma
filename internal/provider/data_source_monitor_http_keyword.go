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

var _ datasource.DataSource = &MonitorHTTPKeywordDataSource{}

// NewMonitorHTTPKeywordDataSource returns a new instance of the HTTP Keyword monitor data source.
func NewMonitorHTTPKeywordDataSource() datasource.DataSource {
	return &MonitorHTTPKeywordDataSource{}
}

// MonitorHTTPKeywordDataSource manages HTTP Keyword monitor data source operations.
type MonitorHTTPKeywordDataSource struct {
	client *kuma.Client
}

// MonitorHTTPKeywordDataSourceModel describes the data model for HTTP Keyword monitor data source.
type MonitorHTTPKeywordDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorHTTPKeywordDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http_keyword"
}

// Schema returns the schema for the data source.
func (*MonitorHTTPKeywordDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get HTTP Keyword monitor information by ID or name",
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
func (d *MonitorHTTPKeywordDataSource) Configure(
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

// Read reads the current state of the data source.
func (d *MonitorHTTPKeywordDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorHTTPKeywordDataSourceModel

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

// readByID fetches the HTTP Keyword monitor data by its ID.
func (d *MonitorHTTPKeywordDataSource) readByID(
	ctx context.Context,
	data *MonitorHTTPKeywordDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var httpKeywordMonitor monitor.HTTPKeyword
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &httpKeywordMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read HTTP Keyword monitor", err.Error())
		return
	}

	data.Name = types.StringValue(httpKeywordMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the HTTP Keyword monitor data by its name.
func (d *MonitorHTTPKeywordDataSource) readByName(
	ctx context.Context,
	data *MonitorHTTPKeywordDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "keyword", &resp.Diagnostics)
	if found == nil {
		return
	}

	var httpKeywordMon monitor.HTTPKeyword
	err := found.As(&httpKeywordMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(httpKeywordMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
