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
	_ resource.Resource                = &NotificationDiscordResource{}
	_ resource.ResourceWithImportState = &NotificationDiscordResource{}
)

// NewNotificationDiscordResource returns a new instance of the Discord notification resource.
func NewNotificationDiscordResource() resource.Resource {
	return &NotificationDiscordResource{}
}

// NotificationDiscordResource defines the resource implementation.
type NotificationDiscordResource struct {
	client *kuma.Client
}

// NotificationDiscordResourceModel describes the resource data model.
type NotificationDiscordResourceModel struct {
	NotificationBaseModel

	WebhookURL    types.String `tfsdk:"webhook_url"`
	Username      types.String `tfsdk:"username"`
	ChannelType   types.String `tfsdk:"channel_type"`
	ThreadID      types.String `tfsdk:"thread_id"`
	PostName      types.String `tfsdk:"post_name"`
	PrefixMessage types.String `tfsdk:"prefix_message"`
	DisableURL    types.Bool   `tfsdk:"disable_url"`
}

func (_ *NotificationDiscordResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_discord"
}

func (_ *NotificationDiscordResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Discord notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "The Discord webhook URL to which notifications will be sent.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username to display as the sender of the Discord notification.",
				Optional:            true,
			},
			"channel_type": schema.StringAttribute{
				MarkdownDescription: "The type of Discord channel associated with the webhook (for example, text, forum, or announcement).",
				Optional:            true,
			},
			"thread_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Discord thread to send the notification to, if applicable.",
				Optional:            true,
			},
			"post_name": schema.StringAttribute{
				MarkdownDescription: "The display name override for the Discord webhook when posting notifications (appears as the sender name in Discord).",
				Optional:            true,
			},
			"prefix_message": schema.StringAttribute{
				MarkdownDescription: "A message to prefix to all Discord notifications.",
				Optional:            true,
			},
			"disable_url": schema.BoolAttribute{
				MarkdownDescription: "If true, disables including the monitor URL in the Discord notification.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		}),
	}
}

// Configure configures the Discord notification resource with the API client.
func (r *NotificationDiscordResource) Configure(
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

// Create creates a new Discord notification resource.
func (r *NotificationDiscordResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationDiscordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	discord := notification.Discord{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		DiscordDetails: notification.DiscordDetails{
			WebhookURL:    data.WebhookURL.ValueString(),
			Username:      data.Username.ValueString(),
			ChannelType:   data.ChannelType.ValueString(),
			ThreadID:      data.ThreadID.ValueString(),
			PostName:      data.PostName.ValueString(),
			PrefixMessage: data.PrefixMessage.ValueString(),
			DisableURL:    data.DisableURL.ValueBool(),
		},
	}

	id, err := r.client.CreateNotification(ctx, discord)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.ID = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the current state of the Discord notification resource.
func (r *NotificationDiscordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationDiscordResourceModel

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

	discord := notification.Discord{}
	err = base.As(&discord)
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "discord"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(discord.Name)
	data.IsActive = types.BoolValue(discord.IsActive)
	data.IsDefault = types.BoolValue(discord.IsDefault)
	data.ApplyExisting = types.BoolValue(discord.ApplyExisting)

	data.WebhookURL = types.StringValue(discord.WebhookURL)
	if discord.Username != "" {
		data.Username = types.StringValue(discord.Username)
	} else {
		data.Username = types.StringNull()
	}

	if discord.ChannelType != "" {
		data.ChannelType = types.StringValue(discord.ChannelType)
	} else {
		data.ChannelType = types.StringNull()
	}

	if discord.ThreadID != "" {
		data.ThreadID = types.StringValue(discord.ThreadID)
	} else {
		data.ThreadID = types.StringNull()
	}

	if discord.PostName != "" {
		data.PostName = types.StringValue(discord.PostName)
	} else {
		data.PostName = types.StringNull()
	}

	if discord.PrefixMessage != "" {
		data.PrefixMessage = types.StringValue(discord.PrefixMessage)
	} else {
		data.PrefixMessage = types.StringNull()
	}

	data.DisableURL = types.BoolValue(discord.DisableURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Discord notification resource.
func (r *NotificationDiscordResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationDiscordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	discord := notification.Discord{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		DiscordDetails: notification.DiscordDetails{
			WebhookURL:    data.WebhookURL.ValueString(),
			Username:      data.Username.ValueString(),
			ChannelType:   data.ChannelType.ValueString(),
			ThreadID:      data.ThreadID.ValueString(),
			PostName:      data.PostName.ValueString(),
			PrefixMessage: data.PrefixMessage.ValueString(),
			DisableURL:    data.DisableURL.ValueBool(),
		},
	}

	err := r.client.UpdateNotification(ctx, discord)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Discord notification resource.
func (r *NotificationDiscordResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationDiscordResourceModel

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
func (_ *NotificationDiscordResource) ImportState(
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
