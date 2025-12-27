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

// NewMonitorGroupDataSource returns a new instance of the monitor group data source.
func NewMonitorGroupDataSource() datasource.DataSource {
	return &MonitorGroupDataSource{}
}

// MonitorGroupDataSource manages monitor group data source operations.
type MonitorGroupDataSource struct {
	client *kuma.Client
}

// MonitorGroupDataSourceModel describes the data model for monitor group data source.
type MonitorGroupDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorGroupDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_group"
}

// Schema returns the schema for the data source.
func (*MonitorGroupDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
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

// Configure configures the data source with the API client.
func (d *MonitorGroupDataSource) Configure(
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
		for _, mon := range monitors {
			if mon.Name != data.Name.ValueString() || mon.Type() != "group" {
				continue
			}

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
			err := mon.As(&groupMon)
			if err != nil {
				resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
				return
			}

			found = &groupMon
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
