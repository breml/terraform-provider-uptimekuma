package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationOpsgenieResource{}
	_ resource.ResourceWithImportState = &NotificationOpsgenieResource{}
)

// NewNotificationOpsgenieResource returns a new instance of the OpsGenie notification resource.
func NewNotificationOpsgenieResource() resource.Resource {
	return &NotificationOpsgenieResource{}
}

// NotificationOpsgenieResource defines the resource implementation.
type NotificationOpsgenieResource struct {
	client *kuma.Client
}

// NotificationOpsgenieResourceModel describes the resource data model.
type NotificationOpsgenieResourceModel struct {
	NotificationBaseModel

	APIKey   types.String `tfsdk:"api_key"`
	Region   types.String `tfsdk:"region"`
	Priority types.Int64  `tfsdk:"priority"`
}

// Metadata returns the metadata for the resource.
func (*NotificationOpsgenieResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_opsgenie"
}

// Schema returns the schema for the resource.
func (*NotificationOpsgenieResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OpsGenie notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "OpsGenie API key",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "OpsGenie region (e.g., 'us', 'eu')",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.OneOf("us", "eu"),
				},
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Alert priority level (1-5)",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3),
				Validators: []validator.Int64{
					int64validator.Between(1, 5),
				},
			},
		}),
	}
}

// Configure configures the OpsGenie notification resource with the API client.
func (r *NotificationOpsgenieResource) Configure(
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

// Create creates a new OpsGenie notification resource.
func (r *NotificationOpsgenieResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationOpsgenieResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	opsgenie := notification.Opsgenie{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OpsgenieDetails: notification.OpsgenieDetails{
			APIKey:   data.APIKey.ValueString(),
			Region:   data.Region.ValueString(),
			Priority: int(data.Priority.ValueInt64()),
		},
	}

	id, err := r.client.CreateNotification(ctx, opsgenie)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the OpsGenie notification resource.
func (r *NotificationOpsgenieResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationOpsgenieResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, err := r.client.GetNotification(ctx, id)
	// Handle error.
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	opsgenie := notification.Opsgenie{}
	err = base.As(&opsgenie)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "opsgenie"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(opsgenie.Name)
	data.IsActive = types.BoolValue(opsgenie.IsActive)
	data.IsDefault = types.BoolValue(opsgenie.IsDefault)
	data.ApplyExisting = types.BoolValue(opsgenie.ApplyExisting)

	data.APIKey = types.StringValue(opsgenie.APIKey)
	data.Region = types.StringValue(opsgenie.Region)
	data.Priority = types.Int64Value(int64(opsgenie.Priority))

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the OpsGenie notification resource.
func (r *NotificationOpsgenieResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationOpsgenieResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	opsgenie := notification.Opsgenie{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OpsgenieDetails: notification.OpsgenieDetails{
			APIKey:   data.APIKey.ValueString(),
			Region:   data.Region.ValueString(),
			Priority: int(data.Priority.ValueInt64()),
		},
	}

	err := r.client.UpdateNotification(ctx, opsgenie)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the OpsGenie notification resource.
func (r *NotificationOpsgenieResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationOpsgenieResourceModel

	// Get resource from state.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*NotificationOpsgenieResource) ImportState(
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
