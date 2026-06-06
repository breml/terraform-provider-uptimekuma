package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorGlobalpingDataSource{}

// NewMonitorGlobalpingDataSource returns a new instance of the Globalping monitor data source.
func NewMonitorGlobalpingDataSource() datasource.DataSource {
	return &MonitorGlobalpingDataSource{}
}

// MonitorGlobalpingDataSource manages Globalping monitor data source operations.
type MonitorGlobalpingDataSource struct {
	client *kuma.Client
}

// MonitorGlobalpingDataSourceModel describes the data model for Globalping monitor data source.
type MonitorGlobalpingDataSourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Subtype          types.String `tfsdk:"subtype"`
	Location         types.String `tfsdk:"location"`
	IPFamily         types.String `tfsdk:"ip_family"`
	Protocol         types.String `tfsdk:"protocol"`
	PingCount        types.Int64  `tfsdk:"ping_count"`
	Hostname         types.String `tfsdk:"hostname"`
	Port             types.Int64  `tfsdk:"port"`
	DNSResolveType   types.String `tfsdk:"dns_resolve_type"`
	DNSResolveServer types.String `tfsdk:"dns_resolve_server"`
	Keyword          types.String `tfsdk:"keyword"`
	InvertKeyword    types.Bool   `tfsdk:"invert_keyword"`
	ExpectedValue    types.String `tfsdk:"expected_value"`
	JSONPath         types.String `tfsdk:"json_path"`
	JSONPathOperator types.String `tfsdk:"json_path_operator"`
}

// Metadata returns the metadata for the data source.
func (*MonitorGlobalpingDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_globalping"
}

// Schema returns the schema for the data source.
func (*MonitorGlobalpingDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Globalping monitor information by ID or name",
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
			"subtype": schema.StringAttribute{
				MarkdownDescription: "Check type performed by Globalping probes",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Probe location selector",
				Computed:            true,
			},
			"ip_family": schema.StringAttribute{
				MarkdownDescription: "IP protocol version",
				Computed:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol used for ping or traceroute checks",
				Computed:            true,
			},
			"ping_count": schema.Int64Attribute{
				MarkdownDescription: "Number of ping packets to send",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Target hostname for DNS or port checks",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Target port for port checks",
				Computed:            true,
			},
			"dns_resolve_type": schema.StringAttribute{
				MarkdownDescription: "DNS record type to resolve",
				Computed:            true,
			},
			"dns_resolve_server": schema.StringAttribute{
				MarkdownDescription: "DNS server to use for resolution",
				Computed:            true,
			},
			"keyword": schema.StringAttribute{
				MarkdownDescription: "Keyword to search for in the HTTP response body",
				Computed:            true,
			},
			"invert_keyword": schema.BoolAttribute{
				MarkdownDescription: "Whether the keyword match logic is inverted",
				Computed:            true,
			},
			"expected_value": schema.StringAttribute{
				MarkdownDescription: "Expected value for JSON path evaluation",
				Computed:            true,
			},
			"json_path": schema.StringAttribute{
				MarkdownDescription: "JSON path expression to evaluate in the response",
				Computed:            true,
			},
			"json_path_operator": schema.StringAttribute{
				MarkdownDescription: "Comparison operator for the JSON path check",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorGlobalpingDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorGlobalpingDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorGlobalpingDataSourceModel

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

// readByID fetches the Globalping monitor data by its ID.
func (d *MonitorGlobalpingDataSource) readByID(
	ctx context.Context,
	data *MonitorGlobalpingDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var globalpingMonitor monitor.Globalping
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &globalpingMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Globalping monitor", err.Error())
		return
	}

	populateGlobalpingDataSourceModel(&globalpingMonitor, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the Globalping monitor data by its name.
func (d *MonitorGlobalpingDataSource) readByName(
	ctx context.Context,
	data *MonitorGlobalpingDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "globalping", &resp.Diagnostics)
	if found == nil {
		return
	}

	var globalpingMonitor monitor.Globalping
	err := found.As(&globalpingMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(globalpingMonitor.ID)
	populateGlobalpingDataSourceModel(&globalpingMonitor, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// populateGlobalpingDataSourceModel populates the data source model from the Globalping monitor API response.
func populateGlobalpingDataSourceModel(globalpingMonitor *monitor.Globalping, data *MonitorGlobalpingDataSourceModel) {
	data.Name = types.StringValue(globalpingMonitor.Name)
	data.Subtype = types.StringValue(string(globalpingMonitor.Subtype))
	data.Location = types.StringValue(globalpingMonitor.Location)
	data.IPFamily = types.StringValue(string(globalpingMonitor.IPFamily))
	data.Protocol = types.StringValue(globalpingMonitor.Protocol)
	data.PingCount = types.Int64Value(int64(globalpingMonitor.PingCount))
	data.Hostname = types.StringValue(globalpingMonitor.Hostname)
	data.Port = types.Int64Value(int64(globalpingMonitor.Port))
	data.DNSResolveType = types.StringValue(string(globalpingMonitor.DNSResolveType))
	data.DNSResolveServer = types.StringValue(globalpingMonitor.DNSResolveServer)
	data.Keyword = types.StringValue(globalpingMonitor.Keyword)
	data.InvertKeyword = types.BoolValue(globalpingMonitor.InvertKeyword)
	data.ExpectedValue = types.StringValue(globalpingMonitor.ExpectedValue)
	data.JSONPath = types.StringValue(globalpingMonitor.JSONPath)
	data.JSONPathOperator = types.StringValue(globalpingMonitor.JSONPathOperator)
}
