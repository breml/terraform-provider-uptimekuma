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
	_ resource.Resource                = &MonitorSQLServerResource{}
	_ resource.ResourceWithImportState = &MonitorSQLServerResource{}
)

// NewMonitorSQLServerResource returns a new instance of the SQL Server monitor resource.
func NewMonitorSQLServerResource() resource.Resource {
	return &MonitorSQLServerResource{}
}

// MonitorSQLServerResource defines the resource implementation.
type MonitorSQLServerResource struct {
	client *kuma.Client
}

// MonitorSQLServerResourceModel describes the resource data model.
type MonitorSQLServerResourceModel struct {
	MonitorBaseModel

	DatabaseConnectionString types.String `tfsdk:"database_connection_string"`
	DatabaseQuery            types.String `tfsdk:"database_query"`
}

// Metadata returns the metadata for the resource.
func (*MonitorSQLServerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_sqlserver"
}

// Schema returns the schema for the resource.
func (*MonitorSQLServerResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SQL Server monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"database_connection_string": schema.StringAttribute{
				MarkdownDescription: "SQL Server connection string",
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

// Configure configures the SQL Server monitor resource with the API client.
func (r *MonitorSQLServerResource) Configure(
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

// Create creates a new SQL Server monitor resource.
func (r *MonitorSQLServerResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorSQLServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sqlserverMonitor := monitor.SQLServer{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		SQLServerDetails: monitor.SQLServerDetails{
			DatabaseConnectionString: data.DatabaseConnectionString.ValueString(),
			DatabaseQuery:            ptrString(data.DatabaseQuery.ValueString()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		sqlserverMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		sqlserverMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		sqlserverMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, &sqlserverMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create SQL Server monitor", err.Error())
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

// Read reads the current state of the SQL Server monitor resource.
func (r *MonitorSQLServerResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorSQLServerResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var sqlserverMonitor monitor.SQLServer
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &sqlserverMonitor)
	// Handle error.
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read SQL Server monitor", err.Error())
		return
	}

	data.Name = types.StringValue(sqlserverMonitor.Name)
	if sqlserverMonitor.Description != nil {
		data.Description = types.StringValue(*sqlserverMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(sqlserverMonitor.Interval)
	data.RetryInterval = types.Int64Value(sqlserverMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(sqlserverMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(sqlserverMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(sqlserverMonitor.UpsideDown)
	data.Active = types.BoolValue(sqlserverMonitor.IsActive)
	data.DatabaseConnectionString = types.StringValue(sqlserverMonitor.DatabaseConnectionString)
	if sqlserverMonitor.DatabaseQuery != nil {
		data.DatabaseQuery = types.StringValue(*sqlserverMonitor.DatabaseQuery)
	} else {
		data.DatabaseQuery = types.StringValue("SELECT 1")
	}

	if sqlserverMonitor.Parent != nil {
		data.Parent = types.Int64Value(*sqlserverMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(sqlserverMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, sqlserverMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, sqlserverMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the SQL Server monitor resource.
func (r *MonitorSQLServerResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorSQLServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorSQLServerResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sqlserverMonitor := monitor.SQLServer{
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
		SQLServerDetails: monitor.SQLServerDetails{
			DatabaseConnectionString: data.DatabaseConnectionString.ValueString(),
			DatabaseQuery:            ptrString(data.DatabaseQuery.ValueString()),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		sqlserverMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		sqlserverMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		sqlserverMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, &sqlserverMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update SQL Server monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the SQL Server monitor resource.
func (r *MonitorSQLServerResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorSQLServerResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete SQL Server monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorSQLServerResource) ImportState(
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

// ptrString returns a pointer to a string.
func ptrString(s string) *string {
	return &s
}
