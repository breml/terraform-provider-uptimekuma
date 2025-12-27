package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorRedisResource{}
	_ resource.ResourceWithImportState = &MonitorRedisResource{}
)

// NewMonitorRedisResource returns a new instance of the Redis monitor resource.
func NewMonitorRedisResource() resource.Resource {
	return &MonitorRedisResource{}
}

// MonitorRedisResource defines the resource implementation.
type MonitorRedisResource struct {
	client *kuma.Client
}

// MonitorRedisResourceModel describes the resource data model.
type MonitorRedisResourceModel struct {
	MonitorBaseModel

	DatabaseConnectionString types.String `tfsdk:"database_connection_string"`
	IgnoreTLS                types.Bool   `tfsdk:"ignore_tls"`
}

func (r *MonitorRedisResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_redis"
}

func (r *MonitorRedisResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Redis monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"database_connection_string": schema.StringAttribute{
				MarkdownDescription: "Redis connection string (e.g., redis://user:password@host:port)",
				Required:            true,
				Sensitive:           true,
			},
			"ignore_tls": schema.BoolAttribute{
				MarkdownDescription: "Ignore TLS/SSL errors for Redis connections",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		}),
	}
}

// Configure configures the Redis monitor resource with the API client.
func (r *MonitorRedisResource) Configure(
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

// Create creates a new Redis monitor resource.
func (r *MonitorRedisResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorRedisResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	redisMonitor := monitor.Redis{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		RedisDetails: monitor.RedisDetails{
			ConnectionString: data.DatabaseConnectionString.ValueString(),
			IgnoreTLS:        data.IgnoreTLS.ValueBool(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		redisMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		redisMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		redisMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, redisMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Redis monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Redis monitor resource.
func (r *MonitorRedisResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorRedisResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var redisMonitor monitor.Redis
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &redisMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to read Redis monitor", err.Error())
		return
	}

	data.Name = types.StringValue(redisMonitor.Name)
	if redisMonitor.Description != nil {
		data.Description = types.StringValue(*redisMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(redisMonitor.Interval)
	data.RetryInterval = types.Int64Value(redisMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(redisMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(redisMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(redisMonitor.UpsideDown)
	data.Active = types.BoolValue(redisMonitor.IsActive)
	data.DatabaseConnectionString = types.StringValue(redisMonitor.ConnectionString)
	data.IgnoreTLS = types.BoolValue(redisMonitor.IgnoreTLS)

	if redisMonitor.Parent != nil {
		data.Parent = types.Int64Value(*redisMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(redisMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, redisMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, redisMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Redis monitor resource.
func (r *MonitorRedisResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorRedisResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorRedisResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	redisMonitor := monitor.Redis{
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
		RedisDetails: monitor.RedisDetails{
			ConnectionString: data.DatabaseConnectionString.ValueString(),
			IgnoreTLS:        data.IgnoreTLS.ValueBool(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		redisMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		redisMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		redisMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, redisMonitor)
	if err != nil {
		resp.Diagnostics.AddError("failed to update Redis monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Redis monitor resource.
func (r *MonitorRedisResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorRedisResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Redis monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (r *MonitorRedisResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
