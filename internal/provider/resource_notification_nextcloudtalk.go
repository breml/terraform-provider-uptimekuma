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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/notification"
)

var (
	_ resource.Resource                = &NotificationNextcloudTalkResource{}
	_ resource.ResourceWithImportState = &NotificationNextcloudTalkResource{}
)

// NewNotificationNextcloudTalkResource returns a new instance of the Nextcloud Talk notification resource.
func NewNotificationNextcloudTalkResource() resource.Resource {
	return &NotificationNextcloudTalkResource{}
}

// NotificationNextcloudTalkResource defines the resource implementation.
type NotificationNextcloudTalkResource struct {
	client *kuma.Client
}

// NotificationNextcloudTalkResourceModel describes the resource data model.
type NotificationNextcloudTalkResourceModel struct {
	NotificationBaseModel

	Host              types.String `tfsdk:"host"`
	ConversationToken types.String `tfsdk:"conversation_token"`
	BotSecret         types.String `tfsdk:"bot_secret"`
	SendSilentUp      types.Bool   `tfsdk:"send_silent_up"`
	SendSilentDown    types.Bool   `tfsdk:"send_silent_down"`
}

// Metadata returns the metadata for the resource.
func (*NotificationNextcloudTalkResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_nextcloudtalk"
}

// Schema returns the schema for the resource.
func (*NotificationNextcloudTalkResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Nextcloud Talk notification resource.",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Nextcloud instance host URL.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"conversation_token": schema.StringAttribute{
				MarkdownDescription: "Conversation/room token for the target chat.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"bot_secret": schema.StringAttribute{
				MarkdownDescription: "Bot secret for authentication and message signing.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"send_silent_up": schema.BoolAttribute{
				MarkdownDescription: "Send UP notifications silently.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"send_silent_down": schema.BoolAttribute{
				MarkdownDescription: "Send DOWN notifications silently.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		}),
	}
}

// Configure configures the Nextcloud Talk notification resource with the API client.
func (r *NotificationNextcloudTalkResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Nextcloud Talk notification resource.
func (r *NotificationNextcloudTalkResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationNextcloudTalkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nextcloudTalk := notification.NextcloudTalk{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		NextcloudTalkDetails: notification.NextcloudTalkDetails{
			Host:              data.Host.ValueString(),
			ConversationToken: data.ConversationToken.ValueString(),
			BotSecret:         data.BotSecret.ValueString(),
			SendSilentUp:      data.SendSilentUp.ValueBool(),
			SendSilentDown:    data.SendSilentDown.ValueBool(),
		},
	}

	id, err := r.client.CreateNotification(ctx, nextcloudTalk)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Nextcloud Talk notification resource.
func (r *NotificationNextcloudTalkResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationNextcloudTalkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()

	base, err := r.client.GetNotification(ctx, id)
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	nextcloudTalk := notification.NextcloudTalk{}
	err = base.As(&nextcloudTalk)
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "NextcloudTalk"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(nextcloudTalk.Name)
	data.IsActive = types.BoolValue(nextcloudTalk.IsActive)
	data.IsDefault = types.BoolValue(nextcloudTalk.IsDefault)
	data.ApplyExisting = types.BoolValue(nextcloudTalk.ApplyExisting)

	data.Host = types.StringValue(nextcloudTalk.Host)
	data.ConversationToken = types.StringValue(nextcloudTalk.ConversationToken)
	data.BotSecret = types.StringValue(nextcloudTalk.BotSecret)
	data.SendSilentUp = types.BoolValue(nextcloudTalk.SendSilentUp)
	data.SendSilentDown = types.BoolValue(nextcloudTalk.SendSilentDown)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Nextcloud Talk notification resource.
func (r *NotificationNextcloudTalkResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationNextcloudTalkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nextcloudTalk := notification.NextcloudTalk{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		NextcloudTalkDetails: notification.NextcloudTalkDetails{
			Host:              data.Host.ValueString(),
			ConversationToken: data.ConversationToken.ValueString(),
			BotSecret:         data.BotSecret.ValueString(),
			SendSilentUp:      data.SendSilentUp.ValueBool(),
			SendSilentDown:    data.SendSilentDown.ValueBool(),
		},
	}

	err := r.client.UpdateNotification(ctx, nextcloudTalk)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Nextcloud Talk notification resource.
func (r *NotificationNextcloudTalkResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationNextcloudTalkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

// ImportState imports an existing resource by ID.
func (*NotificationNextcloudTalkResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be a valid integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
