package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &MaintenanceStatusPagesDataSource{}

// NewMaintenanceStatusPagesDataSource returns a new instance of the maintenance status pages data source.
func NewMaintenanceStatusPagesDataSource() datasource.DataSource {
	return &MaintenanceStatusPagesDataSource{}
}

// MaintenanceStatusPagesDataSource manages maintenance status pages data source operations.
type MaintenanceStatusPagesDataSource struct {
	client *kuma.Client
}

// MaintenanceStatusPagesDataSourceModel describes the data model for maintenance status pages data source.
type MaintenanceStatusPagesDataSourceModel struct {
	MaintenanceID types.Int64 `tfsdk:"maintenance_id"`
	StatusPageIDs types.List  `tfsdk:"status_page_ids"`
}

// Metadata returns the metadata for the data source.
func (*MaintenanceStatusPagesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_status_pages"
}

// Schema returns the schema for the data source.
func (*MaintenanceStatusPagesDataSource) Schema(
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

// Configure configures the data source with the API client.
func (d *MaintenanceStatusPagesDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
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
