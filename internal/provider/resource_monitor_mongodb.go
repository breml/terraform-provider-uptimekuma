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
	_ resource.Resource                = &MonitorMongoDBResource{}
	_ resource.ResourceWithImportState = &MonitorMongoDBResource{}
)

// NewMonitorMongoDBResource returns a new instance of the MongoDB monitor resource.
func NewMonitorMongoDBResource() resource.Resource {
	return &MonitorMongoDBResource{}
}

// MonitorMongoDBResource defines the resource implementation.
type MonitorMongoDBResource struct {
	client *kuma.Client
}

// MonitorMongoDBResourceModel describes the resource data model.
type MonitorMongoDBResourceModel struct {
	MonitorBaseModel

	DatabaseConnectionString types.String `tfsdk:"database_connection_string"`
	DatabaseQuery            types.String `tfsdk:"database_query"`
	JSONPath                 types.String `tfsdk:"json_path"`
	ExpectedValue            types.String `tfsdk:"expected_value"`
}

// Metadata returns the metadata for the resource.
func (*MonitorMongoDBResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_mongodb"
}

// Schema returns the schema for the resource.
func (*MonitorMongoDBResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MongoDB monitor resource",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			"database_connection_string": schema.StringAttribute{
				MarkdownDescription: "MongoDB connection string (e.g., mongodb://username:password@host:port/database)",
				Required:            true,
				Sensitive:           true,
			},
			"database_query": schema.StringAttribute{
				MarkdownDescription: "MongoDB command as JSON (e.g., {\"ping\": 1})",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(`{"ping": 1}`),
			},
			"json_path": schema.StringAttribute{
				MarkdownDescription: "JSONata expression for result validation",
				Optional:            true,
			},
			"expected_value": schema.StringAttribute{
				MarkdownDescription: "Expected value when using json_path",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the MongoDB monitor resource with the API client.
func (r *MonitorMongoDBResource) Configure(
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

// Create creates a new MongoDB monitor resource.
func (r *MonitorMongoDBResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorMongoDBResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mongoDBMonitor := monitor.MongoDB{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		MongoDBDetails: monitor.MongoDBDetails{
			DatabaseConnectionString: data.DatabaseConnectionString.ValueString(),
			DatabaseQuery:            strToPtr(data.DatabaseQuery),
			JSONPath:                 strToPtr(data.JSONPath),
			ExpectedValue:            strToPtr(data.ExpectedValue),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		mongoDBMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		mongoDBMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		mongoDBMonitor.NotificationIDs = notificationIDs
	}

	id, err := r.client.CreateMonitor(ctx, &mongoDBMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create MongoDB monitor", err.Error())
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

// Read reads the current state of the MongoDB monitor resource.
func (r *MonitorMongoDBResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data MonitorMongoDBResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var mongoDBMonitor monitor.MongoDB
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &mongoDBMonitor)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read MongoDB monitor", err.Error())
		return
	}

	data.Name = types.StringValue(mongoDBMonitor.Name)
	if mongoDBMonitor.Description != nil {
		data.Description = types.StringValue(*mongoDBMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(mongoDBMonitor.Interval)
	data.RetryInterval = types.Int64Value(mongoDBMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(mongoDBMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(mongoDBMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(mongoDBMonitor.UpsideDown)
	data.Active = types.BoolValue(mongoDBMonitor.IsActive)
	data.DatabaseConnectionString = types.StringValue(mongoDBMonitor.DatabaseConnectionString)
	if mongoDBMonitor.DatabaseQuery == nil {
		data.DatabaseQuery = types.StringValue(`{"ping": 1}`)
	} else {
		data.DatabaseQuery = ptrToTypes(mongoDBMonitor.DatabaseQuery)
	}

	data.JSONPath = ptrToTypes(mongoDBMonitor.JSONPath)
	data.ExpectedValue = ptrToTypes(mongoDBMonitor.ExpectedValue)

	if mongoDBMonitor.Parent != nil {
		data.Parent = types.Int64Value(*mongoDBMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(mongoDBMonitor.NotificationIDs) > 0 {
		notificationIDs, diags := types.ListValueFrom(ctx, types.Int64Type, mongoDBMonitor.NotificationIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, mongoDBMonitor.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the MongoDB monitor resource.
func (r *MonitorMongoDBResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorMongoDBResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorMongoDBResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mongoDBMonitor := monitor.MongoDB{
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
		MongoDBDetails: monitor.MongoDBDetails{
			DatabaseConnectionString: data.DatabaseConnectionString.ValueString(),
			DatabaseQuery:            strToPtr(data.DatabaseQuery),
			JSONPath:                 strToPtr(data.JSONPath),
			ExpectedValue:            strToPtr(data.ExpectedValue),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		mongoDBMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		mongoDBMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		mongoDBMonitor.NotificationIDs = notificationIDs
	}

	err := r.client.UpdateMonitor(ctx, &mongoDBMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update MongoDB monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the MongoDB monitor resource.
func (r *MonitorMongoDBResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorMongoDBResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete MongoDB monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorMongoDBResource) ImportState(
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
