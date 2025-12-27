package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorPostgresResource{}
	_ resource.ResourceWithImportState = &MonitorPostgresResource{}
)

// NewMonitorPostgresResource returns a new instance of the PostgreSQL monitor resource.
func NewMonitorPostgresResource() resource.Resource {
	return &MonitorPostgresResource{}
}

// MonitorPostgresResource defines the resource implementation.
type MonitorPostgresResource struct {
	client *kuma.Client
}

// MonitorPostgresResourceModel describes the resource data model.
type MonitorPostgresResourceModel struct {
	MonitorBaseModel

	DatabaseConnectionString types.String `tfsdk:"database_connection_string"`
	DatabaseQuery            types.String `tfsdk:"database_query"`
}

// Metadata returns the metadata for the resource.
func (*MonitorPostgresResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_postgres"
}

// Schema returns the schema for the resource.
func (*MonitorPostgresResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PostgreSQL monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"database_connection_string": schema.StringAttribute{
				MarkdownDescription: "PostgreSQL connection string (e.g., postgres://username:password@host:port/database)",
				Required:            true,
				Sensitive:           true,
			},
			"database_query": schema.StringAttribute{
				MarkdownDescription: "SQL query to execute for health check",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("SELECT 1"),
			},
		}),
	}
}

// Configure configures the PostgreSQL monitor resource with the API client.
func (r *MonitorPostgresResource) Configure(
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

// Create creates a new PostgreSQL monitor resource.
func (r *MonitorPostgresResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorPostgresResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	postgresMonitor := monitor.Postgres{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		PostgresDetails: monitor.PostgresDetails{
			DatabaseConnectionString: data.DatabaseConnectionString.ValueString(),
			DatabaseQuery:            data.DatabaseQuery.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		postgresMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		postgresMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		postgresMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, postgresMonitor)
 // Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create PostgreSQL monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

 // Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the PostgreSQL monitor resource.
func (r *MonitorPostgresResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorPostgresResourceModel

 // Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var postgresMonitor monitor.Postgres
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &postgresMonitor)
 // Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to read PostgreSQL monitor", err.Error())
		return
	}

	data.Name = types.StringValue(postgresMonitor.Name)
	if postgresMonitor.Description != nil {
		data.Description = types.StringValue(*postgresMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(postgresMonitor.Interval)
	data.RetryInterval = types.Int64Value(postgresMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(postgresMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(postgresMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(postgresMonitor.UpsideDown)
	data.Active = types.BoolValue(postgresMonitor.IsActive)
	data.DatabaseConnectionString = types.StringValue(postgresMonitor.DatabaseConnectionString)
	data.DatabaseQuery = types.StringValue(postgresMonitor.DatabaseQuery)

	if postgresMonitor.Parent != nil {
		data.Parent = types.Int64Value(*postgresMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(postgresMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, postgresMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, postgresMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

 // Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the PostgreSQL monitor resource.
func (r *MonitorPostgresResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorPostgresResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorPostgresResourceModel

 // Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postgresMonitor := monitor.Postgres{
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
		PostgresDetails: monitor.PostgresDetails{
			DatabaseConnectionString: data.DatabaseConnectionString.ValueString(),
			DatabaseQuery:            data.DatabaseQuery.ValueString(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		postgresMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		postgresMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		postgresMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, postgresMonitor)
 // Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update PostgreSQL monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

 // Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the PostgreSQL monitor resource.
func (r *MonitorPostgresResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorPostgresResourceModel

 // Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
 // Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete PostgreSQL monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorPostgresResource) ImportState(
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
