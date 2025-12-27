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

func (d *MonitorHTTPKeywordDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http_keyword"
}

func (d *MonitorHTTPKeywordDataSource) Schema(
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

// Configure configures the HTTP Keyword monitor data source with the API client.
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

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		var httpKeywordMonitor monitor.HTTPKeyword
		err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &httpKeywordMonitor)
		if err != nil {
			resp.Diagnostics.AddError("failed to read HTTP Keyword monitor", err.Error())
			return
		}

		data.Name = types.StringValue(httpKeywordMonitor.Name)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		monitors, err := d.client.GetMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read monitors", err.Error())
			return
		}

		var found *monitor.HTTPKeyword
		for _, mon := range monitors {
			if mon.Name != data.Name.ValueString() || mon.Type() != "keyword" {
				continue
			}

			if found != nil {
				resp.Diagnostics.AddError(
					"Multiple monitors found",
					fmt.Sprintf(
						"Multiple HTTP Keyword monitors with name '%s' found. Please use 'id' to specify the monitor uniquely.",
						data.Name.ValueString(),
					),
				)
				return
			}

			var httpKeywordMon monitor.HTTPKeyword
			err := mon.As(&httpKeywordMon)
			if err != nil {
				resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
				return
			}

			found = &httpKeywordMon
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"HTTP Keyword monitor not found",
				fmt.Sprintf("No HTTP Keyword monitor with name '%s' found.", data.Name.ValueString()),
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
