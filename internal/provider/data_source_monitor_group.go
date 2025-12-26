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

var _ datasource.DataSource = &MonitorGroupDataSource{}

func NewMonitorGroupDataSource() datasource.DataSource {
	return &MonitorGroupDataSource{}
}

type MonitorGroupDataSource struct {
	client *kuma.Client
}

type MonitorGroupDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *MonitorGroupDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_group"
}

func (d *MonitorGroupDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get monitor group information by ID or name",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Monitor group identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Monitor group name",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (d *MonitorGroupDataSource) Configure(
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

func (d *MonitorGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorGroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		var groupMonitor monitor.Group
		err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &groupMonitor)
		if err != nil {
			resp.Diagnostics.AddError("failed to read monitor group", err.Error())
			return
		}

		data.Name = types.StringValue(groupMonitor.Name)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		monitors, err := d.client.GetMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read monitors", err.Error())
			return
		}

		var found *monitor.Group
		for _, m := range monitors {
			if m.Name == data.Name.ValueString() && m.Type() == "group" {
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple groups found",
						fmt.Sprintf(
							"Multiple monitor groups with name '%s' found. Please use 'id' to specify the group uniquely.",
							data.Name.ValueString(),
						),
					)
					return
				}

				var groupMon monitor.Group
				err := m.As(&groupMon)
				if err != nil {
					resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
					return
				}

				found = &groupMon
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"Monitor group not found",
				fmt.Sprintf("No monitor group with name '%s' found.", data.Name.ValueString()),
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
