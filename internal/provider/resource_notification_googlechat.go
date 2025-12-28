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
	_ resource.Resource                = &NotificationGoogleChatResource{}
	_ resource.ResourceWithImportState = &NotificationGoogleChatResource{}
)

// NewNotificationGoogleChatResource returns a new instance of the Google Chat notification resource.
func NewNotificationGoogleChatResource() resource.Resource {
	return &NotificationGoogleChatResource{}
}

// NotificationGoogleChatResource defines the resource implementation for Google Chat notifications.
type NotificationGoogleChatResource struct {
	client *kuma.Client
}

// NotificationGoogleChatResourceModel describes the resource data model for Google Chat notifications.
type NotificationGoogleChatResourceModel struct {
	NotificationBaseModel

	WebhookURL  types.String `tfsdk:"webhook_url"`
	UseTemplate types.Bool   `tfsdk:"use_template"`
	Template    types.String `tfsdk:"template"`
}

// Metadata returns the metadata for the resource.
func (*NotificationGoogleChatResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_googlechat"
}

// Schema returns the schema for the resource.
func (*NotificationGoogleChatResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Google Chat notification resource",
		Attributes: withNotificationBaseAttributes(map[string]schema.Attribute{
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: "Google Chat webhook URL for sending notifications",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"use_template": schema.BoolAttribute{
				MarkdownDescription: "Enable custom message template",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Custom message template for notifications",
				Optional:            true,
			},
		}),
	}
}

// Configure configures the Google Chat notification resource with the API client.
func (r *NotificationGoogleChatResource) Configure(
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

// Create creates a new Google Chat notification resource.
func (r *NotificationGoogleChatResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data NotificationGoogleChatResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	googleChat := notification.GoogleChat{
		Base: notification.Base{
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GoogleChatDetails: notification.GoogleChatDetails{
			WebhookURL:  data.WebhookURL.ValueString(),
			UseTemplate: data.UseTemplate.ValueBool(),
			Template:    data.Template.ValueString(),
		},
	}

	id, err := r.client.CreateNotification(ctx, googleChat)
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

// Read reads the current state of the Google Chat notification resource.
func (r *NotificationGoogleChatResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data NotificationGoogleChatResourceModel

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

	googleChat := notification.GoogleChat{}
	err = base.As(&googleChat)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError(`failed to convert notification to type "googlechat"`, err.Error())
		return
	}

	data.ID = types.Int64Value(id)
	data.Name = types.StringValue(googleChat.Name)
	data.IsActive = types.BoolValue(googleChat.IsActive)
	data.IsDefault = types.BoolValue(googleChat.IsDefault)
	data.ApplyExisting = types.BoolValue(googleChat.ApplyExisting)

	data.WebhookURL = types.StringValue(googleChat.WebhookURL)
	data.UseTemplate = types.BoolValue(googleChat.UseTemplate)

	if googleChat.Template != "" {
		data.Template = types.StringValue(googleChat.Template)
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the Google Chat notification resource.
func (r *NotificationGoogleChatResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data NotificationGoogleChatResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	googleChat := notification.GoogleChat{
		Base: notification.Base{
			ID:            data.ID.ValueInt64(),
			ApplyExisting: data.ApplyExisting.ValueBool(),
			IsDefault:     data.IsDefault.ValueBool(),
			IsActive:      data.IsActive.ValueBool(),
			Name:          data.Name.ValueString(),
		},
		GoogleChatDetails: notification.GoogleChatDetails{
			WebhookURL:  data.WebhookURL.ValueString(),
			UseTemplate: data.UseTemplate.ValueBool(),
			Template:    data.Template.ValueString(),
		},
	}

	err := r.client.UpdateNotification(ctx, googleChat)
	// Handle error.
	if err != nil {
		resp.Diagnostics.AddError("failed to update notification", err.Error())
		return
	}

	// Populate state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the Google Chat notification resource.
func (r *NotificationGoogleChatResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data NotificationGoogleChatResourceModel

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
func (*NotificationGoogleChatResource) ImportState(
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
