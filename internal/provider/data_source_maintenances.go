package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &MaintenancesDataSource{}

// NewMaintenancesDataSource returns a new instance of the maintenances data source.
func NewMaintenancesDataSource() datasource.DataSource {
	return &MaintenancesDataSource{}
}

// MaintenancesDataSource manages maintenances data source operations.
type MaintenancesDataSource struct {
	client *kuma.Client
}

// MaintenancesDataSourceModel describes the data model for maintenances data source.
type MaintenancesDataSourceModel struct {
	Maintenances types.List `tfsdk:"maintenances"`
}

// MaintenanceDataModel describes the data model for maintenance data.
type MaintenanceDataModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Title            types.String `tfsdk:"title"`
	Description      types.String `tfsdk:"description"`
	Strategy         types.String `tfsdk:"strategy"`
	Active           types.Bool   `tfsdk:"active"`
	Status           types.String `tfsdk:"status"`
	TimezoneResolved types.String `tfsdk:"timezone_resolved"`
	TimezoneOffset   types.String `tfsdk:"timezone_offset"`
}

func (d *MaintenancesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenances"
}

func (d *MaintenancesDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all maintenance windows",
		Attributes: map[string]schema.Attribute{
			"maintenances": schema.ListNestedAttribute{
				MarkdownDescription: "List of maintenance windows",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Maintenance window ID",
							Computed:            true,
						},
						"title": schema.StringAttribute{
							MarkdownDescription: "Name of the maintenance window",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Additional details about the maintenance",
							Computed:            true,
						},
						"strategy": schema.StringAttribute{
							MarkdownDescription: "Scheduling pattern",
							Computed:            true,
						},
						"active": schema.BoolAttribute{
							MarkdownDescription: "Whether the maintenance window is active",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "Current status",
							Computed:            true,
						},
						"timezone_resolved": schema.StringAttribute{
							MarkdownDescription: "Resolved IANA timezone",
							Computed:            true,
						},
						"timezone_offset": schema.StringAttribute{
							MarkdownDescription: "Timezone offset from UTC",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure configures the maintenances data source with the API client.
func (d *MaintenancesDataSource) Configure(
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
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	d.client = client
}

func (d *MaintenancesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MaintenancesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	maintenances, err := d.client.GetMaintenances(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to read maintenances", err.Error())
		return
	}

	maintenanceAttrTypes := map[string]attr.Type{
		"id":                types.Int64Type,
		"title":             types.StringType,
		"description":       types.StringType,
		"strategy":          types.StringType,
		"active":            types.BoolType,
		"status":            types.StringType,
		"timezone_resolved": types.StringType,
		"timezone_offset":   types.StringType,
	}

	maintenanceList := make([]attr.Value, len(maintenances))
	for i, m := range maintenances {
		status := m.Status
		if status == "" {
			status = "unknown"
		}

		objValue, diags := types.ObjectValue(maintenanceAttrTypes, map[string]attr.Value{
			"id":                types.Int64Value(m.ID),
			"title":             types.StringValue(m.Title),
			"description":       types.StringValue(m.Description),
			"strategy":          types.StringValue(m.Strategy),
			"active":            types.BoolValue(m.Active),
			"status":            types.StringValue(status),
			"timezone_resolved": types.StringValue(m.Timezone),
			"timezone_offset":   types.StringValue(m.TimezoneOffset),
		})
		resp.Diagnostics.Append(diags...)
		maintenanceList[i] = objValue
	}

	listValue, diags := types.ListValue(types.ObjectType{AttrTypes: maintenanceAttrTypes}, maintenanceList)
	resp.Diagnostics.Append(diags...)
	data.Maintenances = listValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
