// Package provider implements the Uptime Kuma Terraform provider.
// This file provides HTTP monitor data source functionality.
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

var _ datasource.DataSource = &MonitorHTTPDataSource{}

// NewMonitorHTTPDataSource returns a new instance of the HTTP monitor data source.
func NewMonitorHTTPDataSource() datasource.DataSource {
	return &MonitorHTTPDataSource{}
}

// MonitorHTTPDataSource manages HTTP monitor data source operations.
type MonitorHTTPDataSource struct {
	client *kuma.Client
}

// MonitorHTTPDataSourceModel describes the data model for HTTP monitor data source.
type MonitorHTTPDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	URL  types.String `tfsdk:"url"`
}

// Metadata returns the metadata for the data source.
func (*MonitorHTTPDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_http"
}

// Schema returns the schema for the data source.
func (*MonitorHTTPDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get HTTP monitor information by ID or name",
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
			"url": schema.StringAttribute{
				MarkdownDescription: "URL to monitor",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorHTTPDataSource) Configure(
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
func (d *MonitorHTTPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorHTTPDataSourceModel

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

func (d *MonitorHTTPDataSource) readByID(
	ctx context.Context,
	data *MonitorHTTPDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var httpMonitor monitor.HTTP
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &httpMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read HTTP monitor", err.Error())
		return
	}

	data.Name = types.StringValue(httpMonitor.Name)
	data.URL = types.StringValue(httpMonitor.URL)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *MonitorHTTPDataSource) readByName(
	ctx context.Context,
	data *MonitorHTTPDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "http", &resp.Diagnostics)
	if found == nil {
		return
	}

	var httpMon monitor.HTTP
	err := found.As(&httpMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(httpMon.ID)
	data.URL = types.StringValue(httpMon.URL)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
