// Package provider implements the Uptime Kuma Terraform provider.
// This file provides Push monitor data source functionality.
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

var _ datasource.DataSource = &MonitorPushDataSource{}

// NewMonitorPushDataSource returns a new instance of the Push monitor data source.
func NewMonitorPushDataSource() datasource.DataSource {
	return &MonitorPushDataSource{}
}

// MonitorPushDataSource manages Push monitor data source operations.
type MonitorPushDataSource struct {
	client *kuma.Client
}

// MonitorPushDataSourceModel describes the data model for Push monitor data source.
type MonitorPushDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*MonitorPushDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_push"
}

// Schema returns the schema for the data source.
func (*MonitorPushDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Push monitor information by ID or name",
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
func (d *MonitorPushDataSource) Configure(
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
func (d *MonitorPushDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorPushDataSourceModel

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

func (d *MonitorPushDataSource) readByID(
	ctx context.Context,
	data *MonitorPushDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var pushMonitor monitor.Push
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &pushMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Push monitor", err.Error())
		return
	}

	data.Name = types.StringValue(pushMonitor.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *MonitorPushDataSource) readByName(
	ctx context.Context,
	data *MonitorPushDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "push", &resp.Diagnostics)
	if found == nil {
		return
	}

	var pushMon monitor.Push
	err := found.As(&pushMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(pushMon.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
