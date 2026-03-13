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
	_ resource.Resource                = &NotificationOneChatResource{}
	_ resource.ResourceWithImportState = &NotificationOneChatResource{}
)

// NewNotificationOneChatResource returns a new instance of the OneChat notification resource.
func NewNotificationOneChatResource() resource.Resource {
	return &NotificationOneChatResource{}
}

// NotificationOneChatResource defines the resource implementation.
type NotificationOneChatResource struct {
	client *kuma.Client
}

// NotificationOneChatResourceModel describes the resource data model.
type NotificationOneChatResourceModel struct {
	NotificationBaseModel

	AccessToken types.String `tfsdk:"access_token"`
	ReceiverID  types.String `tfsdk:"receiver_id"`
	BotID       types.String `tfsdk:"bot_id"`
}

// Metadata returns the metadata for the resource.
func (*NotificationOneChatResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_onechat"
}

// Schema returns the schema for the resource.
func (*NotificationOneChatResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OneChat notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Access token for authentication with the OneChat API.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"receiver_id": schema.StringAttribute{
				MarkdownDescription: "Recipient ID (user or group ID) in OneChat.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"bot_id": schema.StringAttribute{
				MarkdownDescription: "Bot ID for sending messages through OneChat.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the OneChat notification resource with the API client.
func (r *NotificationOneChatResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new OneChat notification resource.
func (r *NotificationOneChatResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationOneChatResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	onechat := notification.OneChat{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OneChatDetails: notification.OneChatDetails{
			AccessToken: data.AccessToken.ValueString(),
			ReceiverID:  data.ReceiverID.ValueString(),
			BotID:       data.BotID.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, onechat)
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

// Read reads the current state of the OneChat notification resource.
func (r *NotificationOneChatResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationOneChatResourceModel

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

	onechat := notification.OneChat{}
	err = base.As(&onechat)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "OneChat"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(onechat.Name)
	data.IsActive = types.BoolValue(onechat.IsActive)
	data.IsDefault = types.BoolValue(onechat.IsDefault)
	data.ApplyExisting = types.BoolValue(onechat.ApplyExisting)

	data.AccessToken = types.StringValue(onechat.AccessToken)
	data.ReceiverID = types.StringValue(onechat.ReceiverID)
	data.BotID = types.StringValue(onechat.BotID)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the OneChat notification resource.
func (r *NotificationOneChatResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationOneChatResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	onechat := notification.OneChat{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OneChatDetails: notification.OneChatDetails{
			AccessToken: data.AccessToken.ValueString(),
			ReceiverID:  data.ReceiverID.ValueString(),
			BotID:       data.BotID.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, onechat)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the OneChat notification resource.
func (r *NotificationOneChatResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationOneChatResourceModel

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
func (*NotificationOneChatResource) ImportState(
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
