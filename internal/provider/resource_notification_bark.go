package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationBarkResource{}
	_ resource.ResourceWithImportState = &NotificationBarkResource{}
)

// NewNotificationBarkResource returns a new instance of the Bark notification resource.
func NewNotificationBarkResource() resource.Resource {
	return &NotificationBarkResource{}
}

// NotificationBarkResource defines the resource implementation.
type NotificationBarkResource struct {
	client *kuma.Client
}

// NotificationBarkResourceModel describes the resource data model.
type NotificationBarkResourceModel struct {
	NotificationBaseModel

	Endpoint   types.String `tfsdk:"endpoint"`
	Group      types.String `tfsdk:"group"`
	Sound      types.String `tfsdk:"sound"`
	APIVersion types.String `tfsdk:"api_version"`
}

// Metadata returns the metadata for the resource.
func (*NotificationBarkResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_bark"
}

// Schema returns the schema for the resource.
func (*NotificationBarkResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Bark notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Bark server endpoint URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "Notification group name",
				Optional:            true,
			},
			"sound": schema.StringAttribute{
				MarkdownDescription: "Notification sound",
				Optional:            true,
			},
			"api_version": schema.StringAttribute{
				MarkdownDescription: "API version (v1 or v2)",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Bark notification resource with the API client.
func (r *NotificationBarkResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Bark notification resource.
func (r *NotificationBarkResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationBarkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bark := notification.Bark{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		BarkDetails: notification.BarkDetails{
			Endpoint:   data.Endpoint.ValueString(),
			Group:      data.Group.ValueString(),
			Sound:      data.Sound.ValueString(),
			APIVersion: data.APIVersion.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, bark)
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

// Read reads the current state of the Bark notification resource.
func (r *NotificationBarkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationBarkResourceModel

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

	bark := notification.Bark{}
	err = base.As(&bark)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "bark"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(bark.Name)
	data.IsActive = types.BoolValue(bark.IsActive)
	data.IsDefault = types.BoolValue(bark.IsDefault)
	data.ApplyExisting = types.BoolValue(bark.ApplyExisting)

	data.Endpoint = types.StringValue(bark.Endpoint)
	if bark.Group != "" {
		data.Group = types.StringValue(bark.Group)
	} else {
		data.Group = types.StringNull()
	}

	if bark.Sound != "" {
		data.Sound = types.StringValue(bark.Sound)
	} else {
		data.Sound = types.StringNull()
	}

	if bark.APIVersion != "" {
		data.APIVersion = types.StringValue(bark.APIVersion)
	} else {
		data.APIVersion = types.StringNull()
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Bark notification resource.
func (r *NotificationBarkResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationBarkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bark := notification.Bark{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		BarkDetails: notification.BarkDetails{
			Endpoint:   data.Endpoint.ValueString(),
			Group:      data.Group.ValueString(),
			Sound:      data.Sound.ValueString(),
			APIVersion: data.APIVersion.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, bark)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Bark notification resource.
func (r *NotificationBarkResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationBarkResourceModel

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
func (*NotificationBarkResource) ImportState(
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
