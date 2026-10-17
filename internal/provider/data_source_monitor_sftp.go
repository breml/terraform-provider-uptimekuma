package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var _ datasource.DataSource = &MonitorSFTPDataSource{}

// NewMonitorSFTPDataSource returns a new instance of the SFTP monitor data source.
func NewMonitorSFTPDataSource() datasource.DataSource {
	return &MonitorSFTPDataSource{}
}

// MonitorSFTPDataSource manages SFTP monitor data source operations.
type MonitorSFTPDataSource struct {
	client *kuma.Client
}

// MonitorSFTPDataSourceModel describes the data model for SFTP monitor data source.
//
// The credentials are deliberately not exposed: a data source has no use for
// them that would justify copying secrets into the state of another module.
type MonitorSFTPDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Hostname    types.String `tfsdk:"hostname"`
	Port        types.Int64  `tfsdk:"port"`
	SSHUsername types.String `tfsdk:"ssh_username"`
}

// Metadata returns the metadata for the data source.
func (*MonitorSFTPDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_sftp"
}

// Schema returns the schema for the data source.
func (*MonitorSFTPDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get SFTP monitor information by ID or name. The SSH credentials are " +
			"not exposed. Requires Uptime Kuma 2.5.4 or later.",
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
				MarkdownDescription: "SFTP server IP address or hostname",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "TCP port of the SFTP server. Null while the check falls back to 22.",
				Computed:            true,
			},
			"ssh_username": schema.StringAttribute{
				MarkdownDescription: "User the check authenticates as",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorSFTPDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorSFTPDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorSFTPDataSourceModel

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

// readByID fetches the SFTP monitor data by its ID.
func (d *MonitorSFTPDataSource) readByID(
	ctx context.Context,
	data *MonitorSFTPDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var sftpMonitor monitor.SFTP
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &sftpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read SFTP monitor", err.Error())
		return
	}

	if !monitorTypeMatches(&resp.Diagnostics, data.ID.ValueInt64(), sftpMonitor.Base.Type(), sftpMonitor.Type()) {
		return
	}

	if !dataSourceNameMatches(&resp.Diagnostics, "Monitor", data.ID.ValueInt64(), data.Name, sftpMonitor.Name) {
		return
	}

	data.Name = types.StringValue(sftpMonitor.Name)
	populateSFTPDataSourceDetails(&sftpMonitor, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the SFTP monitor data by its name.
func (d *MonitorSFTPDataSource) readByName(
	ctx context.Context,
	data *MonitorSFTPDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "sftp", &resp.Diagnostics)
	if found == nil {
		return
	}

	var sftpMon monitor.SFTP
	err := found.As(&sftpMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(sftpMon.ID)
	populateSFTPDataSourceDetails(&sftpMon, data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// populateSFTPDataSourceDetails copies the SFTP specific fields into the data source model.
func populateSFTPDataSourceDetails(sftpMonitor *monitor.SFTP, data *MonitorSFTPDataSourceModel) {
	data.Hostname = types.StringValue(sftpMonitor.Hostname)
	data.Port = int64PtrToTypes(sftpMonitor.Port)
	data.SSHUsername = types.StringValue(sftpMonitor.SSHUsername)
}
