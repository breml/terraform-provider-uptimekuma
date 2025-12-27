// Package provider implements the Uptime Kuma Terraform provider.
// This file provides Real Browser monitor data source functionality.
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

var _ datasource.DataSource = &MonitorRealBrowserDataSource{}

// NewMonitorRealBrowserDataSource returns a new instance of the Real Browser monitor data source.
func NewMonitorRealBrowserDataSource() datasource.DataSource {
	return &MonitorRealBrowserDataSource{}
}

// MonitorRealBrowserDataSource manages Real Browser monitor data source operations.
type MonitorRealBrowserDataSource struct {
	client *kuma.Client
}

// MonitorRealBrowserDataSourceModel describes the data model for Real Browser monitor data source.
type MonitorRealBrowserDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorRealBrowserDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_real_browser"
}

// Schema returns the schema for the data source.
func (*MonitorRealBrowserDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Real Browser monitor information by ID or name",
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
func (d *MonitorRealBrowserDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	d.client = client
}

// Read reads the current state of the data source.
func (d *MonitorRealBrowserDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data MonitorRealBrowserDataSourceModel

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

func (d *MonitorRealBrowserDataSource) readByID(
	ctx context.Context,
	data *MonitorRealBrowserDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var realBrowserMonitor monitor.RealBrowser
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &realBrowserMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Real Browser monitor", err.Error())
		return
	}

	data.Name = types.StringValue(realBrowserMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *MonitorRealBrowserDataSource) readByName(
	ctx context.Context,
	data *MonitorRealBrowserDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "real-browser", &resp.Diagnostics)
	if found == nil {
		return
	}

	var realBrowserMon monitor.RealBrowser
	err := found.As(&realBrowserMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(realBrowserMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
