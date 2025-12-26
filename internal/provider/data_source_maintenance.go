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

func NewMaintenanceDataSource() datasource.DataSource {
	return &MaintenanceDataSource{}
}

type MaintenanceDataSource struct {
	client *kuma.Client
}

type MaintenanceDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Title types.String `tfsdk:"title"`
}

func (d *MaintenanceDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_maintenance"
}

func (d *MaintenanceDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
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

func (d *MaintenanceDataSource) Configure(
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

func (d *MaintenanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MaintenanceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		maintenance, err := d.client.GetMaintenance(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("failed to read maintenance", err.Error())
			return
		}

		data.Name = types.StringValue(maintenance.Title)
		data.Title = types.StringValue(maintenance.Title)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		maintenances, err := d.client.GetMaintenances(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read maintenances", err.Error())
			return
		}

		var found *struct {
			ID    int64
			Title string
		}

		for i := range maintenances {
			if maintenances[i].Title == data.Name.ValueString() {
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

				found = &struct {
					ID    int64
					Title string
				}{
					ID:    maintenances[i].ID,
					Title: maintenances[i].Title,
				}
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"Maintenance not found",
				fmt.Sprintf("No maintenance window with title '%s' found.", data.Name.ValueString()),
			)
			return
		}

		data.ID = types.Int64Value(found.ID)
		data.Title = types.StringValue(found.Title)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
