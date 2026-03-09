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
	_ resource.Resource                = &NotificationOneBotResource{}
	_ resource.ResourceWithImportState = &NotificationOneBotResource{}
)

// NewNotificationOneBotResource returns a new instance of the OneBot notification resource.
func NewNotificationOneBotResource() resource.Resource {
	return &NotificationOneBotResource{}
}

// NotificationOneBotResource defines the resource implementation.
type NotificationOneBotResource struct {
	client *kuma.Client
}

// NotificationOneBotResourceModel describes the resource data model.
type NotificationOneBotResourceModel struct {
	NotificationBaseModel

	HTTPAddr    types.String `tfsdk:"http_addr"`
	AccessToken types.String `tfsdk:"access_token"`
	MsgType     types.String `tfsdk:"msg_type"`
	ReceiverID  types.String `tfsdk:"receiver_id"`
}

// Metadata returns the metadata for the resource.
func (*NotificationOneBotResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_onebot"
}

// Schema returns the schema for the resource.
func (*NotificationOneBotResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OneBot notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"http_addr": schema.StringAttribute{
				MarkdownDescription: "HTTP address of the OneBot service (e.g., `http://localhost:5700`).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Access token for authentication with the OneBot service.",
				Optional:            true,
				Sensitive:           true,
			},
			"msg_type": schema.StringAttribute{
				MarkdownDescription: "Message type: `group` for group messages or `private` for private messages.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("group", "private"),
				},
			},
			"receiver_id": schema.StringAttribute{
				MarkdownDescription: "QQ group ID (when `msg_type` is `group`) or user ID (when `msg_type` is `private`).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		}),
	}
}

// Configure configures the OneBot notification resource with the API client.
func (r *NotificationOneBotResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new OneBot notification resource.
func (r *NotificationOneBotResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationOneBotResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	onebot := notification.OneBot{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OneBotDetails: notification.OneBotDetails{
			HTTPAddr:    data.HTTPAddr.ValueString(),
			AccessToken: data.AccessToken.ValueString(),
			MsgType:     data.MsgType.ValueString(),
			ReceiverID:  data.ReceiverID.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, onebot)
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

// Read reads the current state of the OneBot notification resource.
func (r *NotificationOneBotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationOneBotResourceModel

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

	onebot := notification.OneBot{}
	err = base.As(&onebot)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "OneBot"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(onebot.Name)
	data.IsActive = types.BoolValue(onebot.IsActive)
	data.IsDefault = types.BoolValue(onebot.IsDefault)
	data.ApplyExisting = types.BoolValue(onebot.ApplyExisting)

	data.HTTPAddr = types.StringValue(onebot.HTTPAddr)
	data.MsgType = types.StringValue(onebot.MsgType)
	data.ReceiverID = types.StringValue(onebot.ReceiverID)

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the OneBot notification resource.
func (r *NotificationOneBotResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationOneBotResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	onebot := notification.OneBot{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		OneBotDetails: notification.OneBotDetails{
			HTTPAddr:    data.HTTPAddr.ValueString(),
			AccessToken: data.AccessToken.ValueString(),
			MsgType:     data.MsgType.ValueString(),
			ReceiverID:  data.ReceiverID.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, onebot)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the OneBot notification resource.
func (r *NotificationOneBotResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationOneBotResourceModel

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
func (*NotificationOneBotResource) ImportState(
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
