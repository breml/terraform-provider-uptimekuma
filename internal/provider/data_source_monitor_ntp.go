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

var _ datasource.DataSource = &MonitorNTPDataSource{}

// NewMonitorNTPDataSource returns a new instance of the NTP monitor data source.
func NewMonitorNTPDataSource() datasource.DataSource {
	return &MonitorNTPDataSource{}
}

// MonitorNTPDataSource manages NTP monitor data source operations.
type MonitorNTPDataSource struct {
	client *kuma.Client
}

// MonitorNTPDataSourceModel describes the data model for NTP monitor data source.
type MonitorNTPDataSourceModel struct {
	ID                         types.Int64   `tfsdk:"id"`
	Name                       types.String  `tfsdk:"name"`
	Hostname                   types.String  `tfsdk:"hostname"`
	Port                       types.Int64   `tfsdk:"port"`
	Timeout                    types.Float64 `tfsdk:"timeout"`
	NTPStratumThreshold        types.Int64   `tfsdk:"ntp_stratum_threshold"`
	NTPTimeOffsetThreshold     types.Int64   `tfsdk:"ntp_time_offset_threshold"`
	NTPRootDispersionThreshold types.Int64   `tfsdk:"ntp_root_dispersion_threshold"`
}

// Metadata returns the metadata for the data source.
func (*MonitorNTPDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_ntp"
}

// Schema returns the schema for the data source.
func (*MonitorNTPDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get NTP (Network Time Protocol) monitor information by ID or name",
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
				MarkdownDescription: "NTP server IP address or hostname",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "UDP port of the NTP server. Null while the check falls back to 123.",
				Computed:            true,
			},
			"timeout": schema.Float64Attribute{
				MarkdownDescription: "Query timeout in seconds. The column is NOT NULL server-side, " +
					"so this is normally the stored value; null would mean the server reported none, " +
					"in which case the check falls back to 10.",
				Computed: true,
			},
			"ntp_stratum_threshold": schema.Int64Attribute{
				MarkdownDescription: "Stratum at which the monitor is considered down. Null while the " +
					"check falls back to 5.",
				Computed: true,
			},
			"ntp_time_offset_threshold": schema.Int64Attribute{
				MarkdownDescription: "Absolute time offset in milliseconds at which the monitor is " +
					"considered down. Null while the check falls back to 1000.",
				Computed: true,
			},
			"ntp_root_dispersion_threshold": schema.Int64Attribute{
				MarkdownDescription: "Root dispersion in milliseconds at which the monitor is considered " +
					"down. Null while the check falls back to 500.",
				Computed: true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorNTPDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorNTPDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorNTPDataSourceModel

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

// readByID fetches the NTP monitor data by its ID.
func (d *MonitorNTPDataSource) readByID(
	ctx context.Context,
	data *MonitorNTPDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var ntpMonitor monitor.NTP
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &ntpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read NTP monitor", err.Error())
		return
	}

	// GetMonitorAs unmarshals into monitor.NTP without checking the type, so a
	// monitor of another type would decode into plausible-looking empty values.
	if actual := ntpMonitor.Base.Type(); actual != "" && actual != ntpMonitor.Type() {
		resp.Diagnostics.AddError(
			"Monitor type mismatch",
			fmt.Sprintf(
				"Monitor ID %d has type %q, expected %q.",
				data.ID.ValueInt64(), actual, ntpMonitor.Type(),
			),
		)

		return
	}

	data.Name = types.StringValue(ntpMonitor.Name)
	populateNTPDataSourceDetails(&ntpMonitor, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the NTP monitor data by its name.
func (d *MonitorNTPDataSource) readByName(
	ctx context.Context,
	data *MonitorNTPDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "ntp", &resp.Diagnostics)
	if found == nil {
		return
	}

	var ntpMon monitor.NTP
	err := found.As(&ntpMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(ntpMon.ID)
	populateNTPDataSourceDetails(&ntpMon, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// populateNTPDataSourceDetails copies the NTP specific fields into the data source model.
func populateNTPDataSourceDetails(ntpMonitor *monitor.NTP, data *MonitorNTPDataSourceModel) {
	data.Hostname = types.StringValue(ntpMonitor.Hostname)
	data.Port = int64PtrToTypes(ntpMonitor.Port)
	data.NTPStratumThreshold = int64PtrToTypes(ntpMonitor.NTPStratumThreshold)
	data.NTPTimeOffsetThreshold = int64PtrToTypes(ntpMonitor.NTPTimeOffsetThreshold)
	data.NTPRootDispersionThreshold = int64PtrToTypes(ntpMonitor.NTPRootDispersionThreshold)

	// Unlike the resource, nothing here forces a non-null value, so report the
	// column as it is stored and keep null meaning "the check applies its fallback".
	if ntpMonitor.Timeout != nil {
		data.Timeout = types.Float64Value(*ntpMonitor.Timeout)
	} else {
		data.Timeout = types.Float64Null()
	}
}
