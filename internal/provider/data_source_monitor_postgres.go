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

var _ datasource.DataSource = &MonitorPostgresDataSource{}

func NewMonitorPostgresDataSource() datasource.DataSource {
	return &MonitorPostgresDataSource{}
}

type MonitorPostgresDataSource struct {
	client *kuma.Client
}

type MonitorPostgresDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *MonitorPostgresDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_postgres"
}

func (d *MonitorPostgresDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get PostgreSQL monitor information by ID or name",
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
		},
	}
}

func (d *MonitorPostgresDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MonitorPostgresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorPostgresDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		var postgresMonitor monitor.Postgres
		err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &postgresMonitor)
		if err != nil {
			resp.Diagnostics.AddError("failed to read PostgreSQL monitor", err.Error())
			return
		}

		data.Name = types.StringValue(postgresMonitor.Name)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		monitors, err := d.client.GetMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to read monitors", err.Error())
			return
		}

		var found *monitor.Postgres
		for _, m := range monitors {
			if m.Name == data.Name.ValueString() && m.Type() == "postgres" {
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple monitors found",
						fmt.Sprintf("Multiple PostgreSQL monitors with name '%s' found. Please use 'id' to specify the monitor uniquely.", data.Name.ValueString()),
					)
					return
				}

				var postgresMon monitor.Postgres
				err := m.As(&postgresMon)
				if err != nil {
					resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
					return
				}

				found = &postgresMon
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"PostgreSQL monitor not found",
				fmt.Sprintf("No PostgreSQL monitor with name '%s' found.", data.Name.ValueString()),
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
