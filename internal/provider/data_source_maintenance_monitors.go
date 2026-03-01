package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &MaintenanceMonitorsDataSource{}

// NewMaintenanceMonitorsDataSource returns a new instance of the maintenance monitors data source.
func NewMaintenanceMonitorsDataSource() datasource.DataSource {
	return &MaintenanceMonitorsDataSource{}
}

// MaintenanceMonitorsDataSource manages maintenance monitors data source operations.
type MaintenanceMonitorsDataSource struct {
	client *kuma.Client
}

// MaintenanceMonitorsDataSourceModel describes the data model for maintenance monitors data source.
type MaintenanceMonitorsDataSourceModel struct {
	MaintenanceID types.Int64 `tfsdk:"maintenance_id"`
	MonitorIDs    types.List  `tfsdk:"monitor_ids"`
}

// Metadata returns the metadata for the data source.
func (*MaintenanceMonitorsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_monitors"
}

// Schema returns the schema for the data source.
func (*MaintenanceMonitorsDataSource) Schema(
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

// Configure configures the data source with the API client.
func (d *MaintenanceMonitorsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
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
