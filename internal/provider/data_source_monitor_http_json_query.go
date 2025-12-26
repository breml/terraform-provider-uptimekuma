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

func NewMonitorHTTPJSONQueryDataSource() datasource.DataSource {
	return &MonitorHTTPJSONQueryDataSource{}
}

type MonitorHTTPJSONQueryDataSource struct {
	client *kuma.Client
}

type MonitorHTTPJSONQueryDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *MonitorHTTPJSONQueryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http_json_query"
}

func (d *MonitorHTTPJSONQueryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *MonitorHTTPJSONQueryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *kuma.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *MonitorHTTPJSONQueryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorHTTPJSONQueryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		var httpJSONMonitor monitor.HTTPJSONQuery
		err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &httpJSONMonitor)
		if err != nil {
			resp.Diagnostics.AddError("failed to read HTTP JSON Query monitor", err.Error())
			return
		}

		data.Name = types.StringValue(httpJSONMonitor.Name)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		monitors, err := d.client.GetMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read monitors", err.Error())
			return
		}

		var found *monitor.HTTPJSONQuery
		for _, m := range monitors {
			if m.Name == data.Name.ValueString() && m.Type() == "json-query" {
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple monitors found",
						fmt.Sprintf("Multiple HTTP JSON Query monitors with name '%s' found. Please use 'id' to specify the monitor uniquely.", data.Name.ValueString()),
					)
					return
				}

				var httpJSONMon monitor.HTTPJSONQuery
				err := m.As(&httpJSONMon)
				if err != nil {
					resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
					return
				}

				found = &httpJSONMon
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"HTTP JSON Query monitor not found",
				fmt.Sprintf("No HTTP JSON Query monitor with name '%s' found.", data.Name.ValueString()),
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
