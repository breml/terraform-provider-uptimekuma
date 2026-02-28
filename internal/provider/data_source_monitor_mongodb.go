package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorMongoDBDataSource{}

// NewMonitorMongoDBDataSource returns a new instance of the MongoDB monitor data source.
func NewMonitorMongoDBDataSource() datasource.DataSource {
	return &MonitorMongoDBDataSource{}
}

// MonitorMongoDBDataSource manages MongoDB monitor data source operations.
type MonitorMongoDBDataSource struct {
	client *kuma.Client
}

// MonitorMongoDBDataSourceModel describes the data model for MongoDB monitor data source.
type MonitorMongoDBDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorMongoDBDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_mongodb"
}

// Schema returns the schema for the data source.
func (*MonitorMongoDBDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get MongoDB monitor information by ID or name",
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
func (d *MonitorMongoDBDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorMongoDBDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorMongoDBDataSourceModel

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

// readByID fetches the MongoDB monitor data by its ID.
func (d *MonitorMongoDBDataSource) readByID(
	ctx context.Context,
	data *MonitorMongoDBDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var mongoDBMonitor monitor.MongoDB
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &mongoDBMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read MongoDB monitor", err.Error())
		return
	}

	data.Name = types.StringValue(mongoDBMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the MongoDB monitor data by its name.
func (d *MonitorMongoDBDataSource) readByName(
	ctx context.Context,
	data *MonitorMongoDBDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "mongodb", &resp.Diagnostics)
	if found == nil {
		return
	}

	var mongoDBMon monitor.MongoDB
	err := found.As(&mongoDBMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(mongoDBMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
