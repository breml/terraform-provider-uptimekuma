package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &MaintenanceMonitorsDataSource{}

func NewMaintenanceMonitorsDataSource() datasource.DataSource {
	return &MaintenanceMonitorsDataSource{}
}

type MaintenanceMonitorsDataSource struct {
	client *kuma.Client
}

type MaintenanceMonitorsDataSourceModel struct {
	MaintenanceID types.Int64 `tfsdk:"maintenance_id"`
	MonitorIDs    types.List  `tfsdk:"monitor_ids"`
}

func (d *MaintenanceMonitorsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_monitors"
}

func (d *MaintenanceMonitorsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get monitors associated with a maintenance window",
		Attributes: map[string]schema.Attribute{
			"maintenance_id": schema.Int64Attribute{
				MarkdownDescription: "Maintenance window ID",
				Required:            true,
			},
			"monitor_ids": schema.ListAttribute{
				MarkdownDescription: "List of monitor IDs associated with the maintenance window",
				Computed:            true,
				ElementType:         types.Int64Type,
			},
		},
	}
}

func (d *MaintenanceMonitorsDataSource) Configure(
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

func (d *MaintenanceMonitorsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MaintenanceMonitorsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitorIDs, err := d.client.GetMonitorMaintenance(ctx, data.MaintenanceID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read maintenance monitors", err.Error())
		return
	}

	listValue, diags := types.ListValueFrom(ctx, types.Int64Type, monitorIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.MonitorIDs = listValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
