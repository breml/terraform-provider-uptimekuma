package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &MaintenanceDataSource{}

// NewMaintenanceDataSource returns a new instance of the maintenance data source.
func NewMaintenanceDataSource() datasource.DataSource {
	return &MaintenanceDataSource{}
}

// MaintenanceDataSource manages maintenance data source operations.
type MaintenanceDataSource struct {
	client *kuma.Client
}

// MaintenanceDataSourceModel describes the data model for maintenance data source.
type MaintenanceDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Title types.String `tfsdk:"title"`
}

// Metadata returns the metadata for the data source.
func (*MaintenanceDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenance"
}

// Schema returns the schema for the data source.
func (*MaintenanceDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get maintenance window information by ID or name",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Maintenance identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Maintenance name (title)",
				Optional:            true,
				Computed:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Maintenance title",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MaintenanceDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MaintenanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MaintenanceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Attempt to read by ID if provided.
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		maintenance, err := d.client.GetMaintenance(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("failed to read maintenance", err.Error())
			return
		}

		// Populate name and title from API response.
		data.Name = types.StringValue(maintenance.Title)
		data.Title = types.StringValue(maintenance.Title)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Attempt to read by name if ID not provided.
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		maintenances, err := d.client.GetMaintenances(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read maintenances", err.Error())
			return
		}

		// Search for matching maintenance by title.
		var found *struct {
			ID    int64
			Title string
		}

		for i := range maintenances {
			if maintenances[i].Title == data.Name.ValueString() {
				// Error if multiple matches found.
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple maintenances found",
						fmt.Sprintf(
							"Multiple maintenance windows with title '%s' found. Please use 'id' to specify the maintenance uniquely.",
							data.Name.ValueString(),
						),
					)
					return
				}

				// Store matched maintenance record.
				found = &struct {
					ID    int64
					Title string
				}{
					ID:    maintenances[i].ID,
					Title: maintenances[i].Title,
				}
			}
		}

		// Error if no matching maintenance found.
		if found == nil {
			resp.Diagnostics.AddError(
				"Maintenance not found",
				fmt.Sprintf("No maintenance window with title '%s' found.", data.Name.ValueString()),
			)
			return
		}

		// Populate ID and title from matched result.
		data.ID = types.Int64Value(found.ID)
		data.Title = types.StringValue(found.Title)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Error if neither ID nor name provided.
	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
