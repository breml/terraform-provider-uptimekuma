package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorRabbitMQResource{}
	_ resource.ResourceWithImportState = &MonitorRabbitMQResource{}
)

// NewMonitorRabbitMQResource returns a new instance of the RabbitMQ monitor resource.
func NewMonitorRabbitMQResource() resource.Resource {
	return &MonitorRabbitMQResource{}
}

// MonitorRabbitMQResource defines the resource implementation.
type MonitorRabbitMQResource struct {
	client *kuma.Client
}

// MonitorRabbitMQResourceModel describes the resource data model.
type MonitorRabbitMQResourceModel struct {
	MonitorBaseModel

	Nodes    types.String `tfsdk:"nodes"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Timeout  types.Int64  `tfsdk:"timeout"`
}

// Metadata returns the metadata for the resource.
func (*MonitorRabbitMQResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_rabbitmq"
}

// Schema returns the schema for the resource.
func (*MonitorRabbitMQResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "RabbitMQ monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"nodes": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded array of RabbitMQ management API node URLs " +
					"(e.g., `[\"http://rabbitmq.example.com:15672/\"]`)",
				Required: true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for HTTP Basic authentication against the RabbitMQ management API",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for HTTP Basic authentication against the RabbitMQ management API",
				Optional:            true,
				Sensitive:           true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Request timeout in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(48),
			},
		}),
	}
}

// Configure configures the RabbitMQ monitor resource with the API client.
func (r *MonitorRabbitMQResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new RabbitMQ monitor resource.
func (r *MonitorRabbitMQResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorRabbitMQResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rabbitMQMonitor := monitor.RabbitMQ{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		RabbitMQDetails: monitor.RabbitMQDetails{
			Nodes:    data.Nodes.ValueString(),
			Username: strToPtr(data.Username),
			Password: strToPtr(data.Password),
			Timeout:  int64ToPtr(data.Timeout),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		rabbitMQMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		rabbitMQMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		rabbitMQMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, &rabbitMQMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create RabbitMQ monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err = handleMonitorActiveStateCreate(ctx, r.client, id, data.Active)
	if err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError("failed to apply monitor active state", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the RabbitMQ monitor resource.
func (r *MonitorRabbitMQResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorRabbitMQResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var rabbitMQMonitor monitor.RabbitMQ
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &rabbitMQMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read RabbitMQ monitor", err.Error())
		return
	}

	data.Name = types.StringValue(rabbitMQMonitor.Name)
	if rabbitMQMonitor.Description != nil {
		data.Description = types.StringValue(*rabbitMQMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(rabbitMQMonitor.Interval)
	data.RetryInterval = types.Int64Value(rabbitMQMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(rabbitMQMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(rabbitMQMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(rabbitMQMonitor.UpsideDown)
	data.Active = types.BoolValue(rabbitMQMonitor.IsActive)
	data.Nodes = types.StringValue(rabbitMQMonitor.Nodes)
	data.Username = ptrToTypes(rabbitMQMonitor.Username)
	data.Password = ptrToTypes(rabbitMQMonitor.Password)

	if rabbitMQMonitor.Timeout != nil {
		data.Timeout = types.Int64Value(*rabbitMQMonitor.Timeout)
	} else {
		data.Timeout = types.Int64Null()
	}

	if rabbitMQMonitor.Parent != nil {
		data.Parent = types.Int64Value(*rabbitMQMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(rabbitMQMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, rabbitMQMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, rabbitMQMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the RabbitMQ monitor resource.
func (r *MonitorRabbitMQResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorRabbitMQResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorRabbitMQResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rabbitMQMonitor := monitor.RabbitMQ{
		Base: monitor.Base{
			ID:             data.ID.ValueInt64(),
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		RabbitMQDetails: monitor.RabbitMQDetails{
			Nodes:    data.Nodes.ValueString(),
			Username: strToPtr(data.Username),
			Password: strToPtr(data.Password),
			Timeout:  int64ToPtr(data.Timeout),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		rabbitMQMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		rabbitMQMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		rabbitMQMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, &rabbitMQMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update RabbitMQ monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	handleMonitorActiveStateUpdate(ctx, r.client, data.ID.ValueInt64(), state.Active, data.Active, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the RabbitMQ monitor resource.
func (r *MonitorRabbitMQResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorRabbitMQResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete RabbitMQ monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorRabbitMQResource) ImportState(
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
