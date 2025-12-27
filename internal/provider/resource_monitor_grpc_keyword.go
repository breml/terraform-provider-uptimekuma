package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/monitor"
)

var (
	_ resource.Resource                = &MonitorGrpcKeywordResource{}
	_ resource.ResourceWithImportState = &MonitorGrpcKeywordResource{}
)

// NewMonitorGrpcKeywordResource returns a new instance of the gRPC Keyword monitor resource.
func NewMonitorGrpcKeywordResource() resource.Resource {
	return &MonitorGrpcKeywordResource{}
}

// MonitorGrpcKeywordResource defines the resource implementation.
type MonitorGrpcKeywordResource struct {
	client *kuma.Client
}

// MonitorGrpcKeywordResourceModel describes the resource data model.
type MonitorGrpcKeywordResourceModel struct {
	MonitorBaseModel

	GrpcURL         types.String `tfsdk:"grpc_url"`
	GrpcProtobuf    types.String `tfsdk:"grpc_protobuf"`
	GrpcServiceName types.String `tfsdk:"grpc_service_name"`
	GrpcMethod      types.String `tfsdk:"grpc_method"`
	GrpcEnableTLS   types.Bool   `tfsdk:"grpc_enable_tls"`
	GrpcBody        types.String `tfsdk:"grpc_body"`
	Keyword         types.String `tfsdk:"keyword"`
	InvertKeyword   types.Bool   `tfsdk:"invert_keyword"`
}

// Metadata returns the metadata for the resource.
func (*MonitorGrpcKeywordResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitor_grpc_keyword"
}

// Schema returns the schema for the resource.
func (*MonitorGrpcKeywordResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	// Define resource schema attributes and validation.
	resp.Schema = schema.Schema{
		MarkdownDescription: "gRPC Keyword monitor resource checks for the presence (or absence) of a specific keyword in the gRPC response. The monitor makes a gRPC request and searches for the specified keyword in the response. Use `invert_keyword` to reverse the logic: when false (default), finding the keyword means UP; when true, finding the keyword means DOWN.",
		Attributes: withMonitorBaseAttributes(map[string]schema.Attribute{
			// Monitor-specific attributes.
			"grpc_url": schema.StringAttribute{
				MarkdownDescription: "gRPC server URL (e.g., localhost:50051 or example.com:443)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"grpc_protobuf": schema.StringAttribute{
				MarkdownDescription: "Protocol Buffer definition (proto3 syntax)",
				Optional:            true,
				Computed:            true,
			},
			"grpc_service_name": schema.StringAttribute{
				MarkdownDescription: "gRPC service name from the protobuf definition",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"grpc_method": schema.StringAttribute{
				MarkdownDescription: "gRPC method name to call",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"grpc_enable_tls": schema.BoolAttribute{
				MarkdownDescription: "Enable TLS for gRPC connection",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"grpc_body": schema.StringAttribute{
				MarkdownDescription: "Request body in JSON format",
				Optional:            true,
				Computed:            true,
			},
			"keyword": schema.StringAttribute{
				MarkdownDescription: "Keyword to search for in the response body (case-sensitive). The monitor will search for this exact text in the gRPC response.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"invert_keyword": schema.BoolAttribute{
				MarkdownDescription: "Invert keyword match logic. When false (default), finding the keyword means UP and not finding it means DOWN. When true, finding the keyword means DOWN and not finding it means UP.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		}),
	}
}

// Configure configures the resource with the API client.
func (r *MonitorGrpcKeywordResource) Configure(
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
func (r *MonitorGrpcKeywordResource) Create(
	// Extract and validate configuration.
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data MonitorGrpcKeywordResourceModel

	// Extract plan data.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	grpcKeywordMonitor := monitor.GrpcKeyword{
		Base: monitor.Base{
			Name:           data.Name.ValueString(),
			Interval:       data.Interval.ValueInt64(),
			RetryInterval:  data.RetryInterval.ValueInt64(),
			ResendInterval: data.ResendInterval.ValueInt64(),
			MaxRetries:     data.MaxRetries.ValueInt64(),
			UpsideDown:     data.UpsideDown.ValueBool(),
			IsActive:       data.Active.ValueBool(),
		},
		GrpcKeywordDetails: monitor.GrpcKeywordDetails{
			GrpcURL:         data.GrpcURL.ValueString(),
			GrpcProtobuf:    data.GrpcProtobuf.ValueString(),
			GrpcServiceName: data.GrpcServiceName.ValueString(),
			GrpcMethod:      data.GrpcMethod.ValueString(),
			GrpcEnableTLS:   data.GrpcEnableTLS.ValueBool(),
			GrpcBody:        data.GrpcBody.ValueString(),
			Keyword:         data.Keyword.ValueString(),
			InvertKeyword:   data.InvertKeyword.ValueBool(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		grpcKeywordMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		grpcKeywordMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		grpcKeywordMonitor.NotificationIDs = notificationIDs
	}

	// Create monitor via API.
	id, err := r.client.CreateMonitor(ctx, grpcKeywordMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create gRPC Keyword monitor", err.Error())
		return
	}

	data.ID = types.Int64Value(id)

	handleMonitorTagsCreate(ctx, r.client, id, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var createdMonitor monitor.GrpcKeyword
	// Fetch monitor from API.
	err = r.client.GetMonitorAs(ctx, id, &createdMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to read created gRPC Keyword monitor", err.Error())
		return
	}

	r.populateModelFromMonitor(ctx, &data, &createdMonitor, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the resource.
func (r *MonitorGrpcKeywordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorGrpcKeywordResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var grpcKeywordMonitor monitor.GrpcKeyword
	// Fetch monitor from API.
	err := r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &grpcKeywordMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to read gRPC Keyword monitor", err.Error())
		return
	}

	r.populateModelFromMonitor(ctx, &data, &grpcKeywordMonitor, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource.
func (r *MonitorGrpcKeywordResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data MonitorGrpcKeywordResourceModel

	// Extract plan data.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorGrpcKeywordResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grpcKeywordMonitor := monitor.GrpcKeyword{
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
		GrpcKeywordDetails: monitor.GrpcKeywordDetails{
			GrpcURL:         data.GrpcURL.ValueString(),
			GrpcProtobuf:    data.GrpcProtobuf.ValueString(),
			GrpcServiceName: data.GrpcServiceName.ValueString(),
			GrpcMethod:      data.GrpcMethod.ValueString(),
			GrpcEnableTLS:   data.GrpcEnableTLS.ValueBool(),
			GrpcBody:        data.GrpcBody.ValueString(),
			Keyword:         data.Keyword.ValueString(),
			InvertKeyword:   data.InvertKeyword.ValueBool(),
		},
	}

	if !data.Description.IsNull() {
		desc := data.Description.ValueString()
		grpcKeywordMonitor.Description = &desc
	}

	if !data.Parent.IsNull() {
		parent := data.Parent.ValueInt64()
		grpcKeywordMonitor.Parent = &parent
	}

	if !data.NotificationIDs.IsNull() {
		var notificationIDs []int64
		resp.Diagnostics.Append(data.NotificationIDs.ElementsAs(ctx, &notificationIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		grpcKeywordMonitor.NotificationIDs = notificationIDs
	}

	// Update monitor via API.
	err := r.client.UpdateMonitor(ctx, grpcKeywordMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update gRPC Keyword monitor", err.Error())
		return
	}

	handleMonitorTagsUpdate(ctx, r.client, data.ID.ValueInt64(), state.Tags, data.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var updatedMonitor monitor.GrpcKeyword
	// Fetch monitor from API.
	err = r.client.GetMonitorAs(ctx, data.ID.ValueInt64(), &updatedMonitor)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to read updated gRPC Keyword monitor", err.Error())
		return
	}

	r.populateModelFromMonitor(ctx, &data, &updatedMonitor, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource.
func (r *MonitorGrpcKeywordResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data MonitorGrpcKeywordResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete monitor via API.
	err := r.client.DeleteMonitor(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete gRPC Keyword monitor", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*MonitorGrpcKeywordResource) ImportState(
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

func (*MonitorGrpcKeywordResource) populateModelFromMonitor(
	ctx context.Context,
	data *MonitorGrpcKeywordResourceModel,
	grpcKeywordMonitor *monitor.GrpcKeyword,
	diags *diag.Diagnostics,
) {
	data.Name = types.StringValue(grpcKeywordMonitor.Name)
	if grpcKeywordMonitor.Description != nil {
		data.Description = types.StringValue(*grpcKeywordMonitor.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.Interval = types.Int64Value(grpcKeywordMonitor.Interval)
	data.RetryInterval = types.Int64Value(grpcKeywordMonitor.RetryInterval)
	data.ResendInterval = types.Int64Value(grpcKeywordMonitor.ResendInterval)
	data.MaxRetries = types.Int64Value(grpcKeywordMonitor.MaxRetries)
	data.UpsideDown = types.BoolValue(grpcKeywordMonitor.UpsideDown)
	data.Active = types.BoolValue(grpcKeywordMonitor.IsActive)
	data.GrpcURL = types.StringValue(grpcKeywordMonitor.GrpcURL)
	data.GrpcProtobuf = stringOrNull(grpcKeywordMonitor.GrpcProtobuf)
	data.GrpcServiceName = types.StringValue(grpcKeywordMonitor.GrpcServiceName)
	data.GrpcMethod = types.StringValue(grpcKeywordMonitor.GrpcMethod)
	data.GrpcEnableTLS = types.BoolValue(grpcKeywordMonitor.GrpcEnableTLS)
	data.GrpcBody = stringOrNull(grpcKeywordMonitor.GrpcBody)
	data.Keyword = types.StringValue(grpcKeywordMonitor.Keyword)
	data.InvertKeyword = types.BoolValue(grpcKeywordMonitor.InvertKeyword)

	if grpcKeywordMonitor.Parent != nil {
		data.Parent = types.Int64Value(*grpcKeywordMonitor.Parent)
	} else {
		data.Parent = types.Int64Null()
	}

	if len(grpcKeywordMonitor.NotificationIDs) > 0 {
		notificationIDs, diagsLocal := types.ListValueFrom(ctx, types.Int64Type, grpcKeywordMonitor.NotificationIDs)
		diags.Append(diagsLocal...)
		// Check for configuration errors.
		if diags.HasError() {
			return
		}

		data.NotificationIDs = notificationIDs
	} else {
		data.NotificationIDs = types.ListNull(types.Int64Type)
	}

	data.Tags = handleMonitorTagsRead(ctx, grpcKeywordMonitor.Tags, diags)
	// Check for configuration errors.
	if diags.HasError() {
		return
	}
}
