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
	_ resource.Resource                = &MonitorMySQLResource{}
	_ resource.ResourceWithImportState = &MonitorMySQLResource{}
)

// NewMonitorMySQLResource returns a new instance of the MySQL monitor resource.
func NewMonitorMySQLResource() resource.Resource {
	return &MonitorMySQLResource{}
}

// MonitorMySQLResource defines the resource implementation.
type MonitorMySQLResource struct {
	client *kuma.Client
}

// MonitorMySQLResourceModel describes the resource data model.
type MonitorMySQLResourceModel struct {
	MonitorBaseModel

	DatabaseConnectionString types.String `tfsdk:"database_connection_string"`
	DatabaseQuery            types.String `tfsdk:"database_query"`
}

// Metadata returns the metadata for the resource.
func (*MonitorMySQLResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_mysql"
}

// Schema returns the schema for the resource.
func (*MonitorMySQLResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MySQL monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"database_connection_string": schema.StringAttribute{
				MarkdownDescription: "MySQL connection string (e.g., user:password@tcp(host:port)/database)",
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

// Configure configures the MySQL monitor resource with the API client.
func (r *MonitorMySQLResource) Configure(
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

// Create creates a new MySQL monitor resource.
func (r *MonitorMySQLResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorMySQLResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	databaseQuery := data.DatabaseQuery.ValueString()
	mysqlMonitor := monitor.MySQL{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		MySQLDetails: monitor.MySQLDetails{
			DatabaseConnectionString: data.DatabaseConnectionString.ValueString(),
			DatabaseQuery:            &databaseQuery,
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		mysqlMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		mysqlMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		mysqlMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, &mysqlMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create MySQL monitor", err.Error())
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

// Read reads the current state of the MySQL monitor resource.
func (r *MonitorMySQLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorMySQLResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var mysqlMonitor monitor.MySQL
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &mysqlMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to read MySQL monitor", err.Error())
		return
	}

	data.Name = types.StringValue(mysqlMonitor.Name)
	if mysqlMonitor.Description != nil {
		data.Description = types.StringValue(*mysqlMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(mysqlMonitor.Interval)
	data.RetryInterval = types.Int64Value(mysqlMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(mysqlMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(mysqlMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(mysqlMonitor.UpsideDown)
	data.Active = types.BoolValue(mysqlMonitor.IsActive)
	data.DatabaseConnectionString = types.StringValue(mysqlMonitor.DatabaseConnectionString)
	if mysqlMonitor.DatabaseQuery != nil {
		data.DatabaseQuery = types.StringValue(*mysqlMonitor.DatabaseQuery)
	} else {
		data.DatabaseQuery = types.StringNull()
	}

	if mysqlMonitor.Parent != nil {
		data.Parent = types.Int64Value(*mysqlMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(mysqlMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, mysqlMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, mysqlMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the MySQL monitor resource.
func (r *MonitorMySQLResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorMySQLResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorMySQLResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseQuery := data.DatabaseQuery.ValueString()
	mysqlMonitor := monitor.MySQL{
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
		MySQLDetails: monitor.MySQLDetails{
			DatabaseConnectionString: data.DatabaseConnectionString.ValueString(),
			DatabaseQuery:            &databaseQuery,
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		mysqlMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		mysqlMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		mysqlMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, &mysqlMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update MySQL monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the MySQL monitor resource.
func (r *MonitorMySQLResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorMySQLResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete MySQL monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorMySQLResource) ImportState(
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
