package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

// _ ensures the interface is implemented.
var _ datasource.DataSource = &MonitorGroupDataSource{}

// NewMonitorGroupDataSource returns a new instance of the monitor group data source.
func NewMonitorGroupDataSource() datasource.DataSource {
	return &MonitorGroupDataSource{}
}

// MonitorGroupDataSource manages monitor group data source operations.
type MonitorGroupDataSource struct {
	client *kuma.Client
}

// MonitorGroupDataSourceModel describes the data model for monitor group data source.
type MonitorGroupDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorGroupDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_group"
}

// Schema returns the schema for the data source.
func (*MonitorGroupDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get monitor group information by ID or name",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Monitor group identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Monitor group name",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorGroupDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *MonitorGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorGroupDataSourceModel

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

func (d *MonitorGroupDataSource) readByID(
	ctx context.Context,
	data *MonitorGroupDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var groupMonitor monitor.Group
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &groupMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read monitor group", err.Error())
		return
	}

	data.Name = types.StringValue(groupMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *MonitorGroupDataSource) readByName(
	ctx context.Context,
	data *MonitorGroupDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "group", &resp.Diagnostics)
	if found == nil {
		return
	}

	var groupMon monitor.Group
	err := found.As(&groupMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(groupMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
