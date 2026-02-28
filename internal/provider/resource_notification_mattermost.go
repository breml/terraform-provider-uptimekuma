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
	_ resource.Resource                = &NotificationMattermostResource{}
	_ resource.ResourceWithImportState = &NotificationMattermostResource{}
)

// NewNotificationMattermostResource returns a new instance of the Mattermost notification resource.
func NewNotificationMattermostResource() resource.Resource {
	return &NotificationMattermostResource{}
}

// NotificationMattermostResource defines the resource implementation.
type NotificationMattermostResource struct {
	client *kuma.Client
}

// NotificationMattermostResourceModel describes the resource data model.
type NotificationMattermostResourceModel struct {
	NotificationBaseModel

	WebhookURL types.String `tfsdk:"webhook_url"`
	Username   types.String `tfsdk:"username"`
	Channel    types.String `tfsdk:"channel"`
	IconEmoji  types.String `tfsdk:"icon_emoji"`
	IconURL    types.String `tfsdk:"icon_url"`
}

// Metadata returns the metadata for the resource.
func (*NotificationMattermostResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_mattermost"
}

// Schema returns the schema for the resource.
func (*NotificationMattermostResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Mattermost notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "Mattermost webhook URL",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username to display in Mattermost",
				Optional:            true,
			},
			"channel": schema.StringAttribute{
				MarkdownDescription: "Channel name to send notifications to",
				Optional:            true,
			},
			"icon_emoji": schema.StringAttribute{
				MarkdownDescription: "Icon emoji to display in Mattermost",
				Optional:            true,
			},
			"icon_url": schema.StringAttribute{
				MarkdownDescription: "Icon URL to display in Mattermost",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Mattermost notification resource with the API client.
func (r *NotificationMattermostResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new Mattermost notification resource.
func (r *NotificationMattermostResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationMattermostResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mattermost := notification.Mattermost{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		MattermostDetails: notification.MattermostDetails{
			WebhookURL: data.WebhookURL.ValueString(),
			Username:   data.Username.ValueString(),
			Channel:    data.Channel.ValueString(),
			IconEmoji:  data.IconEmoji.ValueString(),
			IconURL:    data.IconURL.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, mattermost)
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

// Read reads the current state of the Mattermost notification resource.
func (r *NotificationMattermostResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationMattermostResourceModel

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

	mattermost := notification.Mattermost{}
	err = base.As(&mattermost)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "mattermost"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(mattermost.Name)
	data.IsActive = types.BoolValue(mattermost.IsActive)
	data.IsDefault = types.BoolValue(mattermost.IsDefault)
	data.ApplyExisting = types.BoolValue(mattermost.ApplyExisting)

	data.WebhookURL = types.StringValue(mattermost.WebhookURL)
	if mattermost.Username != "" {
		data.Username = types.StringValue(mattermost.Username)
	}

	if mattermost.Channel != "" {
		data.Channel = types.StringValue(mattermost.Channel)
	}

	if mattermost.IconEmoji != "" {
		data.IconEmoji = types.StringValue(mattermost.IconEmoji)
	}

	if mattermost.IconURL != "" {
		data.IconURL = types.StringValue(mattermost.IconURL)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Mattermost notification resource.
func (r *NotificationMattermostResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationMattermostResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mattermost := notification.Mattermost{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		MattermostDetails: notification.MattermostDetails{
			WebhookURL: data.WebhookURL.ValueString(),
			Username:   data.Username.ValueString(),
			Channel:    data.Channel.ValueString(),
			IconEmoji:  data.IconEmoji.ValueString(),
			IconURL:    data.IconURL.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, mattermost)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Mattermost notification resource.
func (r *NotificationMattermostResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationMattermostResourceModel

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
func (*NotificationMattermostResource) ImportState(
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
