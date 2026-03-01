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

// NewNotificationSlackResource returns a new instance of the Slack notification resource.
func NewNotificationSlackResource() resource.Resource {
	return &NotificationSlackResource{}
}

// NotificationSlackResource defines the resource implementation.
type NotificationSlackResource struct {
	client *kuma.Client
}

// NotificationSlackResourceModel describes the resource data model.
type NotificationSlackResourceModel struct {
	NotificationBaseModel

	WebhookURL    types.String `tfsdk:"webhook_url"`
	Username      types.String `tfsdk:"username"`
	IconEmoji     types.String `tfsdk:"icon_emoji"`
	Channel       types.String `tfsdk:"channel"`
	RichMessage   types.Bool   `tfsdk:"rich_message"`
	ChannelNotify types.Bool   `tfsdk:"channel_notify"`
}

// Metadata returns the metadata for the resource.
func (*NotificationSlackResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_slack"
}

// Schema returns the schema for the resource.
func (*NotificationSlackResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
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

// Configure configures the Slack notification resource with the API client.
func (r *NotificationSlackResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Slack notification resource.
func (r *NotificationSlackResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
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

// Read reads the current state of the Slack notification resource.
func (r *NotificationSlackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationSlackResourceModel

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

	slack := notification.Slack{}
	err = base.As(&slack)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "slack"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
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

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Slack notification resource.
func (r *NotificationSlackResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationSlackResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	slack := notification.Slack{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
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
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Slack notification resource.
func (r *NotificationSlackResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationSlackResourceModel

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
func (*NotificationSlackResource) ImportState(
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
