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
	_ resource.Resource                = &NotificationSlackResource{}
	_ resource.ResourceWithImportState = &NotificationSlackResource{}
)

func NewNotificationSlackResource() resource.Resource {
	return &NotificationSlackResource{}
}

type NotificationSlackResource struct {
	client *kuma.Client
}

type NotificationSlackResourceModel struct {
	NotificationBaseModel

	WebhookURL    types.String `tfsdk:"webhook_url"`
	Username      types.String `tfsdk:"username"`
	IconEmoji     types.String `tfsdk:"icon_emoji"`
	Channel       types.String `tfsdk:"channel"`
	RichMessage   types.Bool   `tfsdk:"rich_message"`
	ChannelNotify types.Bool   `tfsdk:"channel_notify"`
}

func (r *NotificationSlackResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_slack"
}

func (r *NotificationSlackResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Slack notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "Slack webhook URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username to display in Slack",
				Optional:            true,
			},
			"icon_emoji": schema.StringAttribute{
				MarkdownDescription: "Icon emoji to display in Slack",
				Optional:            true,
			},
			"channel": schema.StringAttribute{
				MarkdownDescription: "Channel name to send notifications to",
				Optional:            true,
			},
			"rich_message": schema.BoolAttribute{
				MarkdownDescription: "Enable rich message formatting",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"channel_notify": schema.BoolAttribute{
				MarkdownDescription: "Notify channel with @channel mention",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		}),
	}
}

func (r *NotificationSlackResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kuma.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *kuma.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *NotificationSlackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NotificationSlackResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	slack := notification.Slack{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SlackDetails: notification.SlackDetails{
			WebhookURL:    data.WebhookURL.ValueString(),
			Username:      data.Username.ValueString(),
			IconEmoji:     data.IconEmoji.ValueString(),
			Channel:       data.Channel.ValueString(),
			RichMessage:   data.RichMessage.ValueBool(),
			ChannelNotify: data.ChannelNotify.ValueBool(),
		},
	}

	id, err := r.client.CreateNotification(ctx, slack)
	if err != nil {
		resp.Diagnostics.AddError("failed to create notification", err.Error())
		return
	}

	tflog.Info(ctx, "Got ID", map[string]any{"id": id})

	data.Id = types.Int64Value(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationSlackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationSlackResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.Id.ValueInt64()

	base, err := r.client.GetNotification(ctx, id)
	if err != nil {
		if errors.Is(err, kuma.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("failed to read notification", err.Error())
		return
	}

	slack := notification.Slack{}
	err = base.As(&slack)
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "slack"`, err.Error())
		return
	}

	data.Id = types.Int64Value(id)
	data.Name = types.StringValue(slack.Name)
	data.IsActive = types.BoolValue(slack.IsActive)
	data.IsDefault = types.BoolValue(slack.IsDefault)
	data.ApplyExisting = types.BoolValue(slack.ApplyExisting)

	data.WebhookURL = types.StringValue(slack.WebhookURL)
	if slack.Username != "" {
		data.Username = types.StringValue(slack.Username)
	}
	if slack.IconEmoji != "" {
		data.IconEmoji = types.StringValue(slack.IconEmoji)
	}
	if slack.Channel != "" {
		data.Channel = types.StringValue(slack.Channel)
	}
	data.RichMessage = types.BoolValue(slack.RichMessage)
	data.ChannelNotify = types.BoolValue(slack.ChannelNotify)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationSlackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NotificationSlackResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	slack := notification.Slack{
		Base: notification.Base{
			ID:            data.Id.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		SlackDetails: notification.SlackDetails{
			WebhookURL:    data.WebhookURL.ValueString(),
			Username:      data.Username.ValueString(),
			IconEmoji:     data.IconEmoji.ValueString(),
			Channel:       data.Channel.ValueString(),
			RichMessage:   data.RichMessage.ValueBool(),
			ChannelNotify: data.ChannelNotify.ValueBool(),
		},
	}

	err := r.client.UpdateNotification(ctx, slack)
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationSlackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NotificationSlackResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("failed to delete notification", err.Error())
		return
	}
}

func (r *NotificationSlackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
