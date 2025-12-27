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

func NewDockerHostDataSource() datasource.DataSource {
	return &DockerHostDataSource{}
}

type DockerHostDataSource struct {
	client *kuma.Client
}

type DockerHostDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *DockerHostDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_docker_host"
}

func (d *DockerHostDataSource) Schema(
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

func (d *DockerHostDataSource) Configure(
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

func (d *DockerHostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DockerHostDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		dockerHost, err := d.client.GetDockerHost(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("failed to read Docker host", err.Error())
			return
		}

		data.Name = types.StringValue(dockerHost.Name)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		dockerHosts := d.client.GetDockerHostList(ctx)

		var found *struct {
			ID   int64
			Name string
		}

		for i := range dockerHosts {
			if dockerHosts[i].Name == data.Name.ValueString() {
				if found != nil {
					resp.Diagnostics.AddError(
						"Multiple Docker hosts found",
						fmt.Sprintf(
							"Multiple Docker hosts with name '%s' found. Please use 'id' to specify the host uniquely.",
							data.Name.ValueString(),
						),
					)
					return
				}

				found = &struct {
					ID   int64
					Name string
				}{
					ID:   dockerHosts[i].ID,
					Name: dockerHosts[i].Name,
				}
			}
		}

		if found == nil {
			resp.Diagnostics.AddError(
				"Docker host not found",
				fmt.Sprintf("No Docker host with name '%s' found.", data.Name.ValueString()),
			)
			return
		}

		data.ID = types.Int64Value(found.ID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError(
		"Missing query parameters",
		"Either 'id' or 'name' must be specified.",
	)
}
