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

var _ datasource.DataSource = &MonitorHTTPDataSource{}

// NewMonitorHTTPDataSource returns a new instance of the HTTP monitor data source.
func NewMonitorHTTPDataSource() datasource.DataSource {
	return &MonitorHTTPDataSource{}
}

// MonitorHTTPDataSource manages HTTP monitor data source operations.
type MonitorHTTPDataSource struct {
	client *kuma.Client
}

// MonitorHTTPDataSourceModel describes the data model for HTTP monitor data source.
type MonitorHTTPDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	URL  types.String `tfsdk:"url"`
}

func (_ *MonitorHTTPDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http"
}

func (_ *MonitorHTTPDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get HTTP monitor information by ID or name",
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
		},
	}
}

// Configure configures the HTTP monitor data source with the API client.
func (d *MonitorHTTPDataSource) Configure(
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

func (d *MonitorHTTPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorHTTPDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is provided, use it directly
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		var httpMonitor monitor.HTTP
		err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &httpMonitor)
		if err != nil {
			resp.Diagnostics.AddError("failed to read HTTP monitor", err.Error())
			return
		}

		data.Name = types.StringValue(httpMonitor.Name)
		data.URL = types.StringValue(httpMonitor.URL)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// If name is provided, search for it
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		monitors, err := d.client.GetMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read monitors", err.Error())
			return
		}

		var found *monitor.HTTP
		for _, mon := range monitors {
			if mon.Name != data.Name.ValueString() || mon.Type() != "http" {
				continue
			}

			if found != nil {
				resp.Diagnostics.AddError(
					"Multiple monitors found",
					fmt.Sprintf(
						"Multiple HTTP monitors with name '%s' found. Please use 'id' to specify the monitor uniquely.",
						data.Name.ValueString(),
					),
				)
				return
			}

			var httpMon monitor.HTTP
			err := mon.As(&httpMon)
			if err != nil {
				resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
				return
			}

			found = &httpMon
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"HTTP monitor not found",
				fmt.Sprintf("No HTTP monitor with name '%s' found.", data.Name.ValueString()),
			)
			return
		}

		data.ID = types.Int64Value(found.ID)
		data.URL = types.StringValue(found.URL)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
