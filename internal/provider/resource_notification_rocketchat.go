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
	_ resource.Resource                = &NotificationRocketChatResource{}
	_ resource.ResourceWithImportState = &NotificationRocketChatResource{}
)

// NewNotificationRocketChatResource returns a new instance of the RocketChat notification resource.
func NewNotificationRocketChatResource() resource.Resource {
	return &NotificationRocketChatResource{}
}

// NotificationRocketChatResource defines the resource implementation.
type NotificationRocketChatResource struct {
	client *kuma.Client
}

// NotificationRocketChatResourceModel describes the resource data model.
type NotificationRocketChatResourceModel struct {
	NotificationBaseModel

	WebhookURL types.String `tfsdk:"webhook_url"`
	Username   types.String `tfsdk:"username"`
	IconEmoji  types.String `tfsdk:"icon_emoji"`
	Channel    types.String `tfsdk:"channel"`
	Button     types.String `tfsdk:"button"`
}

// Metadata returns the metadata for the resource.
func (*NotificationRocketChatResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_rocketchat"
}

// Schema returns the schema for the resource.
func (*NotificationRocketChatResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "RocketChat notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "RocketChat webhook URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username to display in RocketChat",
				Optional:            true,
			},
			"icon_emoji": schema.StringAttribute{
				MarkdownDescription: "Icon emoji to display in RocketChat",
				Optional:            true,
			},
			"channel": schema.StringAttribute{
				MarkdownDescription: "Channel name to send notifications to",
				Optional:            true,
			},
			"button": schema.StringAttribute{
				MarkdownDescription: "Button text to include in notifications",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the RocketChat notification resource with the API client.
func (r *NotificationRocketChatResource) Configure(
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

// Create creates a new RocketChat notification resource.
func (r *NotificationRocketChatResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationRocketChatResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rocketChat := notification.RocketChat{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		RocketChatDetails: notification.RocketChatDetails{
			WebhookURL: data.WebhookURL.ValueString(),
			Username:   data.Username.ValueString(),
			IconEmoji:  data.IconEmoji.ValueString(),
			Channel:    data.Channel.ValueString(),
			Button:     data.Button.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, rocketChat)
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

// Read reads the current state of the RocketChat notification resource.
func (r *NotificationRocketChatResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationRocketChatResourceModel

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

	rocketChat := notification.RocketChat{}
	err = base.As(&rocketChat)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "rocketchat"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(rocketChat.Name)
	data.IsActive = types.BoolValue(rocketChat.IsActive)
	data.IsDefault = types.BoolValue(rocketChat.IsDefault)
	data.ApplyExisting = types.BoolValue(rocketChat.ApplyExisting)

	data.WebhookURL = types.StringValue(rocketChat.WebhookURL)
	if rocketChat.Username != "" {
		data.Username = types.StringValue(rocketChat.Username)
	}

	if rocketChat.IconEmoji != "" {
		data.IconEmoji = types.StringValue(rocketChat.IconEmoji)
	}

	if rocketChat.Channel != "" {
		data.Channel = types.StringValue(rocketChat.Channel)
	}

	if rocketChat.Button != "" {
		data.Button = types.StringValue(rocketChat.Button)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the RocketChat notification resource.
func (r *NotificationRocketChatResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationRocketChatResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rocketChat := notification.RocketChat{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		RocketChatDetails: notification.RocketChatDetails{
			WebhookURL: data.WebhookURL.ValueString(),
			Username:   data.Username.ValueString(),
			IconEmoji:  data.IconEmoji.ValueString(),
			Channel:    data.Channel.ValueString(),
			Button:     data.Button.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, rocketChat)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the RocketChat notification resource.
func (r *NotificationRocketChatResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationRocketChatResourceModel

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
func (*NotificationRocketChatResource) ImportState(
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
