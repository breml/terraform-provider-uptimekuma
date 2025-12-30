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

var _ datasource.DataSource = &MonitorDockerDataSource{}

// NewMonitorDockerDataSource returns a new instance of the Docker monitor data source.
func NewMonitorDockerDataSource() datasource.DataSource {
	return &MonitorDockerDataSource{}
}

// MonitorDockerDataSource manages Docker monitor data source operations.
type MonitorDockerDataSource struct {
	client *kuma.Client
}

// MonitorDockerDataSourceModel describes the data model for Docker monitor data source.
type MonitorDockerDataSourceModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	DockerHostID    types.Int64  `tfsdk:"docker_host_id"`
	DockerContainer types.String `tfsdk:"docker_container"`
}

// Metadata returns the metadata for the data source.
func (*MonitorDockerDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_docker"
}

// Schema returns the schema for the data source.
func (*MonitorDockerDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Docker monitor information by ID or name",
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
			"docker_host_id": schema.Int64Attribute{
				MarkdownDescription: "Docker host ID",
				Computed:            true,
			},
			"docker_container": schema.StringAttribute{
				MarkdownDescription: "Docker container name or ID",
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *MonitorDockerDataSource) Configure(
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
func (d *MonitorDockerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MonitorDockerDataSourceModel

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

// readByID fetches the Docker monitor data by its ID.
func (d *MonitorDockerDataSource) readByID(
	ctx context.Context,
	data *MonitorDockerDataSourceModel,
	resp *datasource.ReadResponse,
) {
	var dockerMonitor monitor.Docker
	err := d.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &dockerMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Docker monitor", err.Error())
		return
	}

	data.Name = types.StringValue(dockerMonitor.Name)
	data.DockerHostID = types.Int64Value(dockerMonitor.DockerHost)
	data.DockerContainer = types.StringValue(dockerMonitor.DockerContainer)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readByName fetches the Docker monitor data by its name.
func (d *MonitorDockerDataSource) readByName(
	ctx context.Context,
	data *MonitorDockerDataSourceModel,
	resp *datasource.ReadResponse,
) {
	found := findMonitorByName(ctx, d.client, data.Name.ValueString(), "docker", &resp.Diagnostics)
	if found == nil {
		return
	}

	var dockerMon monitor.Docker
	err := found.As(&dockerMon)
	if err != nil {
		resp.Diagnostics.AddError("failed to convert monitor type", err.Error())
		return
	}

	data.ID = types.Int64Value(dockerMon.ID)
	data.DockerHostID = types.Int64Value(dockerMon.DockerHost)
	data.DockerContainer = types.StringValue(dockerMon.DockerContainer)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
