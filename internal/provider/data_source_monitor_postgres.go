package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorPostgresDataSource{}

// NewMonitorPostgresDataSource returns a new instance of the PostgreSQL monitor data source.
func NewMonitorPostgresDataSource() datasource.DataSource {
	return &MonitorPostgresDataSource{}
}

// MonitorPostgresDataSource manages PostgreSQL monitor data source operations.
type MonitorPostgresDataSource struct {
	client *kuma.Client
}

// MonitorPostgresDataSourceModel describes the data model for PostgreSQL monitor data source.
type MonitorPostgresDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorPostgresDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_postgres"
}

// Schema returns the schema for the data source.
func (*MonitorPostgresDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
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

// Configure configures the data source with the API client.
func (d *MonitorPostgresDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorPostgresDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorPostgresDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !validateMonitorDataSourceInput(resp, data.ID, data.Name) {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		d.readByID(ctx, &data, resp)
		return
	}

	d.readByName(ctx, &data, resp)
}

// readByID fetches the PostgreSQL monitor data by its ID.
func (d *MonitorPostgresDataSource) readByID(
	ctx context.Context,
	data *MonitorPostgresDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var postgresMonitor monitor.Postgres
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &postgresMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read PostgreSQL monitor", err.Error())
		return
	}

	data.Name = types.StringValue(postgresMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the PostgreSQL monitor data by its name.
func (d *MonitorPostgresDataSource) readByName(
	ctx context.Context,
	data *MonitorPostgresDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "postgres", &resp.Diagnostics)
	if found == nil {
		return
	}

	var postgresMon monitor.Postgres
	err := found.As(&postgresMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(postgresMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
