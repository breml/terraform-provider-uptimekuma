package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorRadiusDataSource{}

// NewMonitorRadiusDataSource returns a new instance of the Radius monitor data source.
func NewMonitorRadiusDataSource() datasource.DataSource {
	return &MonitorRadiusDataSource{}
}

// MonitorRadiusDataSource manages Radius monitor data source operations.
type MonitorRadiusDataSource struct {
	client *kuma.Client
}

// MonitorRadiusDataSourceModel describes the data model for Radius monitor data source.
type MonitorRadiusDataSourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Hostname       types.String `tfsdk:"hostname"`
	RadiusUsername types.String `tfsdk:"radius_username"`
}

// Metadata returns the metadata for the data source.
func (*MonitorRadiusDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_radius"
}

// Schema returns the schema for the data source.
func (*MonitorRadiusDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Radius monitor information by ID or name",
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
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Radius server hostname or IP address",
				Computed:            true,
			},
			"radius_username": schema.StringAttribute{
				MarkdownDescription: "Username for Radius authentication",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorRadiusDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorRadiusDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorRadiusDataSourceModel

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

// readByID fetches the Radius monitor data by its ID from the Uptime Kuma API.
func (d *MonitorRadiusDataSource) readByID(
	ctx context.Context,
	data *MonitorRadiusDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var radiusMonitor monitor.Radius
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &radiusMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Radius monitor", err.Error())
		return
	}

	data.Name = types.StringValue(radiusMonitor.Name)
	data.Hostname = types.StringValue(radiusMonitor.Hostname)
	data.RadiusUsername = types.StringValue(radiusMonitor.Username)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the Radius monitor data by its name from the Uptime Kuma API.
func (d *MonitorRadiusDataSource) readByName(
	ctx context.Context,
	data *MonitorRadiusDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "radius", &resp.Diagnostics)
	if found == nil {
		return
	}

	var radiusMon monitor.Radius
	err := found.As(&radiusMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(radiusMon.ID)
	data.Hostname = types.StringValue(radiusMon.Hostname)
	data.RadiusUsername = types.StringValue(radiusMon.Username)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
