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

var _ datasource.DataSource = &MonitorPM2DataSource{}

// NewMonitorPM2DataSource returns a new instance of the PM2 monitor data source.
func NewMonitorPM2DataSource() datasource.DataSource {
	return &MonitorPM2DataSource{}
}

// MonitorPM2DataSource manages PM2 monitor data source operations.
type MonitorPM2DataSource struct {
	client *kuma.Client
}

// MonitorPM2DataSourceModel describes the data model for PM2 monitor data source.
type MonitorPM2DataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProcessName types.String `tfsdk:"process_name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorPM2DataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_pm2"
}

// Schema returns the schema for the data source.
func (*MonitorPM2DataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get PM2 monitor information by ID or name",
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
			"process_name": schema.StringAttribute{
				MarkdownDescription: "PM2 process name or numeric PM2 id being checked, as stored by " +
					"Uptime Kuma and therefore without surrounding whitespace",
				Computed: true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorPM2DataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorPM2DataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorPM2DataSourceModel

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

// readByID fetches the PM2 monitor data by its ID.
func (d *MonitorPM2DataSource) readByID(
	ctx context.Context,
	data *MonitorPM2DataSourceModel,
	resp *datasource.ReadResponse,
) {
	var pm2Monitor monitor.PM2

	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &pm2Monitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read PM2 monitor", err.Error())
		return
	}

	// GetMonitorAs unmarshals into monitor.PM2 without checking the type, and
	// the wire field is shared with the system-service monitor, so a monitor of
	// another type would decode into plausible-looking values.
	if actual := pm2Monitor.Base.Type(); actual != "" && actual != pm2Monitor.Type() {
		resp.Diagnostics.AddError(
			"Monitor type mismatch",
			fmt.Sprintf(
				"Monitor ID %d has type %q, expected %q.",
				data.ID.ValueInt64(), actual, pm2Monitor.Type(),
			),
		)

		return
	}

	data.Name = types.StringValue(pm2Monitor.Name)
	data.ProcessName = types.StringValue(pm2Monitor.ProcessName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the PM2 monitor data by its name.
func (d *MonitorPM2DataSource) readByName(
	ctx context.Context,
	data *MonitorPM2DataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "pm2", &resp.Diagnostics)
	if found == nil {
		return
	}

	var pm2Monitor monitor.PM2

	err := found.As(&pm2Monitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(pm2Monitor.ID)
	data.ProcessName = types.StringValue(pm2Monitor.ProcessName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
