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

var _ datasource.DataSource = &MonitorRealBrowserDataSource{}

func NewMonitorRealBrowserDataSource() datasource.DataSource {
	return &MonitorRealBrowserDataSource{}
}

type MonitorRealBrowserDataSource struct {
	client *kuma.Client
}

type MonitorRealBrowserDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *MonitorRealBrowserDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_real_browser"
}

func (d *MonitorRealBrowserDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Real Browser monitor information by ID or name",
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

func (d *MonitorRealBrowserDataSource) Configure(
	ctx context.Context,
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

func (d *MonitorRealBrowserDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorRealBrowserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		var realBrowserMonitor monitor.RealBrowser
		err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &realBrowserMonitor)
		if err != nil {
			resp.Diagnostics.AddError("failed to read Real Browser monitor", err.Error())
			return
		}

		data.Name = types.StringValue(realBrowserMonitor.Name)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		monitors, err := d.client.GetMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read monitors", err.Error())
			return
		}

		var found *monitor.RealBrowser
		for _, m := range monitors {
			if m.Name == data.Name.ValueString() && m.Type() == "real-browser" {
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple monitors found",
						fmt.Sprintf(
							"Multiple Real Browser monitors with name '%s' found. Please use 'id' to specify the monitor uniquely.",
							data.Name.ValueString(),
						),
					)
					return
				}

				var realBrowserMon monitor.RealBrowser
				err := m.As(&realBrowserMon)
				if err != nil {
					resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
					return
				}

				found = &realBrowserMon
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"Real Browser monitor not found",
				fmt.Sprintf("No Real Browser monitor with name '%s' found.", data.Name.ValueString()),
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
