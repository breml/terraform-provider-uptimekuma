// Package provider implements the Uptime Kuma Terraform provider.
// This file provides DNS monitor data source functionality.
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

var _ datasource.DataSource = &MonitorDNSDataSource{}

// NewMonitorDNSDataSource returns a new instance of the DNS monitor data source.
func NewMonitorDNSDataSource() datasource.DataSource {
	return &MonitorDNSDataSource{}
}

// MonitorDNSDataSource manages DNS monitor data source operations.
type MonitorDNSDataSource struct {
	client *kuma.Client
}

// MonitorDNSDataSourceModel describes the data model for DNS monitor data source.
type MonitorDNSDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Hostname types.String `tfsdk:"hostname"`
}

// Metadata returns the metadata for the data source.
func (*MonitorDNSDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_dns"
}

// Schema returns the schema for the data source.
func (*MonitorDNSDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get DNS monitor information by ID or name",
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
				MarkdownDescription: "Hostname to resolve",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorDNSDataSource) Configure(
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
func (d *MonitorDNSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorDNSDataSourceModel

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

func (d *MonitorDNSDataSource) readByID(
	ctx context.Context,
	data *MonitorDNSDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var dnsMonitor monitor.DNS
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &dnsMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read DNS monitor", err.Error())
		return
	}

	data.Name = types.StringValue(dnsMonitor.Name)
	data.Hostname = types.StringValue(dnsMonitor.Hostname)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *MonitorDNSDataSource) readByName(
	ctx context.Context,
	data *MonitorDNSDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "dns", &resp.Diagnostics)
	if found == nil {
		return
	}

	var dnsMon monitor.DNS
	err := found.As(&dnsMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(dnsMon.ID)
	data.Hostname = types.StringValue(dnsMon.Hostname)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
