package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorSIPOptionsDataSource{}

// NewMonitorSIPOptionsDataSource returns a new instance of the SIP Options monitor data source.
func NewMonitorSIPOptionsDataSource() datasource.DataSource {
	return &MonitorSIPOptionsDataSource{}
}

// MonitorSIPOptionsDataSource manages SIP Options monitor data source operations.
type MonitorSIPOptionsDataSource struct {
	client *kuma.Client
}

// MonitorSIPOptionsDataSourceModel describes the data model for SIP Options monitor data source.
type MonitorSIPOptionsDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Hostname types.String `tfsdk:"hostname"`
	Port     types.Int64  `tfsdk:"port"`
}

// Metadata returns the metadata for the data source.
func (*MonitorSIPOptionsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_sip_options"
}

// Schema returns the schema for the data source.
func (*MonitorSIPOptionsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get SIP Options monitor information by ID or name",
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
				MarkdownDescription: "Hostname or IP address being monitored",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "SIP port number being monitored",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorSIPOptionsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorSIPOptionsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorSIPOptionsDataSourceModel

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

// readByID fetches the SIP Options monitor data by its ID.
func (d *MonitorSIPOptionsDataSource) readByID(
	ctx context.Context,
	data *MonitorSIPOptionsDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var sipMon monitor.SIPOptions
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &sipMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to read SIP Options monitor", err.Error())
		return
	}

	data.Name = types.StringValue(sipMon.Name)
	data.Hostname = types.StringValue(sipMon.Hostname)
	data.Port = types.Int64Value(int64(sipMon.Port))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the SIP Options monitor data by its name.
func (d *MonitorSIPOptionsDataSource) readByName(
	ctx context.Context,
	data *MonitorSIPOptionsDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "sip-options", &resp.Diagnostics)
	if found == nil {
		return
	}

	var sipMon monitor.SIPOptions
	err := found.As(&sipMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(sipMon.ID)
	data.Hostname = types.StringValue(sipMon.Hostname)
	data.Port = types.Int64Value(int64(sipMon.Port))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
