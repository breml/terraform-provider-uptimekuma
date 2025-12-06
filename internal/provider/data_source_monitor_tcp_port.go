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

var _ datasource.DataSource = &MonitorTCPPortDataSource{}

func NewMonitorTCPPortDataSource() datasource.DataSource {
	return &MonitorTCPPortDataSource{}
}

type MonitorTCPPortDataSource struct {
	client *kuma.Client
}

type MonitorTCPPortDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Hostname types.String `tfsdk:"hostname"`
	Port     types.Int64  `tfsdk:"port"`
}

func (d *MonitorTCPPortDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_tcp_port"
}

func (d *MonitorTCPPortDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get TCP Port monitor information by ID or name",
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
				MarkdownDescription: "Hostname to monitor",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port number",
				Computed:            true,
			},
		},
	}
}

func (d *MonitorTCPPortDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MonitorTCPPortDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorTCPPortDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		var tcpMonitor monitor.TCPPort
		err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &tcpMonitor)
		if err != nil {
			resp.Diagnostics.AddError("failed to read TCP Port monitor", err.Error())
			return
		}
		data.Name = types.StringValue(tcpMonitor.Name)
		data.Hostname = types.StringValue(tcpMonitor.Hostname)
		data.Port = types.Int64Value(int64(tcpMonitor.Port))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		monitors, err := d.client.GetMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read monitors", err.Error())
			return
		}

		var found *monitor.TCPPort
		for _, m := range monitors {
			if m.Name == data.Name.ValueString() && m.Type() == "port" {
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple monitors found",
						fmt.Sprintf("Multiple TCP Port monitors with name '%s' found. Please use 'id' to specify the monitor uniquely.", data.Name.ValueString()),
					)
					return
				}
				var tcpMon monitor.TCPPort
				err := m.As(&tcpMon)
				if err != nil {
					resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
					return
				}
				found = &tcpMon
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"TCP Port monitor not found",
				fmt.Sprintf("No TCP Port monitor with name '%s' found.", data.Name.ValueString()),
			)
			return
		}

		data.ID = types.Int64Value(found.ID)
		data.Hostname = types.StringValue(found.Hostname)
		data.Port = types.Int64Value(int64(found.Port))
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
