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

var _ datasource.DataSource = &MonitorHTTPJSONQueryDataSource{}

// NewMonitorHTTPJSONQueryDataSource returns a new instance of the HTTP JSON Query monitor data source.
func NewMonitorHTTPJSONQueryDataSource() datasource.DataSource {
	return &MonitorHTTPJSONQueryDataSource{}
}

// MonitorHTTPJSONQueryDataSource manages HTTP JSON Query monitor data source operations.
type MonitorHTTPJSONQueryDataSource struct {
	client *kuma.Client
}

// MonitorHTTPJSONQueryDataSourceModel describes the data model for HTTP JSON Query monitor data source.
type MonitorHTTPJSONQueryDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorHTTPJSONQueryDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http_json_query"
}

// Schema returns the schema for the data source.
func (*MonitorHTTPJSONQueryDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get HTTP JSON Query monitor information by ID or name",
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
func (d *MonitorHTTPJSONQueryDataSource) Configure(
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
func (d *MonitorHTTPJSONQueryDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorHTTPJSONQueryDataSourceModel

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

// readByID fetches the HTTP JSON Query monitor data by its ID.
func (d *MonitorHTTPJSONQueryDataSource) readByID(
	ctx context.Context,
	data *MonitorHTTPJSONQueryDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var httpJSONMonitor monitor.HTTPJSONQuery
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &httpJSONMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read HTTP JSON Query monitor", err.Error())
		return
	}

	data.Name = types.StringValue(httpJSONMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the HTTP JSON Query monitor data by its name.
func (d *MonitorHTTPJSONQueryDataSource) readByName(
	ctx context.Context,
	data *MonitorHTTPJSONQueryDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "json-query", &resp.Diagnostics)
	if found == nil {
		return
	}

	var httpJSONMon monitor.HTTPJSONQuery
	err := found.As(&httpJSONMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(httpJSONMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
