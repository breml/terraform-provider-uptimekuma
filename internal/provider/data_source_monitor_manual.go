package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorManualDataSource{}

// NewMonitorManualDataSource returns a new instance of the manual monitor data source.
func NewMonitorManualDataSource() datasource.DataSource {
	return &MonitorManualDataSource{}
}

// MonitorManualDataSource manages manual monitor data source operations.
type MonitorManualDataSource struct {
	client *kuma.Client
}

// MonitorManualDataSourceModel describes the data model for manual monitor data source.
type MonitorManualDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorManualDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_manual"
}

// Schema returns the schema for the data source.
func (*MonitorManualDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get manual monitor information by ID or name. The manual status is " +
			"not exposed, as Uptime Kuma never returns it.",
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
func (d *MonitorManualDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorManualDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorManualDataSourceModel

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

// readByID fetches the manual monitor data by its ID.
func (d *MonitorManualDataSource) readByID(
	ctx context.Context,
	data *MonitorManualDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var manualMonitor monitor.Manual

	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &manualMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read manual monitor", err.Error())
		return
	}

	// The data source exposes only id and name, and GetMonitorAs decodes a
	// monitor of any type without complaint, so the type check is what rejects
	// a monitor of another type.
	if !monitorTypeMatches(
		&resp.Diagnostics, data.ID.ValueInt64(), manualMonitor.Base.Type(), manualMonitor.Type(),
	) {
		return
	}

	if !dataSourceNameMatches(&resp.Diagnostics, "Monitor", data.ID.ValueInt64(), data.Name, manualMonitor.Name) {
		return
	}

	data.Name = types.StringValue(manualMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the manual monitor data by its name.
func (d *MonitorManualDataSource) readByName(
	ctx context.Context,
	data *MonitorManualDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "manual", &resp.Diagnostics)
	if found == nil {
		return
	}

	var manualMonitor monitor.Manual

	err := found.As(&manualMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(manualMonitor.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
