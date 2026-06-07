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
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorOracleDBResource{}
	_ resource.ResourceWithImportState = &MonitorOracleDBResource{}
)

// NewMonitorOracleDBResource returns a new instance of the OracleDB monitor resource.
func NewMonitorOracleDBResource() resource.Resource {
	return &MonitorOracleDBResource{}
}

// MonitorOracleDBResource defines the resource implementation.
type MonitorOracleDBResource struct {
	client *kuma.Client
}

// MonitorOracleDBResourceModel describes the resource data model.
type MonitorOracleDBResourceModel struct {
	MonitorBaseModel

	DatabaseConnectionString types.String `tfsdk:"database_connection_string"`
	DatabaseQuery            types.String `tfsdk:"database_query"`
	BasicAuthUser            types.String `tfsdk:"basic_auth_user"`
	BasicAuthPass            types.String `tfsdk:"basic_auth_pass"`
	Conditions               types.List   `tfsdk:"conditions"`
}

// Metadata returns the metadata for the resource.
func (*MonitorOracleDBResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_oracledb"
}

// Schema returns the schema for the resource.
func (*MonitorOracleDBResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OracleDB monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"database_connection_string": schema.StringAttribute{
				MarkdownDescription: "Oracle EZCONNECT connection string (e.g., host:port/service)",
				Required:            true,
				Sensitive:           true,
			},
			"database_query": schema.StringAttribute{
				MarkdownDescription: "SQL query to execute for health check",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("SELECT 1 FROM DUAL"),
			},
			"basic_auth_user": schema.StringAttribute{
				MarkdownDescription: "Oracle Database username",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"basic_auth_pass": schema.StringAttribute{
				MarkdownDescription: "Oracle Database password",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				Default:             stringdefault.StaticString(""),
			},
			"conditions": conditionsAttribute(),
		}),
	}
}

// Configure configures the OracleDB monitor resource with the API client.
func (r *MonitorOracleDBResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new OracleDB monitor resource.
func (r *MonitorOracleDBResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorOracleDBResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	databaseQuery := data.DatabaseQuery.ValueString()
	oracleDBMonitor := monitor.OracleDB{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		OracleDBDetails: monitor.OracleDBDetails{
			DatabaseConnectionString: data.DatabaseConnectionString.ValueString(),
			DatabaseQuery:            &databaseQuery,
			Username:                 data.BasicAuthUser.ValueString(),
			Password:                 data.BasicAuthPass.ValueString(),
			Conditions:               buildConditions(ctx, data.Conditions, &resp.Diagnostics),
		},
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		oracleDBMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		oracleDBMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		oracleDBMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, &oracleDBMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create OracleDB monitor", err.Error())
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

// Read reads the current state of the OracleDB monitor resource.
func (r *MonitorOracleDBResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorOracleDBResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var oracleDBMonitor monitor.OracleDB
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &oracleDBMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read OracleDB monitor", err.Error())
		return
	}

	if actual := oracleDBMonitor.Base.Type(); actual != "" && actual != oracleDBMonitor.Type() {
		tflog.Warn(ctx, "monitor type changed externally, removing from state", map[string]any{
			"id":            data.ID.ValueInt64(),
			"expected_type": oracleDBMonitor.Type(),
			"actual_type":   actual,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(oracleDBMonitor.Name)
	if oracleDBMonitor.Description != nil {
		data.Description = types.StringValue(*oracleDBMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(oracleDBMonitor.Interval)
	data.RetryInterval = types.Int64Value(oracleDBMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(oracleDBMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(oracleDBMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(oracleDBMonitor.UpsideDown)
	data.Active = types.BoolValue(oracleDBMonitor.IsActive)
	data.DatabaseConnectionString = types.StringValue(oracleDBMonitor.DatabaseConnectionString)
	if oracleDBMonitor.DatabaseQuery != nil {
		data.DatabaseQuery = types.StringValue(*oracleDBMonitor.DatabaseQuery)
	} else {
		// Normalize a missing database query to the schema default ("SELECT 1 FROM DUAL")
		data.DatabaseQuery = types.StringValue("SELECT 1 FROM DUAL")
	}

	data.BasicAuthUser = types.StringValue(oracleDBMonitor.Username)
	data.BasicAuthPass = types.StringValue(oracleDBMonitor.Password)

	if oracleDBMonitor.Parent != nil {
		data.Parent = types.Int64Value(*oracleDBMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(oracleDBMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, oracleDBMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Conditions = populateConditions(ctx, oracleDBMonitor.Conditions, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Tags = handleMonitorTagsRead(ctx, oracleDBMonitor.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the OracleDB monitor resource.
func (r *MonitorOracleDBResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorOracleDBResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorOracleDBResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseQuery := data.DatabaseQuery.ValueString()
	oracleDBMonitor := monitor.OracleDB{
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
		OracleDBDetails: monitor.OracleDBDetails{
			DatabaseConnectionString: data.DatabaseConnectionString.ValueString(),
			DatabaseQuery:            &databaseQuery,
			Username:                 data.BasicAuthUser.ValueString(),
			Password:                 data.BasicAuthPass.ValueString(),
			Conditions:               buildConditions(ctx, data.Conditions, &resp.Diagnostics),
		},
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		oracleDBMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		oracleDBMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		oracleDBMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, &oracleDBMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update OracleDB monitor", err.Error())
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

// Delete deletes the OracleDB monitor resource.
func (r *MonitorOracleDBResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorOracleDBResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete OracleDB monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorOracleDBResource) ImportState(
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
