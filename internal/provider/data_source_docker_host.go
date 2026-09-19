package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
)

var _ datasource.DataSource = &DockerHostDataSource{}

// NewDockerHostDataSource returns a new instance of the Docker host data source.
func NewDockerHostDataSource() datasource.DataSource {
	return &DockerHostDataSource{}
}

// DockerHostDataSource manages Docker host data source operations.
type DockerHostDataSource struct {
	client *kuma.Client
}

// DockerHostDataSourceModel describes the data model for Docker host data source.
type DockerHostDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the metadata for the data source.
func (*DockerHostDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_docker_host"
}

// Schema returns the schema for the data source.
func (*DockerHostDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Docker host information by ID or name",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Docker host identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Docker host name",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// Configure configures the data source with the API client.
func (d *DockerHostDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the current state of the data source.
func (d *DockerHostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DockerHostDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Attempt to read by ID if provided.
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		d.readByID(ctx, &data, resp)
		return
	}

	// Attempt to read by name if ID not provided.
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		d.readByName(ctx, &data, resp)
		return
	}

	// Error if neither ID nor name provided.
	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}

func (d *DockerHostDataSource) readByID(
	ctx context.Context,
	data *DockerHostDataSourceModel,
	resp *datasource.ReadResponse,
) {
	// The Docker host getter serves from the state cache, so a resync is
	// forced once before a miss is believed, see readWithResync.
	dockerHost, found := readWithResync(
		ctx, d.client, data.ID.ValueInt64(), "failed to read Docker host", d.client.GetDockerHost,
		&resp.Diagnostics,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Docker host not found",
			fmt.Sprintf("No Docker host with ID %d found.", data.ID.ValueInt64()),
		)

		return
	}

	// Populate name and set response state.
	data.Name = types.StringValue(dockerHost.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *DockerHostDataSource) readByName(
	ctx context.Context,
	data *DockerHostDataSourceModel,
	resp *datasource.ReadResponse,
) {
	// The list serves from the state cache, so a resync is forced once before
	// a miss is believed, see findWithResync.
	matches, found := findWithResync(ctx, d.client, func(ctx context.Context) ([]int64, bool) {
		var ids []int64

		dockerHosts := d.client.GetDockerHostList(ctx)
		for i := range dockerHosts {
			if dockerHosts[i].Name == data.Name.ValueString() {
				ids = append(ids, dockerHosts[i].ID)
			}
		}

		return ids, len(ids) > 0
	}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Error if no host found with given name.
	if !found {
		resp.Diagnostics.AddError(
			"Docker host not found",
			fmt.Sprintf("No Docker host with name '%s' found.", data.Name.ValueString()),
		)

		return
	}

	// Error if multiple hosts match name.
	if len(matches) > 1 {
		resp.Diagnostics.AddError(
			"Multiple Docker hosts found",
			fmt.Sprintf(
				"Multiple Docker hosts with name '%s' found. Please use 'id' to specify the host uniquely.",
				data.Name.ValueString(),
			),
		)

		return
	}

	// Populate ID and set response state.
	data.ID = types.Int64Value(matches[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
