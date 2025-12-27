package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &MaintenanceStatusPagesDataSource{}

func NewMaintenanceStatusPagesDataSource() datasource.DataSource {
	return &MaintenanceStatusPagesDataSource{}
}

type MaintenanceStatusPagesDataSource struct {
	client *kuma.Client
}

type MaintenanceStatusPagesDataSourceModel struct {
	MaintenanceID types.Int64 `tfsdk:"maintenance_id"`
	StatusPageIDs types.List  `tfsdk:"status_page_ids"`
}

func (d *MaintenanceStatusPagesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_status_pages"
}

func (d *MaintenanceStatusPagesDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get status pages associated with a maintenance window",
		Attributes: map[string]schema.Attribute{
			"maintenance_id": schema.Int64Attribute{
				MarkdownDescription: "Maintenance window ID",
				Required:            true,
			},
			"status_page_ids": schema.ListAttribute{
				MarkdownDescription: "List of status page IDs associated with the maintenance window",
				Computed:            true,
				ElementType:         types.Int64Type,
			},
		},
	}
}

func (d *MaintenanceStatusPagesDataSource) Configure(
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

func (d *MaintenanceStatusPagesDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MaintenanceStatusPagesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageIDs, err := d.client.GetMaintenanceStatusPage(ctx, data.MaintenanceID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to read maintenance status pages", err.Error())
		return
	}

	listValue, diags := types.ListValueFrom(ctx, types.Int64Type, statusPageIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.StatusPageIDs = listValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
