package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	// Ensure MonitorDockerResource satisfies various resource interfaces.
	_ resource.Resource                = &MonitorDockerResource{}
	_ resource.ResourceWithImportState = &MonitorDockerResource{}
)

// NewMonitorDockerResource returns a new instance of the Docker monitor resource.
func NewMonitorDockerResource() resource.Resource {
	return &MonitorDockerResource{}
}

// MonitorDockerResource defines the resource implementation.
type MonitorDockerResource struct {
	client *kuma.Client
}

// MonitorDockerResourceModel describes the resource data model for Docker monitors.
type MonitorDockerResourceModel struct {
	MonitorBaseModel

	DockerHostID    types.Int64  `tfsdk:"docker_host_id"`   // Docker host ID.
	DockerContainer types.String `tfsdk:"docker_container"` // Docker container name or ID.
}

// Metadata returns the metadata for the resource.
func (*MonitorDockerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_docker"
}

// Schema returns the schema for the resource.
func (*MonitorDockerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Docker monitor resource",
		Attributes:          withMonitorBaseAttributes(withDockerMonitorAttributes(map[string]schema.Attribute{})),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorDockerResource) Configure(
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
func (r *MonitorDockerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorDockerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dockerMonitor := buildDockerMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateMonitor(ctx, dockerMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Docker monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildDockerMonitor(ctx context.Context, data *MonitorDockerResourceModel, diags *diag.Diagnostics) monitor.Docker {
	dockerMonitor := monitor.Docker{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		DockerDetails: monitor.DockerDetails{
			DockerHost:      data.DockerHostID.ValueInt64(),
			DockerContainer: data.DockerContainer.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		dockerMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		dockerMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		diags.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if !diags.HasError() {
			dockerMonitor.NotificationIDs = notificationIDs
		}
	}

	return dockerMonitor
}

// populateDockerMonitorBaseFields populates the base Docker monitor fields from the API response.
// Extracts Docker-specific fields from the API response into the model.
func populateDockerMonitorBaseFields(dockerMonitor *monitor.Docker, m *MonitorDockerResourceModel) {
	m.Name = types.StringValue(dockerMonitor.Name)
	if dockerMonitor.Description != nil {
		m.Description = types.StringValue(*dockerMonitor.Description)
	} else {
		m.Description = types.StringNull()
	}

	m.Interval = types.Int64Value(dockerMonitor.Interval)
	m.RetryInterval = types.Int64Value(dockerMonitor.RetryInterval)
	m.ResendInterval = types.Int64Value(dockerMonitor.ResendInterval)
	m.MaxRetries = types.Int64Value(dockerMonitor.MaxRetries)
	m.UpsideDown = types.BoolValue(dockerMonitor.UpsideDown)
	m.Active = types.BoolValue(dockerMonitor.IsActive)
	m.DockerHostID = types.Int64Value(dockerMonitor.DockerHost)
	m.DockerContainer = types.StringValue(dockerMonitor.DockerContainer)
}

// populateOptionalFieldsForDocker populates optional fields for Docker monitor.
// Handles parent group and notification IDs.
// Converts null API values to Terraform null types appropriately.
func populateOptionalFieldsForDocker(
	ctx context.Context,
	dockerMonitor *monitor.Docker,
	m *MonitorDockerResourceModel,
	diags *diag.Diagnostics,
) {
	// Set parent monitor group if present.
	if dockerMonitor.Parent != nil {
		m.Parent = types.Int64Value(*dockerMonitor.Parent)
	} else {
		m.Parent = types.Int64Null()
	}

	// Convert notification IDs list if present.
	if len(dockerMonitor.NotificationIDs) > 0 {
		notificationIDs, d := types.ListValueFrom(ctx, types.Int64Type, dockerMonitor.NotificationIDs)
		diags.Append(d...)
		m.NotificationIDs = notificationIDs
	} else {
		m.NotificationIDs = types.ListNull(types.Int64Type)
	}
}

// Read reads the current state of the resource.
func (r *MonitorDockerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorDockerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var dockerMonitor monitor.Docker
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &dockerMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Docker monitor", err.Error())
		return
	}

	populateDockerMonitorBaseFields(&dockerMonitor, &data)
	populateOptionalFieldsForDocker(ctx, &dockerMonitor, &data, &resp.Diagnostics)

	data.Tags = handleMonitorTagsRead(ctx, dockerMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MonitorDockerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorDockerResourceModel
	var state MonitorDockerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dockerMonitor := buildDockerMonitor(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	dockerMonitor.ID = data.ID.ValueInt64()

	err := r.client.UpdateMonitor(ctx, dockerMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update Docker monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *MonitorDockerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorDockerResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API.
	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Docker monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorDockerResource) ImportState(
	// Import monitor by ID.
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// withDockerMonitorAttributes adds Docker-specific schema attributes to the provided attribute map.
func withDockerMonitorAttributes(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	attrs["docker_host_id"] = schema.Int64Attribute{
		MarkdownDescription: "Docker host ID",
		Required:            true,
	}
	attrs["docker_container"] = schema.StringAttribute{
		MarkdownDescription: "Docker container name or ID to monitor",
		Required:            true,
	}
	return attrs
}
