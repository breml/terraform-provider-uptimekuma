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

var _ datasource.DataSource = &MonitorDNSDataSource{}

func NewMonitorDNSDataSource() datasource.DataSource {
	return &MonitorDNSDataSource{}
}

type MonitorDNSDataSource struct {
	client *kuma.Client
}

type MonitorDNSDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Hostname types.String `tfsdk:"hostname"`
}

func (d *MonitorDNSDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_dns"
}

func (d *MonitorDNSDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get DNS monitor information by ID or name",
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
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname to resolve",
				Computed:            true,
			},
		},
	}
}

func (d *MonitorDNSDataSource) Configure(
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

func (d *MonitorDNSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorDNSDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		var dnsMonitor monitor.DNS
		err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &dnsMonitor)
		if err != nil {
			resp.Diagnostics.AddError("failed to read DNS monitor", err.Error())
			return
		}

		data.Name = types.StringValue(dnsMonitor.Name)
		data.Hostname = types.StringValue(dnsMonitor.Hostname)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		monitors, err := d.client.GetMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read monitors", err.Error())
			return
		}

		var found *monitor.DNS
		for _, m := range monitors {
			if m.Name == data.Name.ValueString() && m.Type() == "dns" {
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple monitors found",
						fmt.Sprintf(
							"Multiple DNS monitors with name '%s' found. Please use 'id' to specify the monitor uniquely.",
							data.Name.ValueString(),
						),
					)
					return
				}

				var dnsMon monitor.DNS
				err := m.As(&dnsMon)
				if err != nil {
					resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
					return
				}

				found = &dnsMon
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"DNS monitor not found",
				fmt.Sprintf("No DNS monitor with name '%s' found.", data.Name.ValueString()),
			)
			return
		}

		data.ID = types.Int64Value(found.ID)
		data.Hostname = types.StringValue(found.Hostname)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
