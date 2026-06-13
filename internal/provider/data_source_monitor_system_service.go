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

var _ datasource.DataSource = &MonitorSystemServiceDataSource{}

// NewMonitorSystemServiceDataSource returns a new instance of the System Service monitor data source.
func NewMonitorSystemServiceDataSource() datasource.DataSource {
	return &MonitorSystemServiceDataSource{}
}

// MonitorSystemServiceDataSource manages System Service monitor data source operations.
type MonitorSystemServiceDataSource struct {
	client *kuma.Client
}

// MonitorSystemServiceDataSourceModel describes the data model for System Service monitor data source.
type MonitorSystemServiceDataSourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	SystemServiceName types.String `tfsdk:"system_service_name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorSystemServiceDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_system_service"
}

// Schema returns the schema for the data source.
func (*MonitorSystemServiceDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get System Service monitor information by ID or name",
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
			"system_service_name": schema.StringAttribute{
				MarkdownDescription: "Name of the systemd unit (Linux) or SCM service (Windows) being monitored",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorSystemServiceDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorSystemServiceDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorSystemServiceDataSourceModel

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

// readByID fetches the System Service monitor data by its ID.
func (d *MonitorSystemServiceDataSource) readByID(
	ctx context.Context,
	data *MonitorSystemServiceDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var systemServiceMon monitor.SystemService
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &systemServiceMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to read System Service monitor", err.Error())
		return
	}

	if actual := systemServiceMon.Base.Type(); actual != "" && actual != systemServiceMon.Type() {
		resp.Diagnostics.AddError(
			"Monitor type mismatch",
			fmt.Sprintf(
				"Monitor ID %d has type %q, expected %q.",
				data.ID.ValueInt64(), actual, systemServiceMon.Type(),
			),
		)
		return
	}

	data.Name = types.StringValue(systemServiceMon.Name)
	data.SystemServiceName = types.StringValue(systemServiceMon.SystemServiceName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// readByName fetches the System Service monitor data by its name.
func (d *MonitorSystemServiceDataSource) readByName(
	ctx context.Context,
	data *MonitorSystemServiceDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "system-service", &resp.Diagnostics)
	if found == nil {
		return
	}

	var systemServiceMon monitor.SystemService
	err := found.As(&systemServiceMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	if actual := systemServiceMon.Base.Type(); actual != "" && actual != systemServiceMon.Type() {
		resp.Diagnostics.AddError(
			"Monitor type mismatch",
			fmt.Sprintf(
				"Monitor %q has type %q, expected %q.",
				data.Name.ValueString(), actual, systemServiceMon.Type(),
			),
		)
		return
	}

	data.ID = types.Int64Value(systemServiceMon.ID)
	data.SystemServiceName = types.StringValue(systemServiceMon.SystemServiceName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
