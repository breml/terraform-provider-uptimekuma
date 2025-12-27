package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/dockerhost"
)

var _ resource.Resource = &DockerHostResource{}

// NewDockerHostResource returns a new instance of the docker host resource.
func NewDockerHostResource() resource.Resource {
	return &DockerHostResource{}
}

// DockerHostResource defines the resource implementation.
type DockerHostResource struct {
	client *kuma.Client
}

// DockerHostResourceModel describes the resource data model.
type DockerHostResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	DockerDaemon types.String `tfsdk:"docker_daemon"`
	DockerType   types.String `tfsdk:"docker_type"`
}

// Metadata returns the metadata for the resource.
func (*DockerHostResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_docker_host"
}

// Schema returns the schema for the resource.
func (*DockerHostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Docker host resource for managing Docker daemon connections in Uptime Kuma",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Docker host identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable name for the Docker host",
			},
			"docker_daemon": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Connection string for the Docker daemon (e.g., unix:///var/run/docker.sock, tcp://host:2375)",
			},
			"docker_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Docker connection type: socket or tcp",
				Validators: []validator.String{
					stringvalidator.OneOf("socket", "tcp"),
				},
			},
		},
	}
}

// Configure configures the resource with the API client.
func (r *DockerHostResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *kuma.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = client
}

// Create creates a new resource.
func (r *DockerHostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DockerHostResourceModel

	// Extract planned configuration.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build Docker host configuration from plan.
	config := dockerhost.Config{
		Name:         data.Name.ValueString(),
		DockerDaemon: data.DockerDaemon.ValueString(),
		DockerType:   data.DockerType.ValueString(),
	}

	// Call API to create Docker host.
	id, err := r.client.CreateDockerHost(ctx, config)
	if err != nil {
		resp.Diagnostics.AddError("failed to create docker host", err.Error())
		return
	}

	// Set computed ID and save state.
	data.ID = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the resource.
func (r *DockerHostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DockerHostResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch current Docker host configuration from API.
	dh, err := r.client.GetDockerHost(ctx, data.ID.ValueInt64())
	if err != nil {
		// Handle resource not found error.
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read docker host", err.Error())
		return
	}

	// Update resource attributes from API response.
	data.Name = types.StringValue(dh.Name)
	data.DockerDaemon = types.StringValue(dh.DockerDaemon)
	data.DockerType = types.StringValue(dh.DockerType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *DockerHostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DockerHostResourceModel

	// Extract planned configuration.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build Docker host configuration with updated values.
	config := dockerhost.Config{
		ID:           data.ID.ValueInt64(),
		Name:         data.Name.ValueString(),
		DockerDaemon: data.DockerDaemon.ValueString(),
		DockerType:   data.DockerType.ValueString(),
	}

	// Call API to update Docker host.
	err := r.client.UpdateDockerHost(ctx, config)
	if err != nil {
		resp.Diagnostics.AddError("failed to update docker host", err.Error())
		return
	}

	// Save updated state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *DockerHostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DockerHostResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to delete Docker host.
	err := r.client.DeleteDockerHost(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete docker host", err.Error())
		return
	}
}
